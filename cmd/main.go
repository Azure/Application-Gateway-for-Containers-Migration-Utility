// Package main is the entry point for the AGIC migration CLI tool.
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	appgwrewrite "github.com/Azure/application-gateway-kubernetes-ingress/pkg/apis/azureapplicationgatewayrewrite/v1beta1"
	cli "github.com/urfave/cli/v3"
	core_v1 "k8s.io/api/core/v1"
	networking_v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl_client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/aggregation"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/migration"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/output"
)

type options struct {
	Migration migration.Options
	Output    output.Options
}

func main() {
	deploymentCategory := "DEPLOYMENT OPTIONS"
	outputCategory := "OUTPUT OPTIONS"

	// Define shared flags that apply to both cluster and files commands
	sharedFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:        "dry-run",
			Usage:       "Execute the migration flow but only produce the migration report.",
			HideDefault: true,
			Category:    outputCategory,
		},
		&cli.StringFlag{
			Name:     "output-dir",
			Aliases:  []string{"o"},
			Usage:    "Directory to write the migrated YAML to. If not specified, output will be written to stdout (current terminal).",
			Category: outputCategory,
		},
		&cli.StringFlag{
			Name:     "output-file",
			Usage:    "File to write migrated YAML to. If not specified, output will be written to stdout (current terminal).",
			Category: outputCategory,
		},
		&cli.StringFlag{
			Name:    "waf-id",
			Aliases: []string{"w"},
			Usage: "Web Application Firewall policy ID to use in Application Gateway for Containers. " +
				"If it is not supplied and the tool is run in cluster mode, then the tool will attempt to scrape the " +
				"ingress controller pod logs for a WAF ID",
			Category: deploymentCategory,
		},
		&cli.StringFlag{
			Name:        "ingress-class",
			Aliases:     []string{"c"},
			Usage:       "Name of the IngressClass to filter by. Only Ingresses with this class will be migrated.",
			DefaultText: aggregation.DefaultIngressClassName,
		},
		&cli.StringFlag{
			Name:     "byo-resource-id",
			Aliases:  []string{"b"},
			Usage:    "The Azure resource ID of a BYO Application Gateway for Containers resource to use for the migration.",
			Category: deploymentCategory,
		},
		&cli.StringFlag{
			Name:     "managed-subnet-id",
			Aliases:  []string{"m"},
			Usage:    "The Azure resource ID of a delegated subnet in your AKS cluster which will be used to create a managed Application Gateway for Containers deployment.",
			Category: deploymentCategory,
		},
		&cli.StringFlag{
			Name:    "provider",
			Aliases: []string{"p"},
			Usage:   "The provider of the source inputs, one of \"agic\", \"nginx\"",
		},
	}

	app := &cli.Command{
		Name:                          "AGIC to AGC Migration Tool",
		Usage:                         "Translates Application Gateway Ingress Controller Ingresses to Application Gateway for Containers manifests",
		UsageText:                     usage(),
		Description:                   description(),
		CustomRootCommandHelpTemplate: template(),
		Action: func(context.Context, *cli.Command) error {
			fmt.Println(usage())
			return errors.New("a command is required: use 'cluster' to read from a Kubernetes cluster or 'files' to read from YAML files")
		},

		Commands: []*cli.Command{
			{
				Name:      "cluster",
				Usage:     "Read Application Gateway Ingress Controller objects from a cluster",
				UsageText: "agic-migration cluster (--byo-resource-id ID | --managed-subnet-id ID) [OPTIONS]",
				Description: "Connects to a Kubernetes cluster and reads AGIC Ingress resources directly. " +
					"Will attempt to use in-cluster configuration first, then fall back to kubeconfig.\n" +
					"Requires exactly one of --byo-resource-id RESOURCE_ID or --managed-subnet-id RESOURCE_ID to be provided.",
				Flags: append(sharedFlags, &cli.StringFlag{
					Name:  "context",
					Usage: "Name of the kubernetes context to use",
				}),
				HideHelpCommand: true,
				Action: func(ctx context.Context, c *cli.Command) error {
					opts, err := setup(c)
					if err != nil {
						return err
					}
					// are we running inside a pod?
					k8sConfig, err := rest.InClusterConfig()
					if err != nil {
						if !errors.Is(err, rest.ErrNotInCluster) {
							return err
						}

						// not running inside a cluster
						k8sConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
							clientcmd.NewDefaultClientConfigLoadingRules(),
							&clientcmd.ConfigOverrides{
								CurrentContext: c.String("context"),
							},
						).ClientConfig()
						if err != nil {
							return err
						}
					}

					k8sClient, err := createK8sClient(k8sConfig)
					if err != nil {
						return fmt.Errorf("failed to create k8s client: %w", err)
					}

					createControllerClient, err := createControllerClient(k8sConfig)
					if err != nil {
						return fmt.Errorf("failed to create controller client: %w", err)
					}
					result, report, err := migration.MigrateCluster(ctx, k8sClient, createControllerClient, opts.Migration)
					if err != nil {
						return err
					}

					return output.Write(opts.Output, report, result)
				},
			},
			{
				Name:      "files",
				Usage:     "Read Application Gateway Ingress Controller manifests from YAML files",
				UsageText: "agic-migration files (--byo-resource-id ID | --managed-subnet-id ID) [OPTIONS] [file1] [file2] ...",
				Description: "Reads AGIC Ingress resources from one or more YAML files on disk. " +
					"Files can contain multiple Kubernetes resources.\n" +
					"Requires exactly one of --byo-resource-id RESOURCE_ID or --managed-subnet-id RESOURCE_ID to be provided.",
				Flags:           sharedFlags,
				ArgsUsage:       "[file1] [file2] ...",
				HideHelpCommand: true,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					opts, err := setup(cmd)
					if err != nil {
						return err
					}

					expanded, err := expandFileAndDirArgs(cmd.Args().Slice())
					if err != nil {
						return err
					}
					if len(expanded) == 0 {
						return fmt.Errorf("no input files found")
					}

					result, report, err := migration.MigrateFiles(ctx, opts.Migration, expanded...)
					if err != nil {
						return err
					}
					return output.Write(opts.Output, report, result)
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func expandFileAndDirArgs(inputs []string) ([]string, error) {
	seen := sets.Set[string]{}

	for _, in := range inputs {
		if in == "" {
			continue
		}

		info, err := os.Stat(in)
		if err != nil {
			return nil, fmt.Errorf("error processing file %q: %w", in, err)
		}

		if !info.IsDir() {
			seen.Insert(in)
			continue
		}

		err = filepath.WalkDir(in, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			switch strings.ToLower(filepath.Ext(p)) {
			case ".yaml", ".yml":
				seen.Insert(p)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error processing file %q: %w", in, err)
		}
	}

	return seen.UnsortedList(), nil
}

func setup(c *cli.Command) (options, error) {
	opts := options{}
	opts.Output.DryRun = c.Bool("dry-run")
	opts.Output.OutputDir = c.String("output-dir")
	opts.Output.OutputFile = c.String("output-file")
	opts.Migration.Conversion.BYOResourceID = c.String("byo-resource-id")
	opts.Migration.Conversion.ManagedSubnetID = c.String("managed-subnet-id")
	opts.Migration.Conversion.GatewayWAFID = c.String("waf-id")
	opts.Migration.Aggregation.IngressClassName = c.String("ingress-class")
	opts.Migration.ProviderName = c.String("provider")

	// Validate mutually exclusive flags
	hasBYO := opts.Migration.Conversion.BYOResourceID != ""
	hasManaged := opts.Migration.Conversion.ManagedSubnetID != ""

	if !hasBYO && !hasManaged {
		return opts, errors.New("exactly one of --byo-resource-id or --managed-subnet-id must be provided")
	}

	if hasBYO && hasManaged {
		return opts, errors.New("--byo-resource-id (-b) and --managed-subnet-id (-m) are mutually exclusive; provide only one")
	}

	if opts.Output.OutputDir != "" && opts.Output.OutputFile != "" {
		return opts, errors.New("--output-dir (-o) and --output-file are mutually exclusive; provide only one")
	}

	// TODO: move the default ingress class name into the provider interface
	if opts.Migration.ProviderName == "nginx" && opts.Migration.Aggregation.IngressClassName == "" {
		opts.Migration.Aggregation.IngressClassName = "nginx"
	}

	return opts, nil
}

func createK8sClient(k8sConfig *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(k8sConfig)
}

// In my experience this client is more of a nuisance to work with than the standard go client, but ingress2gateway
// uses this one so we have to create it.
func createControllerClient(k8sConfig *rest.Config) (ctrl_client.Client, error) {
	scheme := runtime.NewScheme()

	if err := core_v1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add corev1 to scheme: %w", err)
	}

	if err := networking_v1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add networkingv1 to scheme: %w", err)
	}

	if err := appgwrewrite.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add ApplicationGatewayRewrite to scheme: %w", err)
	}

	return ctrl_client.New(k8sConfig, ctrl_client.Options{
		Scheme: scheme,
	})
}

func template() string {
	exampleFmt := `
EXAMPLES:
 	Convert Application Gateway Ingress Controller objects in a cluster and write the new manifests to ./manifests, producing
 	Application Gateway for Containers resources where the ALB Controller will be managing the application resources:

 		$ %s cluster --managed-subnet-id "$SUBNET_ID" --output-dir ./manifests

 	Convert YAML manifests in ./source using a BYO Application Gateway for Containers resource, writing the converted
	manifests to ./manifests:

 		$ %s files --byo-resource-id "$AGC_ID" ./source --output-dir ./manifests

`
	progName := filepath.Base(os.Args[0])
	examples := fmt.Sprintf(exampleFmt, progName, progName)

	return cli.RootCommandHelpTemplate + examples
}

func usage() string {
	format := `%s cluster (--byo-resource-id AGC_ID | --managed-subnet-id SUBNET_ID) [--context NAME] [OPTIONS]
%s files (--byo-resource-id AGC_ID | --managed-subnet-id SUBNET_ID) <file...> [OPTIONS]

Exactly one of --byo-resource-id or --managed-subnet-id must be provided.
Without --output-dir, converted manifests are written to stdout. The migration report always prints to stdout.

By default, only Ingresses with class 'azure/application-gateway' are migrated. Use --ingress-class (-c) to
override this, for example: --ingress-class nginx for NGINX Ingress Controller resources.
Use --provider to specify the ingress controller type (agic or nginx).

Use "%s --help" for more detailed help.
`
	progName := filepath.Base(os.Args[0])

	return fmt.Sprintf(format, progName, progName, progName)
}

func description() string {
	return "Reads Kubernetes Ingress resources (from a live cluster or provided YAML files), " +
		"converts supported annotations and routing rules into Azure Application Gateway for Containers " +
		"Gateway API manifests, and outputs a migration report. Supports both Application Gateway Ingress Controller (AGIC) " +
		"and NGINX Ingress Controller annotations. Source objects are never modified and new objects are only printed to stdout or written to disk."
}

# Application Gateway for Containers Migration Utility

A CLI utility for migrating Kubernetes Ingress resources from **Application Gateway Ingress Controller (AGIC)** or **NGINX Ingress Controller** to **Azure Application Gateway for Containers** using Gateway API.

## Overview

The Application Gateway for Containers Migration Utility automates the translation of ingress annotations and routing rules into Gateway API resources compatible with [Azure Application Gateway for Containers](https://aka.ms/agc). It supports:

- **AGIC** (Application Gateway Ingress Controller) annotations
- **NGINX Ingress Controller** annotations
- Detailed migration reports with actionable recommendations
- Individual or consolidated Gateway API translation ready for deployment

## Prerequisites

- Go 1.25 or later
- Access to a Kubernetes cluster (for cluster mode) or YAML manifests (for files mode)

## Deployment Modes

Application Gateway for Containers supports two deployment strategies:

### Bring Your Own (BYO) Deployment

In BYO mode, you create the AGC resources in Azure first (via Portal, CLI, Terraform, etc.), then reference the resource ID when running the migration utility. Use `--byo-resource-id` to specify the existing AGC resource.

```bash
./agc-migration files --byo-resource-id "/subscriptions/.../applicationGateways/my-agc" ./manifests/*.yaml
```

See [Create Application Gateway for Containers - bring your own deployment](https://learn.microsoft.com/azure/application-gateway/for-containers/quickstart-create-application-gateway-for-containers-byo-deployment) for setup instructions.

### Managed by ALB Controller

In managed mode, the migration utility generates an `ApplicationLoadBalancer` custom resource that the ALB Controller uses to automatically create and manage the AGC resources in Azure. Use `--managed-subnet-id` to specify the delegated subnet.

```bash
./agc-migration cluster --managed-subnet-id "/subscriptions/.../subnets/agc-subnet"
```


See [Create Application Gateway for Containers managed by ALB Controller](https://learn.microsoft.com/azure/application-gateway/for-containers/quickstart-create-application-gateway-for-containers-managed-by-alb-controller) for setup instructions.

## Building the Utility

### Build for all platforms

```bash
./build.sh
```

This creates binaries in the `bin/` directory for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

## Running the Utility

The migration utility reads ingress resources and generates Gateway API YAML configurations that you can review and apply when ready.

> **Note:** This utility does **not** modify any existing resources on a live cluster.

### Migrate from YAML Files

Converts ingress YAML files into Gateway API resources:

```bash
# AGIC ingresses (default provider)
./agc-migration files --byo-resource-id "$AGC_ID" ./manifests/*.yaml

# NGINX ingresses
./agc-migration files --provider nginx --ingress-class nginx --managed-subnet-id "$SUBNET_ID" ./manifests/*.yaml

# Output to a file instead of stdout
./agc-migration files --byo-resource-id "$AGC_ID" --output-file converted.yaml ./manifests/*.yaml

# Output to a directory (one file per resource)
./agc-migration files --byo-resource-id "$AGC_ID" --output-dir ./output ./manifests/*.yaml
```

If you have Go installed you can run it directly without building first:
```bash
# AGIC ingresses (default provider)
go run ./cmd agc-migration files --byo-resource-id "$AGC_ID" ./manifests/*.yaml
```

### Migrate from a Live Cluster

Reads ingresses from your cluster and generates Gateway API resources. This is a **read-only operation** and does not modify or delete any existing ingresses or affect your production workloads.

```bash
# AGIC ingresses (default)
./agc-migration cluster --byo-resource-id "$AGC_ID"

# NGINX ingresses
./agc-migration cluster --provider nginx --ingress-class nginx --managed-subnet-id "$SUBNET_ID"

# Specify a kubeconfig context
./agc-migration cluster --context my-cluster --byo-resource-id "$AGC_ID"
```

### CLI Options

| Option | Description |
|--------|-------------|
| `--byo-resource-id`, `-b` | Azure resource ID of a BYO Application Gateway for Containers |
| `--managed-subnet-id`, `-m` | Azure resource ID of a delegated subnet for managed AGC |
| `--provider`, `-p` | Ingress controller type: `agic` (default) or `nginx` |
| `--ingress-class`, `-c` | IngressClass name to filter by (default: `azure/application-gateway`) |
| `--output-dir`, `-o` | Directory to write converted YAML files (one file per resource) |
| `--output-file` | Single file to write all converted YAML |
| `--dry-run` | Run migration without generating output files |
| `--waf-id`, `-w` | WAF policy ID to use in AGC |
| `--context` | Kubernetes context to use (cluster mode only; defaults to current kubeconfig context) |

## Running Tests

### Run all tests

```bash
go test ./... -v
```

### Run provider unit tests only

```bash
# AGIC provider tests
go test ./conversion/providers/agic/... -v

# NGINX provider tests
go test ./conversion/providers/nginx/... -v
```

### Run E2E tests

```bash
go test ./e2e/... -v
```

### Run tests with coverage

```bash
./coverage.sh
```

## Supported Annotations

### AGIC Annotations

The following Application Gateway Ingress Controller annotations are supported for migration:

| Annotation | Description | AGC Resource |
|------------|-------------|--------------|
| `appgw.ingress.kubernetes.io/request-timeout` | Request timeout in seconds | RoutePolicy |
| `appgw.ingress.kubernetes.io/cookie-based-affinity` | Enable session affinity | RoutePolicy |
| `appgw.ingress.kubernetes.io/backend-hostname` | Backend hostname rewrite | HTTPRoute URLRewrite |
| `appgw.ingress.kubernetes.io/backend-path-prefix` | Backend path prefix rewrite | HTTPRoute URLRewrite |
| `appgw.ingress.kubernetes.io/backend-protocol` | Backend protocol (HTTP/HTTPS) | BackendTLSPolicy |
| `appgw.ingress.kubernetes.io/ssl-redirect` | Redirect HTTP to HTTPS | HTTPRoute (redirect) |
| `appgw.ingress.kubernetes.io/appgw-ssl-profile` | Frontend TLS profile | FrontendTLSPolicy |
| `appgw.ingress.kubernetes.io/appgw-trusted-root-certificate` | Backend trusted CA | BackendTLSPolicy |
| `appgw.ingress.kubernetes.io/hostname-extension` | Additional hostnames | HTTPRoute hostnames |
| `appgw.ingress.kubernetes.io/health-probe-hostname` | Health probe hostname | HealthCheckPolicy |
| `appgw.ingress.kubernetes.io/health-probe-path` | Health probe path | HealthCheckPolicy |
| `appgw.ingress.kubernetes.io/health-probe-port` | Health probe port | HealthCheckPolicy |
| `appgw.ingress.kubernetes.io/health-probe-interval` | Health probe interval | HealthCheckPolicy |
| `appgw.ingress.kubernetes.io/health-probe-timeout` | Health probe timeout | HealthCheckPolicy |
| `appgw.ingress.kubernetes.io/health-probe-unhealthy-threshold` | Unhealthy threshold | HealthCheckPolicy |
| `appgw.ingress.kubernetes.io/health-probe-status-codes` | Expected status codes | HealthCheckPolicy |
| `appgw.ingress.kubernetes.io/waf-policy-for-path` | WAF policy ID | WAFPolicy |
| `appgw.ingress.kubernetes.io/rewrite-rule-set-custom-resource` | Custom rewrite rules | HTTPRoute filters |

### NGINX Annotations

The following NGINX Ingress Controller annotations are supported for migration:

| Annotation | Description | AGC Resource |
|------------|-------------|--------------|
| `nginx.ingress.kubernetes.io/rewrite-target` | URL rewrite target | HTTPRoute URLRewrite |
| `nginx.ingress.kubernetes.io/app-root` | Redirect from / to app root | HTTPRoute (redirect rule) |
| `nginx.ingress.kubernetes.io/use-regex` | Treat paths as regex | HTTPRoute path match |
| `nginx.ingress.kubernetes.io/x-forwarded-prefix` | Add X-Forwarded-Prefix header | HTTPRoute RequestHeaderModifier |
| `nginx.ingress.kubernetes.io/affinity` | Enable cookie session affinity | RoutePolicy |
| `nginx.ingress.kubernetes.io/session-cookie-expires` | Cookie expiration | RoutePolicy |
| `nginx.ingress.kubernetes.io/session-cookie-max-age` | Cookie max age | RoutePolicy |
| `nginx.ingress.kubernetes.io/ssl-redirect` | Redirect HTTP to HTTPS | HTTPRoute (redirect) |
| `nginx.ingress.kubernetes.io/force-ssl-redirect` | Force HTTPS redirect | HTTPRoute (redirect) |
| `nginx.ingress.kubernetes.io/backend-protocol` | Backend protocol (HTTP/HTTPS/GRPC/GRPCS) | BackendTLSPolicy |
| `nginx.ingress.kubernetes.io/upstream-vhost` | Custom Host header to backend | HTTPRoute URLRewrite |
| `nginx.ingress.kubernetes.io/permanent-redirect` | 301 redirect to URL | HTTPRoute RequestRedirect |
| `nginx.ingress.kubernetes.io/permanent-redirect-code` | Custom permanent redirect code | HTTPRoute RequestRedirect |
| `nginx.ingress.kubernetes.io/temporal-redirect` | 302 redirect to URL | HTTPRoute RequestRedirect |
| `nginx.ingress.kubernetes.io/temporal-redirect-code` | Custom temporal redirect code | HTTPRoute RequestRedirect |
| `nginx.ingress.kubernetes.io/from-to-www-redirect` | Redirect to www subdomain | HTTPRoute RequestRedirect |
| `nginx.ingress.kubernetes.io/enable-modsecurity` | Enable WAF | WAFPolicy |
| `nginx.ingress.kubernetes.io/enable-owasp-core-rules` | Enable OWASP rules | WAFPolicy |
| `nginx.ingress.kubernetes.io/modsecurity-transaction-id` | Request tracking ID | Informational (AGC uses trackingId) |
| `nginx.ingress.kubernetes.io/server-alias` | Additional server hostnames | HTTPRoute hostnames |
| `nginx.ingress.kubernetes.io/auth-tls-secret` | mTLS CA certificate secret | FrontendTLSPolicy |
| `nginx.ingress.kubernetes.io/auth-tls-verify-client` | Client cert verification mode | FrontendTLSPolicy |
| `nginx.ingress.kubernetes.io/proxy-read-timeout` | Request timeout | RoutePolicy |
| `nginx.ingress.kubernetes.io/load-balance` | Load balancing algorithm | (round_robin is default) |
| `nginx.ingress.kubernetes.io/canary` | Enable canary deployment | HTTPRoute |
| `nginx.ingress.kubernetes.io/canary-weight` | Canary traffic percentage | HTTPBackendRef.Weight |
| `nginx.ingress.kubernetes.io/canary-weight-total` | Custom weight total (default 100) | Processed with canary-weight |
| `nginx.ingress.kubernetes.io/canary-by-header` | Route by header presence | HTTPHeaderMatch |
| `nginx.ingress.kubernetes.io/canary-by-header-value` | Route by exact header value | HTTPHeaderMatch (Exact) |
| `nginx.ingress.kubernetes.io/canary-by-header-pattern` | Route by header regex pattern | HTTPHeaderMatch (Regex) |

## Migration Report

The utility generates a detailed migration report that includes:

- **Input Resources**: List of ingresses processed
- **Migrated Resources**: Gateway API resources created
- **Annotation Status**: Per-annotation migration status
  - `completed`: Successfully migrated
  - `warning`: Migrated with notes
  - `not-supported`: No AGC equivalent
  - `error`: Migration failed
- **Issues and Recommendations**: Actionable guidance for manual steps

## Project Structure

```
agicmigration/
├── aggregation/       # Resource collection from cluster/files
├── cmd/               # CLI entry point
├── conversion/        # Core conversion logic
│   └── providers/     # Provider implementations
│       ├── agic/      # AGIC annotation handlers
│       └── nginx/     # NGINX annotation handlers
├── crds/              # Custom Resource Definitions (Gateway API ALB extensions)
├── e2e/               # End-to-end tests
├── logging/           # Logging utilities
├── migration/         # High-level migration orchestration
├── output/            # Output formatting and writing
├── reporting/         # Migration report generation
├── resources/         # Resource types, annotations, and issue codes
└── testutil/          # Test utilities
```

## Adding a New Provider

To add support for a new ingress controller:

1. Create a new package under `conversion/providers/<name>/`
2. Implement the `Provider` interface with `GetAnnotationHandlers()`
3. Register the provider in `migration/migration.go`
4. Add issue codes to `resources/<name>_issues.go`

See [conversion/provider.go](conversion/provider.go) for details.

> **Note:** you may also need to make changes to aggregation/ and migration/ until the provider interface is expanded to include these aspects.

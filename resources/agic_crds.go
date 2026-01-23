package resources

import (
	appgwrewrite "github.com/Azure/application-gateway-kubernetes-ingress/pkg/apis/azureapplicationgatewayrewrite/v1beta1"
)

type AppGWRewriteContext struct {
	Object appgwrewrite.AzureApplicationGatewayRewrite
}

func NewAppGWRewriteContext(rewrite appgwrewrite.AzureApplicationGatewayRewrite) *AppGWRewriteContext {
	return &AppGWRewriteContext{
		Object: rewrite,
	}
}

# Changelog

All notable changes to the AGIC Migration Tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [v1.0.0]

### Added
- Initial release of the AGIC Migration Tool
- Support for migrating AGIC (Application Gateway Ingress Controller) annotations
- Support for migrating NGINX Ingress Controller annotations
- Auto-detection of ingress controller type based on ingress class
- BYO (Bring Your Own) deployment mode support
- Managed ALB Controller deployment mode support
- Detailed migration reports with actionable recommendations
- End-to-end test suite

### Supported AGIC Annotations
- `appgw.ingress.kubernetes.io/request-timeout` → RoutePolicy
- `appgw.ingress.kubernetes.io/cookie-based-affinity` → RoutePolicy
- `appgw.ingress.kubernetes.io/backend-hostname` → HTTPRoute URLRewrite
- `appgw.ingress.kubernetes.io/backend-path-prefix` → HTTPRoute URLRewrite
- `appgw.ingress.kubernetes.io/backend-protocol` → BackendTLSPolicy
- `appgw.ingress.kubernetes.io/ssl-redirect` → HTTPRoute (redirect)
- `appgw.ingress.kubernetes.io/appgw-ssl-profile` → FrontendTLSPolicy
- `appgw.ingress.kubernetes.io/appgw-trusted-root-certificate` → BackendTLSPolicy
- `appgw.ingress.kubernetes.io/hostname-extension` → HTTPRoute hostnames
- `appgw.ingress.kubernetes.io/health-probe-*` → HealthCheckPolicy
- `appgw.ingress.kubernetes.io/waf-policy-for-path` → WAFPolicy
- `appgw.ingress.kubernetes.io/rewrite-rule-set-custom-resource` → HTTPRoute filters

### Supported NGINX Annotations
- `nginx.ingress.kubernetes.io/rewrite-target` → HTTPRoute URLRewrite
- `nginx.ingress.kubernetes.io/app-root` → HTTPRoute (redirect rule)
- `nginx.ingress.kubernetes.io/use-regex` → HTTPRoute path match
- `nginx.ingress.kubernetes.io/affinity` → RoutePolicy
- `nginx.ingress.kubernetes.io/session-cookie-*` → RoutePolicy
- `nginx.ingress.kubernetes.io/ssl-redirect` → HTTPRoute (redirect)
- `nginx.ingress.kubernetes.io/force-ssl-redirect` → HTTPRoute (redirect)
- `nginx.ingress.kubernetes.io/backend-protocol` → BackendTLSPolicy
- `nginx.ingress.kubernetes.io/upstream-vhost` → HTTPRoute URLRewrite
- `nginx.ingress.kubernetes.io/permanent-redirect` → HTTPRoute RequestRedirect
- `nginx.ingress.kubernetes.io/temporal-redirect` → HTTPRoute RequestRedirect
- `nginx.ingress.kubernetes.io/from-to-www-redirect` → HTTPRoute RequestRedirect
- `nginx.ingress.kubernetes.io/enable-modsecurity` → WAFPolicy
- `nginx.ingress.kubernetes.io/enable-owasp-core-rules` → WAFPolicy
- `nginx.ingress.kubernetes.io/modsecurity-transaction-id` → Informational (AGC uses trackingId)
- `nginx.ingress.kubernetes.io/auth-tls-*` → FrontendTLSPolicy
- `nginx.ingress.kubernetes.io/proxy-read-timeout` → RoutePolicy
- `nginx.ingress.kubernetes.io/canary` → HTTPRoute (enables canary deployment)
- `nginx.ingress.kubernetes.io/canary-weight` → HTTPBackendRef.Weight
- `nginx.ingress.kubernetes.io/canary-weight-total` → Processed with canary-weight
- `nginx.ingress.kubernetes.io/canary-by-header` → HTTPHeaderMatch
- `nginx.ingress.kubernetes.io/canary-by-header-value` → HTTPHeaderMatch (Exact)
- `nginx.ingress.kubernetes.io/canary-by-header-pattern` → HTTPHeaderMatch (Regex)

[v1.0.0]: https://github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/tree/v1.0.0

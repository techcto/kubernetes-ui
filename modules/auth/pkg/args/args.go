// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package args

import (
	"flag"
	"fmt"
	"net"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"k8s.io/dashboard/csrf"
)

var (
	argPort                   = pflag.Int("port", 8000, "The port for auth service to listen on.")
	argAddress                = pflag.IP("address", net.IPv4(0, 0, 0, 0), "The IP address for auth service to serve on. Set 0.0.0.0 for listening on all interfaces.")
	argKubeconfig             = pflag.String("kubeconfig", "", "path to kubeconfig file")
	argApiServerHost          = pflag.String("apiserver-host", "", "address of the Kubernetes API server to connect to in the format of protocol://address:port, leave it empty if the binary runs inside cluster for local discovery attempt")
	argApiServerSkipTLSVerify = pflag.Bool("apiserver-skip-tls-verify", false, "enable if connection with remote Kubernetes API server should skip TLS verify")
	argApiServerCaBundle      = pflag.String("apiserver-ca-bundle", "", "file containing the x509 certificates used for HTTPS connection to the API Server")

	// OIDC (Keycloak) SSO login - see pkg/routes/oidc. Distinct from the
	// apiserver-* flags above, which are about reaching the K8s API, not
	// authenticating an end user via an external identity provider.
	argOIDCIssuerURL               = pflag.String("oidc-issuer-url", "", "Cloud's Keycloak realm issuer URL, e.g. https://keycloak.example.com/realms/cloud")
	argOIDCClientID                = pflag.String("oidc-client-id", "", "Keycloak client id registered for this cluster")
	argOIDCClientSecret            = pflag.String("oidc-client-secret", "", "Keycloak client secret registered for this cluster")
	argOIDCRedirectURL             = pflag.String("oidc-redirect-url", "", "callback URL registered with Keycloak, e.g. https://dashboard.example.com/api/v1/oidc/callback")
	argOIDCScopes                  = pflag.String("oidc-scopes", "openid profile email", "space-separated OIDC scopes to request")
	argSessionSigningKey           = pflag.String("session-signing-key", "", "key used to sign the session token issued after a successful OIDC login")
	argMarketplaceProductCode      = pflag.String("marketplace-product-code", "", "AWS Marketplace product code; when set, RegisterUsage must succeed before login is enabled")
	argMarketplacePublicKeyVersion = pflag.Int32("marketplace-public-key-version", 1, "AWS Marketplace public key version used by RegisterUsage")
)

func init() {
	// Init klog
	fs := flag.NewFlagSet("", flag.PanicOnError)
	klog.InitFlags(fs)

	// Default log level to 1
	_ = fs.Set("v", "1")

	pflag.CommandLine.AddGoFlagSet(fs)
	pflag.Parse()

	csrf.Ensure()
}

func KubeconfigPath() string {
	return *argKubeconfig
}

func ApiServerHost() string {
	return *argApiServerHost
}

func ApiServerSkipTLSVerify() bool {
	return *argApiServerSkipTLSVerify
}

func ApiServerCaBundle() string {
	return *argApiServerCaBundle
}

func Address() string {
	return fmt.Sprintf("%s:%d", *argAddress, *argPort)
}

func OIDCIssuerURL() string {
	return *argOIDCIssuerURL
}

func OIDCClientID() string {
	return *argOIDCClientID
}

func OIDCClientSecret() string {
	return *argOIDCClientSecret
}

func OIDCRedirectURL() string {
	return *argOIDCRedirectURL
}

func OIDCScopes() []string {
	return strings.Fields(*argOIDCScopes)
}

func OIDCEnabled() bool {
	return *argOIDCIssuerURL != "" && *argOIDCClientID != ""
}

func SessionSigningKey() string {
	return *argSessionSigningKey
}

func MarketplaceProductCode() string {
	return *argMarketplaceProductCode
}

func MarketplacePublicKeyVersion() int32 {
	return *argMarketplacePublicKeyVersion
}

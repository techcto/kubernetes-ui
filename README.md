# SpaceMade Kubernetes UI

```text
 _  __     _                          _             _   _ ___
| |/ /   _| |__   ___ _ __ _ __   ___| |_ ___  ___| | | |_ _|
| ' / | | | '_ \ / _ \ '__| '_ \ / _ \ __/ _ \/ __| | | || |
| . \ |_| | |_) |  __/ |  | | | |  __/ ||  __/\__ \ |_| || |
|_|\_\__,_|_.__/ \___|_|  |_| |_|\___|\__\___||___/\___/|___|
```

[![Go Report Card](https://goreportcard.com/badge/github.com/techcto/kubernetes-ui)](https://goreportcard.com/report/github.com/techcto/kubernetes-ui)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

SpaceMade Kubernetes UI is an open-source, web-based administration console for
Kubernetes. This repository provides its release images, Helm chart, security
updates, SSO integration, and AWS Marketplace packaging.

The project is based on the Apache-licensed Kubernetes Dashboard codebase and
retains its original copyright and license notices. Contributions, issue
reports, and feature requests are welcome in this repository.

## Features

- Workload, configuration, storage, discovery, and cluster administration UI.
- Keycloak and standards-based OIDC login using Authorization Code + PKCE.
- `view`, `edit`, and `admin` identity-provider roles mapped to Kubernetes RBAC.
- Signed application sessions for authenticated dashboard API access.
- AWS Marketplace entitlement enforcement using `RegisterUsage` and EKS IRSA.
- Multi-architecture container release workflows and a maintained Helm chart.

## Install with Helm

```console
helm repo add spacemade-kubernetes-ui https://techcto.github.io/kubernetes-ui/
helm repo update
helm upgrade --install kubernetes-dashboard \
  spacemade-kubernetes-ui/kubernetes-dashboard \
  --create-namespace \
  --namespace kubernetes-dashboard
```

The release images are published from this repository under
`ghcr.io/techcto/kubernetes-ui-*`. See the chart's
[`values.yaml`](charts/kubernetes-dashboard/values.yaml) for configuration.

For the managed AWS Marketplace experience, use the
[`techcto/aws-kubernetes`](https://github.com/techcto/aws-kubernetes) EKS
Quick Start after subscribing. It provisions the EKS stack, installs this UI,
configures IRSA for Marketplace metering, and can configure private SSO.

## SSO configuration

The auth service supports these flags:

| Flag | Purpose |
|---|---|
| `--oidc-issuer-url` | OIDC issuer URL, such as a Keycloak realm |
| `--oidc-client-id` | Registered OIDC client ID |
| `--oidc-client-secret` | Registered confidential-client secret |
| `--oidc-redirect-url` | Registered callback URL |
| `--oidc-scopes` | Space-separated scopes; defaults to `openid profile email` |
| `--session-signing-key` | Key used to sign application session tokens |
| `--marketplace-product-code` | Enables AWS Marketplace entitlement enforcement |

The Marketplace product code is optional for local and community deployments.
When configured, the auth service must successfully register Marketplace usage
before it starts accepting logins.

## Documentation and support

- [Documentation](docs/README.md)
- [User guide](docs/user/README.md)
- [Access control](docs/user/access-control/README.md)
- [Developer guide](DEVELOPMENT.md)
- [Issues and feature requests](https://github.com/techcto/kubernetes-ui/issues)
- [Security policy](SECURITY.md)
- [Contributing](CONTRIBUTING.md)

## License and attribution

Licensed under the [Apache License 2.0](LICENSE). This project is derived from
Kubernetes Dashboard and preserves the original Kubernetes Authors copyright
notices throughout the source tree.

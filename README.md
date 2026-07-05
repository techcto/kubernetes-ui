```
 _          _                          _                        _
| | ___   _| |__   ___ _ __ _ __   ___| |_ ___  ___       _   _(_)
| |/ / | | | '_ \ / _ \ '__| '_ \ / _ \ __/ _ \/ __|_____| | | | |
|   <| |_| | |_) |  __/ |  | | | |  __/ ||  __/\__ \_____| |_| | |
|_|\_\\__,_|_.__/ \___|_|  |_| |_|\___|\__\___||___/      \__,_|_|
```

# kubernetes-ui

[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes/dashboard)](https://goreportcard.com/report/github.com/kubernetes/dashboard)
[![Coverage Status](https://codecov.io/github/kubernetes/dashboard/coverage.svg?branch=master)](https://codecov.io/github/kubernetes/dashboard?branch=master)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/kubernetes/dashboard/blob/master/LICENSE)

## 🚀 Relaunching as a new open source project

`kubernetes-ui` (`techcto/kubernetes-ui`) picks up where `kubernetes/dashboard` left off, as its own
actively-developed project — and we're already moving fast. Huge thanks to everyone who built and
maintained the original project over the years; this fork exists because of the solid foundation
they left behind. **We're looking for developers who want to help take it to the next level** — open
an issue or PR if you'd like to get involved.

### What's new in this fork

#### 🔐 SSO login for Keycloak

We added full **SSO login support for Keycloak** (`modules/auth/pkg/routes/oidc`) — a complete OIDC
Authorization Code flow with PKCE. Upstream never shipped this: it only ever had bare bearer-token
login (`modules/auth/api/v1/login.go`), and login itself was tracked as a `priority/critical-urgent`
issue ([#2353](https://github.com/kubernetes/dashboard/issues/2353)) that was still open when the
project was archived.

On successful login:
- The user's Keycloak client role (`view` / `edit` / `admin`) is extracted straight from the ID
  token's `resource_access` claim.
- That role maps 1:1 onto Kubernetes' own built-in `view` / `edit` / `admin` ClusterRoles — no
  custom RBAC to maintain.
- A signed session token is issued for the frontend to use going forward.

**Configuration** (see `pkg/args/args.go`):

| Flag | Purpose |
|---|---|
| `--oidc-issuer-url` | Keycloak realm issuer URL |
| `--oidc-client-id` / `--oidc-client-secret` | Keycloak client credentials |
| `--oidc-redirect-url` | Registered callback URL |
| `--oidc-scopes` | Space-separated OIDC scopes to request |
| `--session-signing-key` | Key used to sign the post-login session token |

**Still to wire up end-to-end**: `modules/api` currently treats its incoming `Authorization` header
as a raw Kubernetes bearer token. It needs to decode this module's session token instead, and make
the actual API calls via the cluster's `dashboard-api` ServiceAccount (see
`charts/network/templates/admin-role.yaml` in the `aws-kubernetes` repo), using
`Impersonate-User` / `Impersonate-Group` headers set from the session's username/role — Kubernetes'
own native mechanism for "authenticate via an external IdP, then act as that user for RBAC purposes,"
since EKS's managed control plane won't trust an arbitrary external OIDC provider directly. The
`web` frontend also needs to start sending the session token it gets back.

## Introduction

Kubernetes Dashboard is a general purpose, web-based UI for Kubernetes clusters. It allows users to manage applications running in the cluster and troubleshoot them, as well as manage the cluster itself.

As of version 7.0.0, we have dropped support for Manifest-based installation. Only Helm-based installation is supported now. Due to multi-container setup and hard dependency on Kong gateway API proxy
it would not be feasible to easily support Manifest-based installation.

Additionally, we have changed the versioning scheme and dropped `appVersion` from Helm chart. It is because, with a multi-container setup, every module is now versioned separately. Helm chart version
can be considered an app version now.

![Dashboard UI workloads page](docs/images/overview.png)

## Installation

Kubernetes Dashboard supports only Helm-based installation currently as it is faster and gives us better control
over all dependencies required by Dashboard to run. We now use a single-container, DBless [Kong](https://hub.docker.com/r/kong/kong-gateway) installation
as a gateway that connects all our containers and exposes the UI. Users can then use any ingress controller or proxy
in front of kong gateway. To find out more about ways to customize your installation check out [helm chart values](charts/kubernetes-dashboard/values.yaml).

In order to install Kubernetes Dashboard simply run:
```console
# Add kubernetes-dashboard repository
helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/
# Deploy a Helm Release named "kubernetes-dashboard" using the kubernetes-dashboard chart
helm upgrade --install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --create-namespace --namespace kubernetes-dashboard
```

For more information about our Helm chart visit [ArtifactHub](https://artifacthub.io/packages/helm/k8s-dashboard/kubernetes-dashboard).

## Documentation

Dashboard documentation can be found in the [docs](docs/README.md) directory which contains:

* [Common](docs/common/README.md): Entry-level overview.
* [User Guide](docs/user/README.md): Helpful information for users.
* [How to access Dashboard](docs/user/accessing-dashboard/README.md) - Everything you need to know to get access to you Kubernetes Dashboard instance after installation.
* [Access Control](docs/user/access-control/README.md): Find out how to control access to your Kubernetes Dashboard and [create sample user](docs/user/access-control/creating-sample-user.md) that can be used to log in.
* [Developer Guide](DEVELOPMENT.md): Important information for contributors that would like to test, run and work on Dashboard locally.

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

* [**#sig-ui on Kubernetes Slack**](https://kubernetes.slack.com)
* [**kubernetes-sig-ui mailing list** ](https://groups.google.com/forum/#!forum/kubernetes-sig-ui)
* [**Issue tracker**](https://github.com/kubernetes/dashboard/issues)
* [**SIG info**](https://github.com/kubernetes/community/tree/master/sig-ui)
* [**Roles**](ROLES.md)

### Contribution

Learn how to start contributing to the [Contributing Guideline](CONTRIBUTING.md).

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

## License

[Apache License 2.0](https://github.com/kubernetes/dashboard/blob/master/LICENSE)

----
_Copyright 2019 [The Kubernetes Dashboard Authors](https://github.com/kubernetes/dashboard/graphs/contributors)_

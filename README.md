<br />
<div align="center">
  <a href="https://apollogeddon.github.io/ignition-tofu">
    <img src="webpage/public/favicon.png" alt="Logo" width="100" height="100">
  </a>
  <h3 align="center">Ignition Tofu Provider</h3>
  <p align="center">
    A Terraform and OpenTofu provider for Ignition that manages projects, connections and gateway settings through its REST API.
    <br />
    <a href="https://apollogeddon.github.io/ignition-tofu"><strong>Read the docs</strong></a>
    <br />
    <br />
    <a href="https://apollogeddon.github.io/ignition-tofu/docs/guides/installation/">Getting started</a>
    &middot;
    <a href="https://apollogeddon.github.io/ignition-tofu/docs/guides/capabilities/">Capabilities</a>
    &middot;
    <a href="https://apollogeddon.github.io/ignition-tofu/docs/guides/architecture/">Architecture</a>
  </p>
</div>

## Overview

`ignition` is an [OpenTofu](https://opentofu.org/) and Terraform provider that configures Ignition 8.3 gateways through the gateway's REST API. Use it to keep projects, database connections, tag providers, security settings, redundancy and Gateway Network settings in version control alongside the rest of your infrastructure.

The provider is developed and tested with OpenTofu. It uses plugin protocol 6, so Terraform 1.0 and later can also load it.

## Features

- **Infrastructure as code**: declare gateway resources in HCL, review changes with `plan`, and apply them with `apply`.
- **Secure by design**: database, SMTP, notification and identity provider secrets are encrypted through the gateway's own encryption endpoint before they are written to the gateway configuration.
- **Drift detection**: `plan` reads each resource back from the gateway and shows changes made in the Designer or the gateway web interface.
- **Gateway-wide settings**: redundancy, Gateway Network connections and settings, and OIDC and SAML identity providers.

## Requirements

- OpenTofu 1.6 or later, or Terraform 1.0 or later
- An Ignition gateway, version 8.3 or later (the REST API this provider uses was introduced in 8.3.0)
- An API key for that gateway

## Installation

The provider is not published to the OpenTofu or Terraform registry. Each [GitHub release](https://github.com/apollogeddon/ignition-tofu/releases) contains a zip for each platform, a `SHA256SUMS` file with its GPG signature, and the provider manifest, and the [docs site](https://apollogeddon.github.io/ignition-tofu/) serves every release as a provider network mirror. Point your CLI configuration (`~/.tofurc`) at it once:

```hcl
provider_installation {
  network_mirror {
    url     = "https://apollogeddon.github.io/ignition-tofu/providers/"
    include = ["registry.opentofu.org/apollogeddon/ignition"]
  }
  direct {
    exclude = ["registry.opentofu.org/apollogeddon/ignition"]
  }
}
```

Then require the provider, and `tofu init` installs it for your platform:

```hcl
terraform {
  required_providers {
    ignition = {
      source  = "apollogeddon/ignition"
      version = "~> 1.2"
    }
  }
}
```

For Terraform, use `~/.terraformrc` and `registry.terraform.io/apollogeddon/ignition`. See the [installation guide](https://apollogeddon.github.io/ignition-tofu/docs/guides/installation/) for Windows, offline installs from a local filesystem mirror, and using a locally built provider.

## Quick start

Configure the provider with the gateway's address and an API key, then declare resources:

```hcl
provider "ignition" {
  host  = "http://localhost:8088"
  token = var.ignition_token
}

resource "ignition_project" "example" {
  name        = "MyEnterpriseProject"
  title       = "Enterprise Dashboard"
  description = "Managed by OpenTofu"
  enabled     = true
}
```

Instead of setting `host` and `token` in configuration, you can set these environment variables:

| Variable | Description |
| :--- | :--- |
| `IGNITION_HOST` | The base URL of the gateway, for example `http://10.10.1.5:8088`. |
| `IGNITION_TOKEN` | An API key created on the gateway. |

## Supported resources

| Area | Resources |
| :--- | :--- |
| Core | `ignition_project`, `ignition_database_connection`, `ignition_tag_provider`, `ignition_user_source`, `ignition_identity_provider` |
| Connectivity | `ignition_opc_ua_connection`, `ignition_device`, `ignition_gan_outgoing` |
| Gateway settings | `ignition_redundancy`, `ignition_gan_settings`, `ignition_smtp_profile` |
| Alarming and auditing | `ignition_alarm_journal`, `ignition_alarm_notification_profile`, `ignition_audit_profile` |
| Data storage | `ignition_store_forward` |

Data sources are available for projects, database connections, tag providers, user sources, SMTP profiles and store-and-forward engines. See the [documentation](https://apollogeddon.github.io/ignition-tofu) for every attribute.

## Documentation

The full documentation is at [apollogeddon.github.io/ignition-tofu](https://apollogeddon.github.io/ignition-tofu/):

- [Getting started](https://apollogeddon.github.io/ignition-tofu/docs/guides/installation/): installing and configuring the provider.
- [Capabilities](https://apollogeddon.github.io/ignition-tofu/docs/guides/capabilities/): the Ignition resources and features it supports.
- [Architecture](https://apollogeddon.github.io/ignition-tofu/docs/guides/architecture/): how the provider is built.

CI and release automation are described in [`.github/WORKFLOWS.md`](.github/WORKFLOWS.md), and security reporting in [`.github/SECURITY.md`](.github/SECURITY.md).

## Contributing

Pull requests are welcome. The provider's tooling is pinned under `.forgego/` and run through [Task](https://taskfile.dev/):

```bash
go tool -modfile=.forgego/task/go.mod task hooks     # install the git hooks, once per clone
go tool -modfile=.forgego/task/go.mod task lint      # format and lint
go tool -modfile=.forgego/task/go.mod task test      # unit tests
docker compose up -d                                 # start a test gateway on localhost:8088
go tool -modfile=.forgego/task/go.mod task test:acc  # acceptance tests against that gateway
go tool -modfile=.forgego/task/go.mod task docs      # generate the provider reference into docs/
```

`task test` and `task test:acc` download the OpenTofu release pinned in `.github/scripts/install-tofu.sh` into `.bin/` and run against it. Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/), which release-please uses to version releases. `task docs` generates the provider reference into `docs/` with tfplugindocs, in the Terraform registry's layout, from each resource's schema descriptions and its examples under `examples/` (`resource.tf`, `data-source.tf` and `import.sh`). To preview it on the webpage, run `bash .github/scripts/ci/migrate-docs.sh` next. Git ignores both outputs, and CI generates them again for every build. Every attribute needs a description; `task test` fails on one without.

## License

Released under the [MIT License](LICENSE).

Ignition is a trademark of Inductive Automation.

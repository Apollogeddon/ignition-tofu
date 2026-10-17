---
title: Getting started
order: 1
description: How to install and configure the Ignition Terraform Provider.
---

This guide shows how to install the `ignition` provider from a GitHub release, connect it to an Ignition gateway, and check that it works.

## Requirements

- OpenTofu or Terraform. The provider is developed and tested with OpenTofu.
- An Ignition gateway, version 8.3 or later. The REST API the provider uses was introduced in Ignition 8.3.0.
- An API key for that gateway, with permission to read and change its configuration.

## Install the provider

The provider is not yet published to the OpenTofu or Terraform registry, so `tofu init` and `terraform init` cannot download it from a registry. Install it from a [GitHub release](https://github.com/apollogeddon/ignition-tofu/releases) instead. Each release contains:

| File | Contents |
| :--- | :--- |
| `terraform-provider-ignition_<version>_<os>_<arch>.zip` | The provider binary for one platform. |
| `terraform-provider-ignition_<version>_SHA256SUMS` | SHA-256 checksums of the zips. |
| `terraform-provider-ignition_<version>_SHA256SUMS.sig` | A detached GPG signature of the checksums. |
| `terraform-provider-ignition_<version>_manifest.json` | The provider manifest (plugin protocol 6.0). |

Releases are built for Linux, macOS, Windows and FreeBSD.

### Use the network mirror (recommended)

This site serves every release as a [provider network mirror](https://opentofu.org/docs/cli/config/config-file/#network_mirror) at `https://apollogeddon.github.io/ignition-tofu/providers/`. Point your CLI configuration at it once, and `tofu init` (or `terraform init`) then downloads the provider for your platform, picks up new releases, and checks each download against its release's `SHA256SUMS`.

Add this to your CLI configuration: `~/.tofurc` for OpenTofu, or `%APPDATA%\tofu.rc` on Windows.

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

For Terraform, put it in `~/.terraformrc` (`%APPDATA%\terraform.rc` on Windows) and use `registry.terraform.io/apollogeddon/ignition` in both lists. Every other provider still installs from its registry, through `direct`. If your CLI configuration already has a `provider_installation` block, add the `network_mirror` block and the `exclude` to it.

Then declare the provider in your configuration, for example in `versions.tf`:

```hcl
terraform {
  required_providers {
    ignition = {
      source = "apollogeddon/ignition"
    }
  }
}
```

Run `tofu init`. It records the provider's checksums in `.terraform.lock.hcl`. To record them for other platforms too, for a team or CI on a different OS, run `tofu providers lock -platform=linux_amd64 -platform=darwin_arm64` with the platforms you use.

The mirror carries checksums but not the GPG signature, so OpenTofu checks each zip against `SHA256SUMS` without verifying the signature. To verify it yourself, check `SHA256SUMS.sig` from the release with `gpg --verify`.

### Use a local filesystem mirror

To install without the network mirror, for example on a machine without internet access, place a release in a local filesystem mirror.

The zips are named in the packed layout that a [filesystem mirror](https://opentofu.org/docs/cli/config/config-file/#filesystem_mirror) expects, so you can place a zip in a mirror directory without unpacking it. On Linux and macOS, OpenTofu and Terraform search `~/.terraform.d/plugins` by default; on Windows the directory is `%APPDATA%\terraform.d\plugins`.

Download the release for your platform and verify its checksum:

<!-- x-release-please-start-version -->

```bash
VERSION=1.2.5 # the release to install
PLATFORM=linux_amd64 # for example darwin_arm64 or windows_amd64
MIRROR="$HOME/.terraform.d/plugins/registry.opentofu.org/apollogeddon/ignition"

mkdir -p "$MIRROR"
curl -fsSLO --output-dir "$MIRROR" "https://github.com/apollogeddon/ignition-tofu/releases/download/v${VERSION}/terraform-provider-ignition_${VERSION}_${PLATFORM}.zip"
curl -fsSLO --output-dir "$MIRROR" "https://github.com/apollogeddon/ignition-tofu/releases/download/v${VERSION}/terraform-provider-ignition_${VERSION}_SHA256SUMS"
(cd "$MIRROR" && sha256sum --check --ignore-missing "terraform-provider-ignition_${VERSION}_SHA256SUMS")
```

<!-- x-release-please-end -->

On macOS, use `shasum -a 256 --check --ignore-missing` in place of `sha256sum --check --ignore-missing`.

The directory path is the provider's full source address: `registry.opentofu.org/apollogeddon/ignition` matches `source = "apollogeddon/ignition"` in OpenTofu. Terraform expands the same short address to `registry.terraform.io/apollogeddon/ignition`, so use `registry.terraform.io` in the path when you use Terraform.

Declare the provider as for the network mirror and run `tofu init` (or `terraform init`). It installs the provider from the mirror directory and records its checksums in `.terraform.lock.hcl`.

The default mirror directories apply only when your CLI configuration has no `provider_installation` block. If it has one, add the mirror to it explicitly:

```hcl
provider_installation {
  filesystem_mirror {
    path    = "/home/me/.terraform.d/plugins"
    include = ["registry.opentofu.org/apollogeddon/ignition"]
  }
  direct {
    exclude = ["registry.opentofu.org/apollogeddon/ignition"]
  }
}
```

### Use a provider you built yourself

To run a provider built from source, build it and point a `dev_overrides` block in your CLI configuration (`~/.tofurc`, or `~/.terraformrc` for Terraform) at the directory that contains the binary:

```bash
go build -o "$HOME/ignition-provider/terraform-provider-ignition" .
```

```hcl
provider_installation {
  dev_overrides {
    "registry.opentofu.org/apollogeddon/ignition" = "/home/me/ignition-provider"
  }
  direct {}
}
```

For Terraform, use `registry.terraform.io/apollogeddon/ignition` as the key. With `dev_overrides`, skip `tofu init` for this provider and run `tofu plan` or `tofu apply` directly. OpenTofu prints a warning while an override is active.

## Configure the provider

The provider needs the gateway's base URL and an API key:

```hcl
provider "ignition" {
  host               = "http://localhost:8088"
  token              = var.ignition_token
  allow_insecure_tls = false # set to true for a gateway with a self-signed certificate
}
```

| Argument | Environment variable | Description |
| :--- | :--- | :--- |
| `host` | `IGNITION_HOST` | The base URL of the gateway, for example `http://10.10.1.5:8088`. |
| `token` | `IGNITION_TOKEN` | An API key created on the gateway. Marked sensitive. |
| `allow_insecure_tls` | | Skips TLS certificate verification. Defaults to `false`. |
| `request_timeout` | | The longest one request may take, in seconds, including its retries: a gateway that is down or answering with errors fails the run after this. From 1 to 3600; defaults to `60`. Raise it for a slow host, or a gateway whose modules take longer than a minute to restart after a configuration change. |

`allow_insecure_tls` is useful for local Docker gateways or gateways with the default self-signed certificate. Avoid it in production.

Keep the API key out of your `.tf` files. When `host` or `token` is not set, the provider reads the matching environment variable, so the provider block can be empty:

```hcl
provider "ignition" {}
```

If an attribute comes from a value only known after apply, such as the address of a gateway created in the same configuration, OpenTofu stops the plan with an error naming the attribute. A plan that can't reach the gateway would leave the resources already in state unrefreshed, so a gateway replaced in the same run would silently lack them. Apply what sets the value first (for example `tofu apply -target=<that resource>`), then plan the rest, or set the attribute from a value known at plan time. Terraform, which can defer a provider, plans this provider's resources in a later round instead.

```bash
export IGNITION_HOST="http://localhost:8088"
export IGNITION_TOKEN="<your API key>"
```

## Create an API key

1. Sign in to the gateway web interface as an administrator.
2. Open the API keys page in the gateway's security settings and create a key.
3. Give the key a descriptive name, for example `opentofu`, and grant it the security levels it needs to read and change the gateway configuration.
4. Copy the key when the gateway shows it. You cannot view it again later.

A key without the required security levels still authenticates, but the gateway rejects every request with `403 Forbidden`.

## Verify the installation

Read an existing project with a data source, then run `tofu init` and `tofu plan`:

```hcl
data "ignition_project" "example" {
  name = "MyProject"
}

output "project_description" {
  value = data.ignition_project.example.description
}
```

If the plan succeeds and shows the output, the provider can reach the gateway and the API key is valid.

#!/bin/bash
# Generates the provider docs into docs/ with tfplugindocs, from the schema OpenTofu
# reports for a freshly built provider.
set -euo pipefail

TF_PLUGIN_DOCS_VERSION="v0.24.0"

root=$(git rev-parse --show-toplevel)
cd "$root"
tofu=$(.github/scripts/install-tofu.sh)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "Exporting the provider schema with OpenTofu..."
go build -o "$work/bin/terraform-provider-ignition" .
mkdir -p "$work/config"
cat >"$work/tofurc" <<CONFIG
provider_installation {
  dev_overrides {
    "registry.opentofu.org/apollogeddon/ignition" = "$work/bin"
  }
  direct {}
}
CONFIG
cat >"$work/config/main.tf" <<'CONFIG'
terraform {
  required_providers {
    ignition = { source = "apollogeddon/ignition" }
  }
}
CONFIG
TF_CLI_CONFIG_FILE="$work/tofurc" "$tofu" -chdir="$work/config" providers schema -json >"$work/schema.json"

# tfplugindocs finds a provider named ignition only under Terraform's registry address
jq '.provider_schemas |= with_entries(.key = "registry.terraform.io/hashicorp/ignition")' \
  "$work/schema.json" >"$work/providers-schema.json"

echo "Generating documentation..."
go run "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@$TF_PLUGIN_DOCS_VERSION" generate \
  --provider-name ignition --providers-schema "$work/providers-schema.json"

# the exported schema has no defaults, so tfplugindocs can't show them: add them from the provider's own schemas
go run ./internal/tools/docdefaults docs

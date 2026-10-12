---
page_title: "Bootstrapping a Gateway with Docker"
subcategory: ""
description: |-
  Provisioning and configuring an Ignition Gateway from a single apply.
---

# Bootstrapping a gateway with Docker

This guide shows how to create an Ignition gateway and configure it in a single `tofu apply` (or `terraform apply`).

The `ignition` provider configures a gateway that is already running. It cannot install Ignition, because there is no REST API to call until the gateway exists. To build a gateway from a seed backup, restore it on first start and then configure it, pair this provider with the [`kreuzwerker/docker`](https://registry.terraform.io/providers/kreuzwerker/docker/latest) provider in the same root module.

This is the pattern this repository's `docker-compose.yml` uses for acceptance testing: copy a seed `.gwbk` backup into a derived image, restore it against the running container, then restart the gateway to apply the restore.

The `ignition` provider is not yet published to a public registry. Install it from a GitHub release first, as described in the [installation guide](https://apollogeddon.github.io/ignition-tofu/docs/guides/installation/). The example below uses the source address `apollogeddon/ignition`, as the installation guide does, so the network or filesystem mirror it describes serves it to OpenTofu and Terraform alike.

## The pattern

```terraform
terraform {
  required_providers {
    docker   = { source = "kreuzwerker/docker" }
    ignition = { source = "apollogeddon/ignition" }
  }
}

resource "docker_image" "gateway" {
  name = "my-ignition-gateway"
  build {
    context    = path.module
    dockerfile = "Dockerfile" # FROM inductiveautomation/ignition:...  +  COPY seed.gwbk /restore.gwbk
  }
}

resource "docker_container" "gateway" {
  name  = "ignition-gateway"
  image = docker_image.gateway.image_id

  env = [
    "ACCEPT_IGNITION_EULA=Y",
    "GATEWAY_ADMIN_USERNAME=admin",
    "GATEWAY_ADMIN_PASSWORD=password123",
    "IGNITION_EDITION=standard",
  ]

  ports {
    internal = 8088
    external = 8088 # fixed; see "Things to know" below
  }

  healthcheck {
    test     = ["CMD", "curl", "-f", "http://localhost:8088/system/gwinfo"]
    interval = "10s"
    timeout  = "5s"
    retries  = 12
  }

  wait         = true
  wait_timeout = 300

  # Restore against the running gateway, then restart to apply the restore.
  # See "Things to know" for why this is preferred over the `-r` boot flag.
  provisioner "local-exec" {
    command = <<-EOT
      set -e
      docker exec ${self.name} /usr/local/bin/ignition/gwcmd.sh -s /restore.gwbk -y
      docker exec ${self.name} /usr/local/bin/ignition/gwcmd.sh -r
      timeout 120 bash -c 'until curl -sf http://localhost:8088/system/gwinfo >/dev/null; do sleep 3; done'
    EOT
  }
}

provider "ignition" {
  host = "http://localhost:8088"
}

resource "ignition_tag_provider" "example" {
  name       = "MyTags"
  type       = "STANDARD"
  depends_on = [docker_container.gateway]
}
```

A complete, runnable version of this configuration is in
[`examples/bootstrap-docker`](https://github.com/apollogeddon/ignition-tofu/tree/main/examples/bootstrap-docker).
It builds its image from this repository's
[`bootstrap/Dockerfile`](https://github.com/apollogeddon/ignition-tofu/blob/main/bootstrap/Dockerfile).

## Things to know

### Restore against the running gateway

Ignition's Docker image can restore a backup on first start if you pass
`-r /restore.gwbk` as the container's `command`. That is simpler than the
provisioner above, but for a backup that contains API keys with non-default
security level grants, the grants may not survive the restore: the key still
authenticates, but every API call returns `403 Forbidden`. The boot-time
restore runs before the gateway has finished initializing its default
security configuration.

Restoring against a running gateway, as above, does not have this problem. It
is the same sequence `docker-compose.yml` uses for this repository's
acceptance tests. If your seed backup has only the default security
configuration, you can try the `-r` flag, but confirm that an API call with
your key succeeds before you rely on it.

### Use a fixed external port

A `provider` block cannot use another resource's computed attributes in the
same apply: OpenTofu and Terraform need to know how to reach the provider's
API before they can plan the resources that use it. Fixing the external port
at `8088` lets the `ignition` provider's `host` be a plain string rather than
a reference to the container.

### Add `depends_on` to every `ignition_*` resource

`provider` blocks do not support `depends_on`, and because `host` is a static
string rather than a reference to the container, OpenTofu cannot infer that
the gateway must exist first. Without an explicit dependency it may create the
container and configure the gateway in parallel. For a configuration with
many resources, put them in a child module and set `depends_on` once on the
`module` block:

```terraform
module "gateway_config" {
  source     = "./config"
  depends_on = [docker_container.gateway]
}
```

`wait = true` on `docker_container` blocks its apply until the health check
passes, and the `local-exec` provisioner blocks until the gateway has
restarted after the restore. By the time a dependent resource runs, the
gateway accepts API calls. The client's retries (up to 10 per request) are a
safety margin, not what makes this reliable.

## Other platforms

The same approach works for a virtual machine or a bare-metal install. Use a
`terraform_data` or `null_resource` resource with a `remote-exec` or
`local-exec` provisioner (or cloud-init or user data) to run the Ignition
installer and an equivalent restore step. Then configure the provider the same
way: a static host that is known in advance, and an explicit `depends_on`.
These steps depend on your cloud and operating system, so this repository does
not include a worked example.

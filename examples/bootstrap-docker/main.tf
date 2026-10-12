terraform {
  required_providers {
    docker = {
      source = "kreuzwerker/docker"
    }
    ignition = {
      source = "apollogeddon/ignition"
    }
  }
}

# Bake the seed backup into the gateway image so the very first boot restores
# a known configuration (see bootstrap/Dockerfile). This is the same pattern
# used by docker-compose.yml for acceptance testing, expressed as a native
# Terraform resource instead of Docker Compose.
resource "docker_image" "gateway" {
  name = "ignition-tofu-bootstrap"
  build {
    context    = "${path.module}/../.."
    dockerfile = "bootstrap/Dockerfile"
  }
}

resource "docker_container" "gateway" {
  name  = "ignition-gateway"
  image = docker_image.gateway.image_id

  # No "-r" boot-time restore flag here: for this backup, restoring via that
  # native mechanism does not preserve the API key's security-level grants —
  # the token authenticates but every request 403s. Restore live instead
  # (below), which is what this repo's own docker-compose.yml does
  # successfully for acceptance testing.
  env = [
    "ACCEPT_IGNITION_EULA=Y",
    "GATEWAY_ADMIN_USERNAME=admin",
    "GATEWAY_ADMIN_PASSWORD=password123",
    "IGNITION_EDITION=standard",
  ]

  ports {
    internal = 8088
    # Fixed, not random: the "ignition" provider block below needs a host it
    # can reference before this container exists, and provider blocks can't
    # depend on a resource's computed attributes in the same apply.
    external = 8088
  }

  healthcheck {
    test     = ["CMD", "curl", "-f", "http://localhost:8088/system/gwinfo"]
    interval = "10s"
    timeout  = "5s"
    retries  = 12
  }

  # Block this resource's own apply until the gateway is actually answering
  # requests, not just "container started" — required before gwcmd.sh below
  # can talk to it.
  wait         = true
  wait_timeout = 300

  # Restore live against the running gateway, then restart it to apply the
  # restore, then wait for it to come back up — gwcmd.sh triggers a JVM
  # restart that docker_container's own wait/healthcheck can't observe,
  # since it happens entirely inside the container after Terraform already
  # considers this resource created.
  provisioner "local-exec" {
    command = <<-EOT
      set -e
      docker exec ${self.name} /usr/local/bin/ignition/gwcmd.sh -s /restore.gwbk -y
      docker exec ${self.name} /usr/local/bin/ignition/gwcmd.sh -r
      timeout 120 bash -c 'until curl -sf http://localhost:8088/system/gwinfo >/dev/null; do sleep 3; done'
    EOT
  }
}

# host is a static value matching the fixed port above, not a reference to
# docker_container.gateway — see the comment above.
provider "ignition" {
  host = "http://localhost:8088"
  # token is read from the IGNITION_TOKEN environment variable if unset here.
}

# Every ignition_* resource needs an explicit dependency on the container:
# provider blocks don't support depends_on, and since "host" above is a
# static string rather than a reference, Terraform has no other way to know
# the gateway must exist first. For a real configuration with many
# resources, put them in a child module and set depends_on on the module
# call instead of repeating this on every resource.
resource "ignition_tag_provider" "example" {
  name = "BootstrapDemo"
  type = "STANDARD"

  depends_on = [docker_container.gateway]
}

terraform {
  required_providers {
    ignition = {
      source  = "apollogeddon/ignition"
      version = "~> 1.3"
    }
  }
}

variable "ignition_token" {
  description = "The gateway API token, e.g. from TF_VAR_ignition_token; never commit one"
  type        = string
  sensitive   = true
}

# host and token fall back to IGNITION_HOST and IGNITION_TOKEN when they are left out
provider "ignition" {
  host            = "https://ignition.example.com"
  token           = var.ignition_token
  request_timeout = 120
}

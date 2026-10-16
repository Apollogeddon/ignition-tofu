---
title: Capabilities
order: 3
description: Overview of supported Ignition resources and features.
---

This page lists the gateway configuration the `ignition` provider manages, the data sources it offers, and how to bring existing gateway configuration under management. Each resource's attributes are described in the reference section.

## Resources

### Core

| Resource | Description |
| :--- | :--- |
| `ignition_project` | Projects, including inheritance from a parent project and project-level defaults. |
| `ignition_database_connection` | Database connections: MariaDB, MySQL, PostgreSQL, SQL Server and Oracle. |
| `ignition_tag_provider` | Realtime tag providers. |
| `ignition_user_source` | User sources, such as internal, Active Directory and database sources. |
| `ignition_identity_provider` | Identity providers: internal, OpenID Connect (OIDC) and SAML. |

### Connectivity

| Resource | Description |
| :--- | :--- |
| `ignition_opc_ua_connection` | Outgoing OPC UA client connections. |
| `ignition_device` | Devices, such as Modbus TCP, Siemens and simulator devices, with type-specific parameters given as JSON. |
| `ignition_gan_outgoing` | Outgoing Gateway Network connections to other gateways. |

### Gateway settings

| Resource | Description |
| :--- | :--- |
| `ignition_redundancy` | Singleton. The gateway's redundancy role (independent, master or backup) and synchronization settings. |
| `ignition_gan_settings` | Singleton. General Gateway Network settings, such as SSL requirements and the security policy for incoming connections. |
| `ignition_smtp_profile` | SMTP profiles for email notifications and reports. |

### Alarming and auditing

| Resource | Description |
| :--- | :--- |
| `ignition_alarm_journal` | Alarm journals for alarm history, of type `DATASOURCE`, `LOCAL` or `REMOTE`. |
| `ignition_audit_profile` | Audit profiles, of type `database`, `local`, `remote` or `edge`. |
| `ignition_alarm_notification_profile` | Email alarm notification profiles. |

### Data storage

| Resource | Description |
| :--- | :--- |
| `ignition_store_forward` | Store-and-forward engines that buffer data while a database is unavailable. |

## Data sources

Data sources read configuration that already exists on a gateway, including configuration that OpenTofu does not manage. They are available for:

- `ignition_project`
- `ignition_database_connection`
- `ignition_tag_provider`
- `ignition_user_source`
- `ignition_smtp_profile`
- `ignition_store_forward`

For example, to create a project that inherits from an existing one:

```hcl
data "ignition_project" "global" {
  name = "global"
}

resource "ignition_project" "site_a" {
  name   = "site_a"
  parent = data.ignition_project.global.name
}
```

## Import existing configuration

To manage configuration that already exists on a gateway, import it with `tofu import` (or `terraform import`). Every named resource is imported by its name on the gateway:

```bash
# Import the database connection named "ProductionDB"
tofu import ignition_database_connection.main ProductionDB

# Import the project named "MainDashboard"
tofu import ignition_project.main MainDashboard
```

The singleton resources, `ignition_redundancy` and `ignition_gan_settings`, do not support import and do not need it. Every gateway already has these settings, so creating the resource takes over the existing settings and applies your configuration to them.

Destroying a singleton cannot remove the settings from the gateway:

- Destroying `ignition_redundancy` resets the gateway to the independent role with default settings.
- Destroying `ignition_gan_settings` removes the resource from state and leaves the gateway's settings as they are.

## Behavior

- **Type-specific settings**: resources such as `ignition_identity_provider`, `ignition_alarm_journal` and `ignition_audit_profile` take a `type` attribute, and the settings that apply depend on it. `ignition_device` takes its type-specific settings as a JSON `parameters` string. Set only the settings you need: the gateway gives the rest its defaults, and the provider compares only the keys you set, so those defaults don't show as drift.
- **Secrets**: passwords and client secrets are encrypted by the gateway's encryption endpoint before the provider writes them to the gateway configuration. They are marked sensitive, so `plan` output hides them, but like any secret in your configuration they are stored in your state. Store state somewhere secure.
- **Drift detection**: `plan` reads each resource from the gateway and shows changes made in the Designer or the gateway web interface. `apply` reverts them to your configuration.

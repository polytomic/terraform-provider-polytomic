terraform {
  required_providers {
    polytomic = {
      source = "polytomic/polytomic"
    }
  }
}

provider "polytomic" {
  # Configuration comes from environment variables:
  # POLYTOMIC_DEPLOYMENT_URL
  # POLYTOMIC_API_KEY or POLYTOMIC_DEPLOYMENT_KEY or POLYTOMIC_PARTNER_KEY
}

data "polytomic_caller_identity" "self" {}

locals {
  organization_id = data.polytomic_caller_identity.self.organization_id
}

# Connection data sources
data "polytomic_hubspot_connection" "hubspot" {
  id = "6723ae9f-2b55-42d2-bc85-eed65a291cf8"
}

data "polytomic_postgresql_connection" "postgresql" {
  id = "5b740e0a-80f8-49dc-835c-d24e57fad02b"
}

# Create a model using PostgreSQL with a SQL query
resource "polytomic_model" "postgres_users" {
  connection_id = data.polytomic_postgresql_connection.postgresql.id
  name          = "Terraform Test Users"
  configuration = jsonencode({
    query = "SELECT * FROM users"
  })
  fields     = ["email", "first_name", "last_name", "id"]
  identifier = "id"
}

# Create a model sync from PostgreSQL to HubSpot contacts
resource "polytomic_sync" "postgres_to_hubspot_contacts" {
  name   = "Terraform Test: Postgres to HubSpot Contacts"
  active = false
  mode   = "updateOrCreate"
  schedule = {
    frequency = "manual"
  }

  # Target configuration
  target = {
    configuration = jsonencode({
      ignore_additional_emails = true
    })
    connection_id = data.polytomic_hubspot_connection.hubspot.id
    object        = "contacts"
    search_values = jsonencode({
      targetObject = "contacts"
    })
  }

  # Identity mapping using email field
  identity = {
    function = "Equality"
    source = {
      field    = "email"
      model_id = polytomic_model.postgres_users.id
    }
    target = "email"
  }

  # Field mappings
  fields = [
    {
      source = {
        field    = "first_name"
        model_id = polytomic_model.postgres_users.id
      }
      target = "firstname"
    },
    {
      source = {
        field    = "last_name"
        model_id = polytomic_model.postgres_users.id
      }
      target = "lastname"
    }
  ]

  sync_all_records = false
}

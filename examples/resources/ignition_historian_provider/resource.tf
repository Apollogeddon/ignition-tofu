# Store tag history in the "production_db" connection, a table per month,
# deleting months older than two years. Tags name it in historyProvider.
resource "ignition_historian_provider" "example" {
  name            = "Plant"
  database        = "production_db"
  partition_units = "MONTH"
  pruning_enabled = true
  pruning_age     = 2
  pruning_units   = "YEAR"
}

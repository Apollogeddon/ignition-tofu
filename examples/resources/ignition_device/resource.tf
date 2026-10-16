resource "ignition_device" "example" {
  name = "Simulator1"
  type = "ProgrammableSimulatorDevice"
  # Only the settings you set: the gateway gives the rest its defaults, and
  # the provider compares only these keys, so the defaults never show as drift.
  parameters = jsonencode({
    timeIntervalRate = 1000
  })
}

# Import the device named "Simulator1" on the gateway. The imported state holds
# every setting, so the first plan updates "parameters" to the keys you set;
# that update resets any setting you leave out to its default, so set each one
# you have changed on the gateway.
tofu import ignition_device.example Simulator1

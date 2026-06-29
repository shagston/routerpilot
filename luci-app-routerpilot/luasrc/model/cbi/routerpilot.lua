-- RouterPilot configuration model
-- UCI config: /etc/config/routerpilot

m = Map("routerpilot", translate("RouterPilot Settings"),
	translate("Configure the RouterPilot network automation daemon."))

s = m:section(TypedSection, "settings", translate("Daemon Settings"))
s.anonymous = true

o = s:option(Value, "port", translate("Listen port"))
o.default = "8080"
o.datatype = "port"
o.description = translate("HTTP server port for the RouterPilot API and Web UI.")

o = s:option(Value, "host", translate("Listen host"))
o.default = "0.0.0.0"
o.datatype = "host"
o.description = translate("Interface to bind to. Use 0.0.0.0 for all interfaces.")

o = s:option(ListValue, "log_level", translate("Log level"))
o.default = "info"
o:value("debug", translate("Debug"))
o:value("info", translate("Info"))
o:value("warn", translate("Warning"))
o:value("error", translate("Error"))

o = s:option(Flag, "read_only", translate("Read-only mode"))
o.default = "0"
o.description = translate("Block all write and admin operations. Safe for monitoring.")

o = s:option(Flag, "dry_run", translate("Dry-run mode"))
o.default = "0"
o.description = translate("Simulate all operations without making actual changes.")

return m

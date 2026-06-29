local m = Map("routerpilot", translate("RouterPilot — General Settings"),
	translate("Configure the RouterPilot daemon behavior."))

local s = m:section(TypedSection, "settings", translate("Daemon"))
s.anonymous = true

local port = s:option(Value, "port", translate("HTTP port"))
port.default = "8080"
port.datatype = "port"

local host = s:option(Value, "host", translate("Bind address"))
host.default = "0.0.0.0"
host.datatype = "host"

local log_level = s:option(ListValue, "log_level", translate("Log level"))
log_level.default = "info"
log_level:value("debug")
log_level:value("info")
log_level:value("warn")
log_level:value("error")

local read_only = s:option(Flag, "read_only", translate("Read-only mode"))
read_only.default = "1"
read_only.description = translate("Block all write/admin operations.")

local dry_run = s:option(Flag, "dry_run", translate("Dry-run mode"))
dry_run.default = "0"
dry_run.description = translate("Simulate all operations (no real changes).")

local plugin_dir = s:option(Value, "plugin_dir", translate("Plugin directory"))
plugin_dir.default = "plugins"
plugin_dir.description = translate("Directory for external plugin binaries.")

return m

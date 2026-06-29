local m = Map("routerpilot", translate("RouterPilot — Scheduled Tasks"),
	translate("Define cron-like automation rules. RouterPilot will execute these intents on a schedule."))

local s = m:section(TypedSection, "schedule", translate("Tasks"))
s.template = "cbi/tblsection"
s.addremove = true
s.anonymous = true

s:option(Flag, "enabled", translate("Enabled")).default = "0"

local name = s:option(Value, "name", translate("Name"))
name.datatype = "string"
name.rmempty = false

local cron = s:option(Value, "cron", translate("Cron expression"))
cron.datatype = "string"
cron.default = "0 4 * * *"
cron.description = translate("Format: minute hour day month weekday. Example: 0 4 * * * = daily at 4:00")

local intent = s:option(Value, "intent", translate("Intent"))
intent.datatype = "string"
intent.rmempty = false
intent.description = translate("RouterPilot intent to execute (e.g. dns.flush, diagnose, system.reboot)")

local args = s:option(Value, "args", translate("Arguments (JSON)"))
args.datatype = "string"
args.default = "{}"
args.description = translate("Optional JSON arguments for the intent. Example: {\"target\":\"8.8.8.8\"}")

return m

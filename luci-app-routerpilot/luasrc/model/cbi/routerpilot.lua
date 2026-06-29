local m = Map("routerpilot", translate("RouterPilot Settings"))

-- General
local gen = m:section(TabSection, "general", translate("General"))
gen.addremove = false

local port = gen:option(Value, "port", translate("HTTP port"))
port.default = "8080"
port.datatype = "port"

local host = gen:option(Value, "host", translate("Bind address"))
host.default = "0.0.0.0"
host.datatype = "host"

local log_level = gen:option(ListValue, "log_level", translate("Log level"))
log_level.default = "info"
log_level:value("debug")
log_level:value("info")
log_level:value("warn")
log_level:value("error")

local read_only = gen:option(Flag, "read_only", translate("Read-only mode"))
read_only.default = "1"

local dry_run = gen:option(Flag, "dry_run", translate("Dry-run mode"))
dry_run.default = "0"

-- Telegram
local tel = m:section(TabSection, "telegram", translate("Telegram"))
tel.addremove = false

local tel_enabled = tel:option(Flag, "tel_enabled", translate("Enable"))
tel_enabled.default = "0"

local tel_token = tel:option(Value, "tel_token", translate("Bot token"))
tel_token.datatype = "string"
tel_token.password = true

local tel_ids = tel:option(Value, "tel_allowed_ids", translate("Allowed chat IDs"))
tel_ids.datatype = "string"

-- LLM
local llm = m:section(TabSection, "llm", translate("LLM"))
llm.addremove = false

local ptype = llm:option(ListValue, "llm_type", translate("Planner type"))
ptype.default = "simple"
ptype:value("simple", translate("Simple (rule-based)"))
ptype:value("llm", translate("LLM (OpenAI-compatible)"))

local api_key = llm:option(Value, "llm_api_key", translate("API key"))
api_key.datatype = "string"
api_key.password = true

local endpoint = llm:option(Value, "llm_endpoint", translate("API endpoint"))
endpoint.default = "https://api.openai.com/v1"
endpoint.datatype = "string"

local model = llm:option(Value, "llm_model", translate("Model name"))
model.default = "gpt-4"
model.datatype = "string"

-- Schedule tasks
local sched = m:section(TabSection, "schedule", translate("Schedule"))
sched.addremove = false

local tasks = sched:option(DynamicList, "sched_tasks", translate("Tasks (cron format)"))
tasks.datatype = "string"
tasks.description = translate("Format: <name> <cron> <intent> <args_json>. Example: daily_reboot 0 4 * * * system.reboot {}")

return m

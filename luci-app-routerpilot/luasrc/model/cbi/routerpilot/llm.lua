local m = Map("routerpilot", translate("RouterPilot — LLM Settings"),
	translate("Configure the LLM planner. RouterPilot supports simple rule-based planning (no API key needed) or LLM-based planning via OpenAI-compatible APIs."))

local s = m:section(TypedSection, "llm", translate("LLM Provider"))
s.anonymous = true

local ptype = s:option(ListValue, "type", translate("Planner type"))
ptype.default = "simple"
ptype:value("simple", translate("Simple (rule-based, no API key)"))
ptype:value("llm", translate("LLM (OpenAI-compatible API)"))

local api_key = s:option(Value, "api_key", translate("API key"))
api_key.datatype = "string"
api_key.password = true
api_key.description = translate("API key for the LLM provider (e.g. OpenAI, Ollama).")

local endpoint = s:option(Value, "endpoint", translate("API endpoint"))
endpoint.default = "https://api.openai.com/v1"
endpoint.datatype = "string"
endpoint.description = translate("API base URL. For Ollama: http://<host>:11434/v1")

local model = s:option(Value, "model", translate("Model name"))
model.default = "gpt-4"
model.datatype = "string"
model.description = translate("Model identifier (e.g. gpt-4, gpt-3.5-turbo, llama3)")

return m

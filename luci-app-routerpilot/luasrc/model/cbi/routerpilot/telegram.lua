local m = Map("routerpilot", translate("RouterPilot — Telegram Bot"),
	translate("Configure Telegram bot integration. Create a bot via @BotFather on Telegram and paste the token here."))

local s = m:section(TypedSection, "telegram", translate("Telegram Bot"))
s.anonymous = true

local enabled = s:option(Flag, "enabled", translate("Enable"))
enabled.default = "0"

local token = s:option(Value, "token", translate("Bot token"))
token.datatype = "string"
token.password = true
token.description = translate("Token from @BotFather (format: 123456:ABC-DEF...). Stored in UCI, passed as ROUTERPILOT_TELEGRAM_TOKEN.")

local allowed_ids = s:option(Value, "allowed_ids", translate("Allowed chat IDs"))
allowed_ids.datatype = "string"
allowed_ids.description = translate("Comma-separated Telegram user/chat IDs allowed to send commands. Leave empty to allow all.")

return m

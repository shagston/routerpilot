module("luci.controller.routerpilot", package.seeall)

function index()
	entry({"admin", "services", "routerpilot"}, alias("admin", "services", "routerpilot", "dashboard"), _("RouterPilot"), 60).dependent = false

	entry({"admin", "services", "routerpilot", "dashboard"}, template("routerpilot_dashboard"), _("Dashboard"), 10).dependent = false
	entry({"admin", "services", "routerpilot", "execute"},   template("routerpilot_execute"),   _("Execute"),     20).dependent = false
	entry({"admin", "services", "routerpilot", "chat"},       template("routerpilot_chat"),       _("Chat"),        30).dependent = false

	entry({"admin", "services", "routerpilot", "settings"},        alias("admin", "services", "routerpilot", "settings", "general"), _("Settings"), 40).dependent = false
	entry({"admin", "services", "routerpilot", "settings", "general"},   cbi("routerpilot/general"),   _("General"),   10).dependent = false
	entry({"admin", "services", "routerpilot", "settings", "telegram"},  cbi("routerpilot/telegram"),  _("Telegram"),  20).dependent = false
	entry({"admin", "services", "routerpilot", "settings", "llm"},       cbi("routerpilot/llm"),       _("LLM"),       30).dependent = false
	entry({"admin", "services", "routerpilot", "settings", "schedule"},  cbi("routerpilot/schedule"),  _("Schedule"),  40).dependent = false
end

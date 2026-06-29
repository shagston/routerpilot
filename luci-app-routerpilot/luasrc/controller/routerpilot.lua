module("luci.controller.routerpilot", package.seeall)

function index()
	entry({"admin", "services", "routerpilot"}, alias("admin", "services", "routerpilot", "dashboard"), _("RouterPilot"), 60)

	entry({"admin", "services", "routerpilot", "dashboard"}, template("routerpilot_dashboard"), _("Dashboard"), 10)
	entry({"admin", "services", "routerpilot", "execute"},   template("routerpilot_execute"),   _("Execute"),     20)
	entry({"admin", "services", "routerpilot", "chat"},       template("routerpilot_chat"),       _("Chat"),        30)
	entry({"admin", "services", "routerpilot", "settings"},   cbi("routerpilot"),                _("Settings"),    40)
end
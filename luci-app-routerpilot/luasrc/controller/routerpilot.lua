module("luci.controller.routerpilot", package.seeall)

function index()
	entry({"admin", "services", "routerpilot"}, alias("admin", "services", "routerpilot", "dashboard"), _("RouterPilot"), 60).dependent = false
	entry({"admin", "services", "routerpilot", "dashboard"}, template("routerpilot_dashboard"), _("Dashboard"), 10)
	entry({"admin", "services", "routerpilot", "console"}, template("routerpilot_console"), _("Console"), 20)
	entry({"admin", "services", "routerpilot", "config"}, cbi("routerpilot"), _("Settings"), 30)
end

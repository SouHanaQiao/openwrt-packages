module("luci.controller.AdGuardHome", package.seeall)

local fs = require "nixio.fs"
local http = require "luci.http"
local sys = require "luci.sys"

function index()
	local page = entry({"admin", "services", "AdGuardHome"},
		alias("admin", "services", "AdGuardHome", "base"),
		_("AdGuard Home"), 10)
	page.dependent = true
	page.acl_depends = { "luci-app-adguardhome-advanced" }

	entry({"admin", "services", "AdGuardHome", "base"},
		cbi("AdGuardHome/base"), _("Base Setting"), 1).leaf = true
	entry({"admin", "services", "AdGuardHome", "log"},
		form("AdGuardHome/log"), _("Log"), 2).leaf = true
	entry({"admin", "services", "AdGuardHome", "manual"},
		cbi("AdGuardHome/manual"), _("Manual Config"), 3).leaf = true
	entry({"admin", "services", "AdGuardHome", "status"},
		call("act_status")).leaf = true
	entry({"admin", "services", "AdGuardHome", "reloadconfig"},
		call("reload_config")).leaf = true
end

function act_status()
	local result = {
		running = (sys.call("/etc/init.d/adguardhome status >/dev/null 2>&1") == 0),
		official = true
	}
	http.prepare_content("application/json")
	http.write_json(result)
end

function reload_config()
	fs.remove("/tmp/AdGuardHometmpconfig.yaml")
	sys.call("/etc/init.d/adguardhome restart >/dev/null 2>&1")
	http.prepare_content("application/json")
	http.write_json({ ok = true })
end

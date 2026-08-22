local fs = require "nixio.fs"
local sys = require "luci.sys"
local uci = require "luci.model.uci".cursor()

local m = Map("AdGuardHome", "AdGuard Home")
m.description = translate("Feature-rich frontend for OpenWrt's official AdGuard Home package. The core binary is managed only by apk.")
m:section(SimpleSection).template = "AdGuardHome/AdGuardHome_status"

local s = m:section(TypedSection, "AdGuardHome")
s.anonymous = true
s.addremove = false

local o = s:option(Flag, "enabled", translate("Enable"))
o.default = 1
o.rmempty = false

o = s:option(DummyValue, "version", translate("Official core version"))
o.cfgvalue = function()
	local value = sys.exec("/usr/bin/AdGuardHome --version 2>/dev/null") or ""
	return value:gsub("%s+$", "")
end
o.description = translate("The binary comes from OpenWrt's official adguardhome package. Upgrade it through Software or apk, not from this page.")

local httpport = uci:get("AdGuardHome", "AdGuardHome", "httpport") or "3000"
o = s:option(Value, "httpport", translate("Browser management port"))
o.default = "3000"
o.datatype = "port"
o.rmempty = false
o.description = translate("<input type=\"button\" class=\"cbi-button cbi-button-apply\" value=\"Open AdGuard Home Web\" onclick=\"window.open('http://'+window.location.hostname+':" .. httpport .. "/')\"/>")

o = s:option(Value, "configpath", translate("Config Path"))
o.default = "/etc/adguardhome/adguardhome.yaml"
o.rmempty = false
o.description = translate("Official configuration file used by /etc/init.d/adguardhome.")
o.validate = function(self, value)
	if not value or value == "" or value == "/etc" then
		return nil, translate("Configuration file must be stored in its own directory")
	end
	return value
end

o = s:option(Value, "workdir", translate("Work dir"))
o.default = "/var/lib/adguardhome"
o.rmempty = false
o.description = translate("Official working directory containing filters, logs and statistics.")

o = s:option(ListValue, "logfile", translate("Runtime log"))
o:value("syslog", translate("System log"))
o:value("", translate("Disabled"))
o.default = "syslog"
o.rmempty = true

o = s:option(Flag, "verbose", translate("Verbose log"))
o.default = 0
o.rmempty = false

o = s:option(DummyValue, "service_user", translate("Service account"))
o.cfgvalue = function()
	return "adguardhome:adguardhome"
end

o = s:option(DummyValue, "dns_chain", translate("DNS chain"))
o.cfgvalue = function()
	return "AdGuard Home :53 → SmartDNS :6053; dnsmasq DHCP/DNS helper :1745"
end
o.description = translate("This frontend does not alter firewall rules, DNS redirection, or the official binary.")

function m.on_commit()
	local enabled = uci:get("AdGuardHome", "AdGuardHome", "enabled") or "1"
	local configpath = uci:get("AdGuardHome", "AdGuardHome", "configpath") or "/etc/adguardhome/adguardhome.yaml"
	local workdir = uci:get("AdGuardHome", "AdGuardHome", "workdir") or "/var/lib/adguardhome"
	local verbose = uci:get("AdGuardHome", "AdGuardHome", "verbose") or "0"

	uci:set("adguardhome", "config", "config_file", configpath)
	uci:set("adguardhome", "config", "work_dir", workdir)
	uci:set("adguardhome", "config", "user", "adguardhome")
	uci:set("adguardhome", "config", "group", "adguardhome")
	uci:set("adguardhome", "config", "verbose", verbose)
	uci:commit("adguardhome")

	if enabled == "1" then
		sys.call("/etc/init.d/adguardhome enable >/dev/null 2>&1")
		sys.call("/etc/init.d/adguardhome restart >/dev/null 2>&1 &")
	else
		sys.call("/etc/init.d/adguardhome stop >/dev/null 2>&1")
		sys.call("/etc/init.d/adguardhome disable >/dev/null 2>&1")
	end
end

return m

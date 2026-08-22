local fs = require "nixio.fs"
local sys = require "luci.sys"
local uci = require "luci.model.uci".cursor()

local configpath = uci:get("AdGuardHome", "AdGuardHome", "configpath")
	 or "/etc/adguardhome/adguardhome.yaml"
local binpath = "/usr/bin/AdGuardHome"

local m = Map("AdGuardHome", translate("Manual Config"))
m.description = translate("Edit the YAML used by OpenWrt's official AdGuard Home service. Configuration is validated before it is installed.")

local s = m:section(TypedSection, "AdGuardHome")
s.anonymous = true
s.addremove = false

local o = s:option(TextValue, "escconf")
o.rows = 66
o.wrap = "off"
o.rmempty = false
o.cfgvalue = function()
	return fs.readfile("/tmp/AdGuardHometmpconfig.yaml")
		or fs.readfile(configpath)
		or ""
end
o.validate = function(self, value)
	fs.writefile("/tmp/AdGuardHometmpconfig.yaml", value:gsub("\r\n", "\n"))
	local command = string.format("%s --config %q --check-config >/tmp/AdGuardHometest.log 2>&1", binpath, "/tmp/AdGuardHometmpconfig.yaml")
	if sys.call(command) == 0 then
		return value
	end
	return nil, fs.readfile("/tmp/AdGuardHometest.log") or translate("Configuration validation failed")
end
o.write = function(self, section, value)
	fs.writefile(configpath, value:gsub("\r\n", "\n"))
	sys.call("chown adguardhome:adguardhome " .. string.format("%q", configpath))
	sys.call("chmod 600 " .. string.format("%q", configpath))
	fs.remove("/tmp/AdGuardHometmpconfig.yaml")
end
o.template = "AdGuardHome/yamleditor"

function m.on_commit()
	sys.call("/etc/init.d/adguardhome restart >/dev/null 2>&1 &")
end

return m

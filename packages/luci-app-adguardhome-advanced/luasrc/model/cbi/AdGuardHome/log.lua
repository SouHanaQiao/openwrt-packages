local sys = require "luci.sys"

local f = SimpleForm("logview", translate("AdGuard Home Log"))
f.reset = false
f.submit = false

local t = f:field(TextValue, "conf")
t.rmempty = true
t.rows = 30
t.readonly = "readonly"
t.cfgvalue = function()
	return sys.exec("logread -e AdGuardHome | tail -n 500") or ""
end

return f

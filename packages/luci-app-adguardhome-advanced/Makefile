# SPDX-License-Identifier: Apache-2.0
#
# Advanced LuCI frontend adapted from rufengsuixing/luci-app-adguardhome.
# It deliberately uses OpenWrt's official adguardhome package, init script,
# UCI configuration and binary. It does not download or replace the core.

include $(TOPDIR)/rules.mk

PKG_NAME:=luci-app-adguardhome-advanced
PKG_RELEASE:=1
PKG_MAINTAINER:=SouHanaQiao

LUCI_TITLE:=Advanced LuCI frontend for official AdGuard Home
LUCI_PKGARCH:=all
LUCI_DEPENDS:=+adguardhome +luci-compat
LUCI_DESCRIPTION:=Feature-rich LuCI UI using OpenWrt's official AdGuard Home service and binary

define Package/$(PKG_NAME)/conffiles
/etc/config/AdGuardHome
endef

define Package/$(PKG_NAME)/postinst
#!/bin/sh
if [ -z "$${IPKG_INSTROOT}" ]; then
	uci -q set AdGuardHome.AdGuardHome=AdGuardHome
	uci -q set AdGuardHome.AdGuardHome.enabled='1'
	uci -q set AdGuardHome.AdGuardHome.httpport='3000'
	uci -q set AdGuardHome.AdGuardHome.configpath='/etc/adguardhome/adguardhome.yaml'
	uci -q set AdGuardHome.AdGuardHome.workdir='/var/lib/adguardhome'
	uci -q set AdGuardHome.AdGuardHome.logfile='syslog'
	uci -q set AdGuardHome.AdGuardHome.binpath='/usr/bin/AdGuardHome'
	uci -q commit AdGuardHome
	/etc/init.d/adguardhome enable >/dev/null 2>&1
	/etc/init.d/adguardhome start >/dev/null 2>&1
	rm -f /tmp/luci-indexcache
	rm -rf /tmp/luci-modulecache 2>/dev/null
fi
exit 0
endef

define Package/$(PKG_NAME)/prerm
#!/bin/sh
if [ -z "$${IPKG_INSTROOT}" ]; then
	rm -f /tmp/luci-indexcache
	rm -rf /tmp/luci-modulecache 2>/dev/null
fi
exit 0
endef

include $(TOPDIR)/feeds/luci/luci.mk

# call BuildPackage - OpenWrt buildroot signature

include $(TOPDIR)/rules.mk

PKG_NAME:=clash_subscriber
PKG_VERSION:=1.0
PKG_RELEASE:=1

PKG_LICENSE:=MIT
PKG_MAINTAINER:=pacey <wangqinqian@126.com>

include $(INCLUDE_DIR)/package.mk

PKG_BUILD_DIR := $(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)

define Package/clash_subscriber
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=Simple Go HTTP Server for OpenWrt
  DEPENDS:=
  ARCH:=aarch64
endef

define Package/clash_subscriber/description
  A simple Go HTTP server written in Go.
endef

define Build/Prepare
	mkdir -p $(PKG_BUILD_DIR)
	cp -r $(CURDIR)/src/* $(PKG_BUILD_DIR)/
endef

define Build/Compile
	(cd $(PKG_BUILD_DIR); \
	GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -o clash_subscriber main.go)
endef

define Package/clash_subscriber/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/clash_subscriber $(1)/usr/bin/

	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) $(CURDIR)/files/etc/init.d/clash_subscriber $(1)/etc/init.d/clash_subscriber
endef

$(eval $(call BuildPackage,clash_subscriber))
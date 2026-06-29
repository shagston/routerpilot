#
# Copyright (C) 2024 shagston
#
# This is free software, licensed under the MIT License.
#

include $(TOPDIR)/rules.mk

PKG_NAME:=routerpilot
PKG_VERSION:=0.1.0
PKG_RELEASE:=1

PKG_SOURCE_PROTO:=git
PKG_SOURCE_URL:=https://github.com/shagston/routerpilot.git
PKG_SOURCE_VERSION:=HEAD
PKG_SOURCE_SUBDIR:=$(PKG_NAME)-$(PKG_VERSION)

PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_USE_MIPS16:=0

GO_PKG:=github.com/shagston/routerpilot
GO_PKG_BUILD_PKG:=github.com/shagston/routerpilot/cmd/routerpilot
GO_PKG_LDFLAGS:=-s -w

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

define Package/routerpilot
	SECTION:=net
	CATEGORY:=Network
	TITLE:=AI-native network automation daemon
	URL:=https://github.com/shagston/routerpilot
	MAINTAINER:=shagston
	DEPENDS:=$(GO_ARCH_DEPENDS)
	MENU:=1
endef

define Package/routerpilot/description
  RouterPilot is a real-time network diagnostics and automation daemon
  for OpenWrt. It provides a REST API, Web UI, WebSocket console, and
  Telegram bot interface for network management, with offline KB-based
  diagnosis when no LLM is available.
endef

define Package/routerpilot/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(GO_PKG_BUILD_DIR)/bin/routerpilot $(1)/usr/bin/
endef

$(eval $(call GoBinPackage,routerpilot))
$(eval $(call BuildPackage,routerpilot))

#
# Copyright (C) 2024 shagston
#
# This is free software, licensed under the MIT License.
#

include $(TOPDIR)/rules.mk

LUCI_TITLE:=RouterPilot - AI-native network automation daemon
LUCI_DESCRIPTION:=Real-time network diagnostics, automation, and offline troubleshooting for OpenWrt
LUCI_DEPENDS:=+routerpilot +luci-base
LUCI_PKGARCH:=all

PKG_NAME:=luci-app-routerpilot
PKG_VERSION:=0.1.0
PKG_RELEASE:=1

PKG_MAINTAINER:=shagston

include $(TOPDIR)/feeds/luci/luci.mk

# call BuildPackage - OpenWrt buildroot integration

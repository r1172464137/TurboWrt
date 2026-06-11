include $(TOPDIR)/rules.mk

PKG_NAME:=devman
PKG_VERSION:=0.1.0
PKG_RELEASE:=1

PKG_SOURCE_PROTO:=local
PKG_SOURCE_DIR:=src

include $(INCLUDE_DIR)/package.mk

define Package/devman
	SECTION:=utils
	CATEGORY:=Utilities
	TITLE:=Device Manager Daemon
	DEPENDS:=+libsqlite3
	URL:=https://github.com/r1172464137/TurboWrt
endef

define Package/devman/description
	Device profile manager - DHCP fingerprint, conntrack, nftables
endef

define Build/Compile
	cd $(PKG_BUILD_DIR) && \
	CC=$(TARGET_CC) \
	CGO_ENABLED=1 \
	CGO_CFLAGS="$(TARGET_CFLAGS)" \
	CGO_LDFLAGS="$(TARGET_LDFLAGS)" \
	GOOS=linux \
	GOARCH=amd64 \
	go build -ldflags="-s -w" -o devman .
endef

define Package/devman/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/devman $(1)/usr/bin/devman
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/devman.init $(1)/etc/init.d/devman
	$(INSTALL_DIR) $(1)/var/lib
endef

$(eval $(call BuildPackage,devman))

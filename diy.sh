#!/bin/bash
# DIY script - runs after feeds update and before defconfig

# === Batch 1: mitigations=off ===
sed -i 's/@CMDLINE@ noinitrd/@CMDLINE@ mitigations=off noinitrd/g' target/linux/x86/image/grub-efi.cfg
sed -i 's/@CMDLINE@ noinitrd/@CMDLINE@ mitigations=off noinitrd/g' target/linux/x86/image/grub-pc.cfg
sed -i 's/@CMDLINE@ noinitrd/@CMDLINE@ mitigations=off noinitrd/g' target/linux/x86/image/grub-iso.cfg

# === Batch 2: Nginx HTTP only (remove SSL) ===
cat > feeds/packages/net/nginx-util/files/nginx.config << 'EOF'

config main global
	option uci_enable 'true'

config server '_lan'
	list listen '80 default_server'
	list listen '[::]:80 default_server'
	option server_name '_lan'
	list include 'restrict_locally'
	list include 'conf.d/*.locations'
	option access_log 'off; # logd openwrt'
EOF

# === Batch 3: Nginx tuning ===
sed -i 's/client_max_body_size 128M/client_max_body_size 2048M/' feeds/packages/net/nginx-util/files/uci.conf.template
sed -i '/client_max_body_size/a\\tclient_body_buffer_size 8192M;' feeds/packages/net/nginx-util/files/uci.conf.template
sed -i '/client_max_body_size/a\\tserver_names_hash_bucket_size 128;' feeds/packages/net/nginx-util/files/uci.conf.template
sed -i 's/large_client_header_buffers 2 1k/large_client_header_buffers 4 32k/' feeds/packages/net/nginx-util/files/uci.conf.template
sed -i '/ubus_parallel_req/a\\tubus_script_timeout 600;' feeds/packages/net/nginx/files-luci-support/60_nginx-luci-support

# === Batch 4: uwsgi ===
sed -i 's/buffer-size = 10000/buffer-size = 131072/g' feeds/packages/net/uwsgi/files-luci-support/luci-webui.ini
sed -i 's/buffer-size = 10000/buffer-size = 131072/g' feeds/packages/net/uwsgi/files-luci-support/luci-cgi_io.ini
sed -i 's/threads = 1/threads = 2/' feeds/packages/net/uwsgi/files-luci-support/luci-webui.ini
sed -i 's/processes = 3/processes = 4/' feeds/packages/net/uwsgi/files-luci-support/luci-webui.ini
sed -i 's/cheaper = 1/cheaper = 2/' feeds/packages/net/uwsgi/files-luci-support/luci-webui.ini
sed -i 's/logger = luci syslog:uwsgi-luci/#logger = luci syslog:uwsgi-luci/' feeds/packages/net/uwsgi/files-luci-support/luci-webui.ini
sed -i 's/idle = 360/cgi-timeout = 600\nidle = 360/' feeds/packages/net/uwsgi/files-luci-support/luci-webui.ini
sed -i 's/idle = 360/cgi-timeout = 600\nidle = 360/' feeds/packages/net/uwsgi/files-luci-support/luci-cgi_io.ini
sed -i 's/procd_set_param stderr 1/procd_set_param stderr 0/' feeds/packages/net/uwsgi/files/uwsgi.init

# === Batch 5: rpcd ===
sed -i 's/option timeout 30/option timeout 60/' package/system/rpcd/files/rpcd.config

# === Batch 6: rpc.js ===
sed -i 's/rpctimeout \?\? 20/rpctimeout ?? 60/' feeds/luci/modules/luci-base/htdocs/luci-static/resources/rpc.js

# === Batch 7: ttyd no login + disable CPD ===
sed -i 's/command .\/bin\/login/command \/bin\/sh/' feeds/packages/utils/ttyd/files/ttyd.config
# Fix Docker package versions (APK doesn't allow "v" prefix)
sed -i 's/PKG_VERSION:=v0\.3\.4/PKG_VERSION:=0.3.4/' feeds/luci/collections/luci-lib-docker/Makefile 2>/dev/null || true
sed -i 's/PKG_VERSION:=v0\.5\.26/PKG_VERSION:=0.5.26/' feeds/luci/applications/luci-app-dockerman/Makefile 2>/dev/null || true
# Disable cgroupfs-mount (missing Makefile in feeds)
echo '# CONFIG_PACKAGE_cgroupfs-mount is not set' >> .config

# === Batch 8: Firewall offload UI removal ===
sed -i '/Netfilter flow offload/,/};/c\		/* Netfilter flow offload - disabled, handled by TurboACC */' feeds/luci/applications/luci-app-firewall/htdocs/luci-static/resources/view/firewall/zones.js

# === Batch 9: ImmortalWrt APK repo ===
mkdir -p package/base-files/files/etc/apk/keys
mkdir -p package/base-files/files/etc/apk/repositories.d
cat > package/base-files/files/etc/apk/keys/immortalwrt-snapshots.pem << 'EOF'
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE4gEuRXAk4+ArZvB5VKgE/o9zrxif
a+38TgvDUZuoAp4x2rya/uVUJvSCkTHxR5gjqstI9W60dRnWErT44Uu0nw==
-----END PUBLIC KEY-----
EOF
cat > package/base-files/files/etc/apk/repositories.d/immortalwrt.list << 'EOF'
https://downloads.immortalwrt.org/snapshots/targets/x86/64/packages/packages.adb
https://downloads.immortalwrt.org/snapshots/packages/x86_64/luci/packages.adb
https://downloads.immortalwrt.org/snapshots/packages/x86_64/base/packages.adb
EOF
cat > package/base-files/files/etc/apk/keys/openwrt-snapshots.pem << 'EOF'
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEqDM0+yparYvbHosRPBhvT5Z3MEXz
AFYrTnqJrnURywsKpD+ZKCLjPluvoHe/FABIvIuHLvICALA3IMjhm0Z0cA==
-----END PUBLIC KEY-----
EOF

# === Batch 10: uci-defaults for nginx HTTP ===
cat > package/base-files/files/etc/uci-defaults/99-disable-https-redirect << 'EOF'
#!/bin/sh
uci delete nginx._redirect2ssl 2>/dev/null
uci set nginx._lan.listen='80 default_server'
uci add_list nginx._lan.listen='[::]:80 default_server'
uci commit nginx
/etc/init.d/nginx restart 2>/dev/null
exit 0
EOF

# === Batch 11: conntrack ===
echo "net.netfilter.nf_conntrack_tcp_max_retrans=5" >> package/kernel/linux/files/sysctl-nf-conntrack.conf

# === Batch 12: DHCPv6 hotplug ===
patch -p1 < $GITHUB_WORKSPACE/patches/odhcp6c/1002-odhcp6c-support-dhcpv6-hotplug.patch

echo "diy.sh completed"

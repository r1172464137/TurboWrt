#!/bin/bash
# DIY script - runs after feeds update and before defconfig

# === Batch 1: mitigations=off ===
sed -i 's/@CMDLINE@ noinitrd/@CMDLINE@ mitigations=off noinitrd/g' target/linux/x86/image/grub-efi.cfg
sed -i 's/@CMDLINE@ noinitrd/@CMDLINE@ mitigations=off noinitrd/g' target/linux/x86/image/grub-pc.cfg
sed -i 's/@CMDLINE@ noinitrd/@CMDLINE@ mitigations=off noinitrd/g' target/linux/x86/image/grub-iso.cfg

# === Batch 2: TEO for x86/64 ===
cat >> target/linux/x86/64/config-6.12 << EOF
# CPU idle governor TEO
CONFIG_CPU_IDLE_GOV_MENU=n
CONFIG_CPU_IDLE_GOV_TEO=y
EOF

# === Batch 3: conntrack ===
echo "net.netfilter.nf_conntrack_tcp_max_retrans=5" >> package/kernel/linux/files/sysctl-nf-conntrack.conf

# === Batch 4: Nginx ===
sed -i 's/client_max_body_size 128M/client_max_body_size 2048M/' feeds/packages/net/nginx-util/files/uci.conf.template
sed -i '/client_max_body_size/a\\tclient_body_buffer_size 8192M;' feeds/packages/net/nginx-util/files/uci.conf.template
sed -i '/client_max_body_size/a\\tserver_names_hash_bucket_size 128;' feeds/packages/net/nginx-util/files/uci.conf.template
sed -i 's/large_client_header_buffers 2 1k/large_client_header_buffers 4 32k/' feeds/packages/net/nginx-util/files/uci.conf.template
sed -i '/ubus_parallel_req/a\\tubus_script_timeout 600;' feeds/packages/net/nginx/files-luci-support/60_nginx-luci-support

# === Batch 5: uwsgi ===
sed -i 's/buffer-size = 10000/buffer-size = 131072/g' feeds/packages/net/uwsgi/files-luci-support/luci-webui.ini
sed -i 's/buffer-size = 10000/buffer-size = 131072/g' feeds/packages/net/uwsgi/files-luci-support/luci-cgi_io.ini
sed -i 's/threads = 1/threads = 2/' feeds/packages/net/uwsgi/files-luci-support/luci-webui.ini
sed -i 's/processes = 3/processes = 4/' feeds/packages/net/uwsgi/files-luci-support/luci-webui.ini
sed -i 's/cheaper = 1/cheaper = 2/' feeds/packages/net/uwsgi/files-luci-support/luci-webui.ini
sed -i 's/logger = luci syslog:uwsgi-luci/#logger = luci syslog:uwsgi-luci/' feeds/packages/net/uwsgi/files-luci-support/luci-webui.ini
sed -i 's/idle = 360/cgi-timeout = 600\nidle = 360/' feeds/packages/net/uwsgi/files-luci-support/luci-webui.ini
sed -i 's/idle = 360/cgi-timeout = 600\nidle = 360/' feeds/packages/net/uwsgi/files-luci-support/luci-cgi_io.ini
sed -i 's/procd_set_param stderr 1/procd_set_param stderr 0/' feeds/packages/net/uwsgi/files/uwsgi.init

# === Batch 6: rpcd ===
sed -i 's/option timeout 30/option timeout 60/' package/system/rpcd/files/rpcd.config

# === Batch 7: rpc.js ===
sed -i 's/rpctimeout \?\? 20/rpctimeout ?? 60/' feeds/luci/modules/luci-base/htdocs/luci-static/resources/rpc.js

echo "diy.sh completed"

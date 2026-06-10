#!/bin/bash
# Setup script for CI build - clones packages, applies patches, configures
set -e

OPENWRT_DIR="$1"
WORKSPACE="$2"

cd "$OPENWRT_DIR"

echo "=== Cloning third-party packages ==="
git clone --depth 1 https://github.com/gdy666/luci-app-lucky.git package/new/luci-app-lucky
git clone --depth 1 https://github.com/sbwml/luci-app-openlist2.git package/new/luci-app-openlist2
git clone --depth 1 https://github.com/sbwml/luci-app-mosdns.git package/new/luci-app-mosdns
git clone --depth 1 https://github.com/sbwml/openwrt_helloworld.git package/new/openwrt_helloworld
git clone --depth 1 https://github.com/sbwml/luci-app-quickfile.git package/new/luci-app-quickfile
git clone --depth 1 https://github.com/sirpdboy/luci-app-ddns-go.git package/new/luci-app-ddns-go
git clone --depth 1 https://github.com/UnblockNeteaseMusic/luci-app-unblockneteasemusic.git package/new/luci-app-unblockneteasemusic
git clone --depth 1 https://github.com/EasyTier/luci-app-easytier.git package/new/luci-app-easytier
git clone --depth 1 https://github.com/timsaya/luci-app-bandix-plus.git package/new/luci-app-bandix-plus
git clone --depth 1 https://github.com/timsaya/openwrt-bandix-plus.git package/new/openwrt-bandix-plus
git clone --depth 1 https://github.com/jerrykuku/luci-theme-argon.git package/new/luci-theme-argon
git clone --depth 1 https://github.com/eamonxg/luci-theme-aurora.git package/new/luci-theme-aurora
git clone --depth 1 https://github.com/eamonxg/luci-theme-shadcn.git package/new/luci-theme-shadcn
git clone --depth 1 https://github.com/eamonxg/luci-app-aurora-config.git package/new/luci-app-aurora-config

echo "=== Fixing package structures ==="
fix_nested() {
  local dir="$1" sub="$2"
  [ -d "$dir/$sub" ] && cp -rf "$dir/$sub/"* "$dir/" && rm -rf "$dir/$sub"
}
fix_nested "package/new/luci-app-bandix-plus" "luci-app-bandix-plus"
fix_nested "package/new/openwrt-bandix-plus" "openwrt-bandix-plus"
fix_nested "package/new/luci-app-lucky" "luci-app-lucky"
fix_nested "package/new/luci-app-easytier" "luci-app-easytier"
fix_nested "package/new/luci-app-ddns-go" "luci-app-ddns-go"

echo "=== Creating helloworld symlinks ==="
cd package/new
for sub in luci-app-homeproxy luci-app-nikki luci-app-passwall luci-app-dae luci-app-openclash luci-app-ssr-plus; do
  [ -d "openwrt_helloworld/$sub" ] && ln -sf "openwrt_helloworld/$sub" "$sub" 2>/dev/null || true
done
cd "$OPENWRT_DIR"

# cachewrtbuild handles toolchain caching + skip; remove this block

echo "=== Applying config (merge with toolchain settings) ==="
if [ -f .config ]; then
  cat "$WORKSPACE/config.seed" >> .config
else
  cp "$WORKSPACE/config.seed" .config
fi

echo "=== Applying kernel patches ==="
cp -rf "$WORKSPACE/patches/kernel/bbr3/"* target/linux/generic/backport-6.12/
cp -rf "$WORKSPACE/patches/kernel/other/"* target/linux/generic/
cat "$WORKSPACE/patches/config-append.txt" >> target/linux/generic/config-6.12

echo "=== Running diy.sh ==="
GITHUB_WORKSPACE="$WORKSPACE" bash "$WORKSPACE/diy.sh"

echo "=== Adding turboacc (no SFE) ==="
curl -sSL https://raw.githubusercontent.com/chenmozhijin/turboacc/luci/add_turboacc.sh -o add_turboacc.sh
bash add_turboacc.sh --no-sfe

rm -f target/linux/generic/hack-6.12/952-add-net-conntrack-events-support-multiple-registrant.patch

echo "=== Running defconfig ==="
make defconfig

# Compile host tools AFTER final config (prevent config hash mismatch in make world)
echo "=== Compiling host tools ==="
make tools/compile -j$(nproc)

echo "=== Setup complete ==="

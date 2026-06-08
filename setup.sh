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
git clone --depth 1 https://github.com/timsaya/luci-app-bandix.git package/new/luci-app-bandix
git clone --depth 1 https://github.com/timsaya/openwrt-bandix.git package/new/openwrt-bandix
git clone --depth 1 https://github.com/jerrykuku/luci-theme-argon.git package/new/luci-theme-argon
git clone --depth 1 https://github.com/eamonxg/luci-theme-aurora.git package/new/luci-theme-aurora
git clone --depth 1 https://github.com/eamonxg/luci-theme-shadcn.git package/new/luci-theme-shadcn
git clone --depth 1 https://github.com/eamonxg/luci-app-aurora-config.git package/new/luci-app-aurora-config

echo "=== Fixing package structures ==="
fix_nested() {
  local dir="$1" sub="$2"
  [ -d "$dir/$sub" ] && cp -rf "$dir/$sub/"* "$dir/" && rm -rf "$dir/$sub"
}
fix_nested "package/new/luci-app-bandix" "luci-app-bandix"
fix_nested "package/new/openwrt-bandix" "openwrt-bandix"
fix_nested "package/new/luci-app-lucky" "luci-app-lucky"
fix_nested "package/new/luci-app-easytier" "luci-app-easytier"
fix_nested "package/new/luci-app-ddns-go" "luci-app-ddns-go"

echo "=== Creating helloworld symlinks ==="
cd package/new
for sub in luci-app-homeproxy luci-app-nikki luci-app-passwall luci-app-dae luci-app-openclash luci-app-ssr-plus; do
  [ -d "openwrt_helloworld/$sub" ] && ln -sf "openwrt_helloworld/$sub" "$sub" 2>/dev/null || true
done
cd "$OPENWRT_DIR"

echo "=== Applying config ==="
cp "$WORKSPACE/config.seed" .config

echo "=== Applying kernel patches ==="
cp -rf "$WORKSPACE/patches/kernel/bbr3/"* target/linux/generic/backport-6.12/
cp -rf "$WORKSPACE/patches/kernel/other/"* target/linux/generic/
cat "$WORKSPACE/patches/config-append.txt" >> target/linux/generic/config-6.12

echo "=== Running diy.sh ==="
GITHUB_WORKSPACE="$WORKSPACE" bash "$WORKSPACE/diy.sh"

echo "=== Downloading prebuilt toolchain ==="
TOOLCHAIN_URL="https://downloads.openwrt.org/releases/25.12.4/targets/x86/64"
TOOLCHAIN_FILE="openwrt-toolchain-25.12.4-x86-64_gcc-14.3.0_musl.Linux-x86_64.tar.zst"
wget -q "$TOOLCHAIN_URL/$TOOLCHAIN_FILE" -O /tmp/toolchain.tar.zst
tar --zstd -xf /tmp/toolchain.tar.zst -C /tmp/
rm -f /tmp/toolchain.tar.zst
# Run ext-toolchain script
./scripts/ext-toolchain.sh --toolchain /tmp/openwrt-toolchain-*/toolchain-*/ --config x86/64

echo "=== Adding turboacc (no SFE) ==="
curl -sSL https://raw.githubusercontent.com/chenmozhijin/turboacc/luci/add_turboacc.sh -o add_turboacc.sh
bash add_turboacc.sh --no-sfe

# Remove 952 patch again - turboacc re-adds it but kernel 6.12.87 already has it
rm -f target/linux/generic/hack-6.12/952-add-net-conntrack-events-support-multiple-registrant.patch

echo "=== Running defconfig ==="
make defconfig

echo "=== Setup complete ==="

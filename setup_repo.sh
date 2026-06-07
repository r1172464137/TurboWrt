#!/bin/bash
set -e
cd /home/compile/openwrt-ci
git init -b main
git config user.email "builder@openwrt.local"
git config user.name "OpenWrt Builder"
git add -A
git commit -m "OpenWrt 25.12.4 CI build scripts

Includes:
- Build workflow with GitHub Actions
- Custom .config seed with all packages
- BBRv3 kernel patches (19)
- Kernel config tweaks (TEO, Netkit, mitigations, conntrack)
- Nginx/uwsgi/rpcd performance tuning
- TurboACC fullcone NAT support"
git remote add origin https://github.com/r1172464137/openwrt-custom.git
git push -u origin main --force
echo "PUSH DONE"

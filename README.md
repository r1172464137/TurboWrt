<p align="center">
<img width="768" src="https://raw.githubusercontent.com/QiuSimons/Others/master/YAOF.png" >
</p>

<h1 align="center">OpenWrt 25.12.4 定制固件</h1>
<p align="center">
基于官方 OpenWrt v25.12.4 编译，针对 x86/64 软路由优化
</p>

---

### 特性

- 基于 **官方 OpenWrt 25.12.4** 编译，默认管理地址 192.168.1.1
- **HTTP 默认**，无需 HTTPS 自签名证书
- 预配置了大量插件，可无脑 opkg/apk kmod
- 内置 **ImmortalWrt 软件源**，刷机后可直接 `apk add` 更多包
- **BBRv3** 默认启用
- **FullCone NAT**（防火墙中默认开启）
- **Flow Offloading** 流量分载
- **O2 编译**，CFLAG 优化
- 物理 Reset 按键可用

---

### 包含的插件

**加速相关：**
- `luci-app-turboacc` — Flow Offload / FullCone / BBR
- `kmod-nft-fullcone` — 全锥型 NAT
- `kmod-nft-offload` — 流量分载
- BBRv3 — 最新 TCP 拥塞控制算法

**代理相关：**
- `luci-app-passwall` — 全能代理
- `luci-app-homeproxy` — 基于 sing-box 的代理
- `luci-app-nikki` — 轻量代理
- `luci-app-dae` — eBPF 代理

**网络工具：**
- `luci-app-mosdns` — DNS 分流/广告过滤
- `luci-app-bandix` — 流量监控
- `luci-app-easytier` — 组网工具
- `luci-app-lucky` — 综合工具
- `luci-app-ddns-go` — DDNS
- `zerotier` — 虚拟组网
- `frpc` / `frps` — 内网穿透（已移除，可在 ImmortalWrt 源安装）

**系统工具：**
- `luci-app-ttyd` — Web 终端（免密直接进 shell）
- `luci-app-filemanager` — 文件管理
- `luci-app-watchcat` — 看门狗
- `luci-app-sqm` — QoS 流量整形
- `luci-app-upnp` — UPnP
- `luci-app-aurora-config` — Aurora 主题配置
- `coremark` / `htop` / `ethtool` / `jq` / `yq`

**主题：**
- `luci-theme-argon` — Argon 主题
- `luci-theme-aurora` — Aurora 主题
- `luci-theme-shadcn` — ShadCN 风格主题

---

### 系统优化

| 优化项 | 说明 |
|--------|------|
| `mitigations=off` | 关闭 CPU 安全缓解，提升性能 |
| TEO 空闲调度 | 替代 Menu governor，适合路由场景 |
| Netkit | 内核新网络框架 |
| BBRv3 | 19 个内核补丁，最新的 TCP 拥塞控制 |
| Nginx 调优 | 缓冲区翻倍，超时延长 |
| uwsgi 调优 | 线程/进程数增加，性能提升 |
| rpcd 超时 | 30s → 60s |
| conntrack 调优 | 减少 TCP 重传误判 |

### 内核补丁

- **BBRv3** — 19 个补丁，Google 最新 TCP 拥塞控制
- **IGC EEE 关断** — 解决 i225/i226 网卡断流
- **WireGuard hotplug** — WG 接口触发 hotplug 事件
- **TCP collapse 跳过** — Cloudflare 出品，减少高吞吐延迟抖动
- **BTF 静默** — 关掉 BTF 警告信息
- **DHCPv6 hotplug** — 获取 IPv6 时触发 hotplug 事件

---

### 使用方法

#### 前往 Release 页面下载

选择对应设备格式的固件，解压后写入磁盘（推荐 ext4-combined-efi）。

#### GitHub Actions 云编译

1. Fork/Clone 本仓库
2. 前往 Actions 页面
3. 点击 **Run workflow**
4. 等待编译完成（约 2-3 小时）
5. 下载生成的固件 Artifacts

#### 刷机后

```bash
# 更新源
apk update

# 从 ImmortalWrt 源安装额外包
apk add luci-app-openclash
apk add luci-app-ssr-plus
```

---

### 鸣谢

| 项目 | 说明 |
|------|------|
| [OpenWrt](https://github.com/openwrt/openwrt) | 官方 OpenWrt 源码 |
| [chenmozhijin/turboacc](https://github.com/chenmozhijin/turboacc) | FullCone NAT 加速 |
| [QiuSimons/YAOF](https://github.com/QiuSimons/YAOF) | 部分补丁来源 |
| [ImmortalWrt](https://github.com/immortalwrt) | 软件源及 APK 签名 |
| [sbwml](https://github.com/sbwml) | 各种软件包 |
| 各第三方插件作者 | |

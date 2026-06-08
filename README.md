<div align="center">

# ⚡ OpenWrt 25.12.4 定制固件

<p align="center">
  <b>基于官方 OpenWrt v25.12.4 · 专为 x86/64 软路由优化</b>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/OpenWrt-25.12.4-00a854?style=flat-square&logo=openwrt&logoColor=white"/>
  <img src="https://img.shields.io/badge/Kernel-6.12.87-3572A5?style=flat-square&logo=linux&logoColor=white"/>
  <img src="https://img.shields.io/badge/Arch-x86__64-0078D6?style=flat-square&logo=amd&logoColor=white"/>
  <img src="https://img.shields.io/badge/BBRv3-enabled-FF6F00?style=flat-square"/>
  <img src="https://img.shields.io/badge/FullCone-enabled-4CAF50?style=flat-square"/>
  <img src="https://img.shields.io/github/downloads/r1172464137/openwrt-custom-fw/total?style=flat-square"/>
</p>

</div>

---

## 📋 特性一览

<table>
<tr>
<td width="50%">

### 🚀 性能优化
- `mitigations=off` — 关闭 CPU 安全缓解
- **TEO 空闲调度** — 更低延迟
- **Netkit** — 内核新网络框架
- **BBRv3** — 最新 TCP 拥塞控制
- **FullCone NAT** — turboacc 管理
- **Flow Offloading** — 流量分载
- **conntrack 调优** — 减少重传误判

</td>
<td width="50%">

### 🧩 内核补丁
- 🔹 **BBRv3** × 19 补丁
- 🔹 **IGC EEE 关断** — i225/i226 防断流
- 🔹 **WireGuard hotplug**
- 🔹 **TCP collapse 跳过**
- 🔹 **BTF 静默**
- 🔹 **DHCPv6 hotplug**
- 🔹 **odhcpd/odhcp6c** 更新

</td>
</tr>
<tr>
<td>

### 🌐 网络增强
- **HTTP 默认** — 无需 HTTPS 证书
- **Nginx 调优** — 缓冲区翻倍
- **uwsgi 调优** — 线程/进程增加
- **rpcd 超时** — 30s → 60s
- **ImmortalWrt 源预置** — 刷机即用
- **迷你 UPnP** — 更新+补丁

</td>
<td>

### 🛠️ 系统优化
- **golang / Node.js** 最新版
- **fstool** — 一键格式化挂盘
- **自定义 nft 规则** — 防火墙增强
- **`fuck` 命令** — 一键重置

</td>
</tr>
</table>

---

## 📦 包含的插件

<details open>
<summary><b>点击展开/收起</b></summary>

### ⚡ 加速相关
| 插件 | 说明 |
|------|------|
| `luci-app-turboacc` | Flow Offload / FullCone / BBR 控制 |
| `kmod-nft-fullcone` | 全锥型 NAT |
| `kmod-nft-offload` | 流量分载 |
| BBRv3 | Google 最新 TCP 拥塞控制算法 |

### 🔒 代理相关
| 插件 | 说明 |
|------|------|
| `luci-app-passwall` | 全能代理 |
| `luci-app-homeproxy` | 基于 sing-box 的代理 |
| `luci-app-nikki` | 轻量代理 |
| `luci-app-dae` | eBPF 代理 |

### 🌍 网络工具
| 插件 | 说明 |
|------|------|
| `luci-app-mosdns` | DNS 分流 / 广告过滤 |
| `luci-app-bandix` | 网络流量监控 |
| `luci-app-easytier` | 组网工具 |
| `luci-app-lucky` | 综合工具 |
| `luci-app-ddns-go` | DDNS 动态域名 |
| `zerotier` | 虚拟组网 |
| `luci-app-sqm` | QoS 流量整形 |
| `luci-app-upnp` | UPnP 端口映射 |

### 🖥️ 系统工具
| 插件 | 说明 |
|------|------|
| `luci-app-ttyd` | Web 终端（免密） |
| `luci-app-filemanager` | 文件管理 |
| `luci-app-watchcat` | 系统看门狗 |
| `luci-app-aurora-config` | Aurora 主题配置 |
| `coremark` / `htop` / `ethtool` | 系统工具 |

### 🎨 主题
| 主题 | 说明 |
|------|------|
| `luci-theme-argon` | 最流行的 Argon 主题 |
| `luci-theme-aurora` | Aurora 主题 |
| `luci-theme-shadcn` | ShadCN 风格主题 |

</details>

---

## 🚀 使用方法

### 下载固件

前往 [Releases](https://github.com/r1172464137/openwrt-custom-fw/releases) 页面下载。

推荐使用 **`ext4-combined-efi.img.gz`**（UEFI 引导）。

### 刷机

```bash
# 解压
gunzip openwrt-x86-64-generic-ext4-combined-efi.img.gz

# 写入磁盘（替换 /dev/sdX 为你的磁盘）
dd if=openwrt-x86-64-generic-ext4-combined-efi.img of=/dev/sdX bs=4M status=progress
```

### GitHub Actions 云编译

1. 前往 [Actions](https://github.com/r1172464137/openwrt-custom-fw/actions) 页面
2. 点击 **Run workflow**
3. 等待编译完成（约 2-3 小时）
4. 下载生成的固件 Artifacts

### 刷机后

```bash
# 从 ImmortalWrt 源安装更多包
apk update
apk add luci-app-openclash
apk add luci-app-ssr-plus
```

---

## 📄 鸣谢

<p align="center">
  <a href="https://github.com/openwrt/openwrt"><img src="https://img.shields.io/badge/OpenWrt-官方源码-00a854?style=flat-square&logo=openwrt"/></a>
  <a href="https://github.com/chenmozhijin/turboacc"><img src="https://img.shields.io/badge/turboacc-FullCone_NAT-FF6F00?style=flat-square"/></a>
  <a href="https://github.com/QiuSimons/YAOF"><img src="https://img.shields.io/badge/YAOF-补丁参考-1a73e8?style=flat-square"/></a>
  <a href="https://github.com/immortalwrt"><img src="https://img.shields.io/badge/ImmortalWrt-软件源-e65100?style=flat-square"/></a>
  <a href="https://github.com/sbwml"><img src="https://img.shields.io/badge/sbwml-软件包-6f42c1?style=flat-square"/></a>
  <a href="https://github.com/wukongdaily/img-installer"><img src="https://img.shields.io/badge/img--installer-ISO生成-4FC08D?style=flat-square"/></a>
</p>

# Package provenance

此仓库聚合多个上游项目的 OpenWrt 打包文件。各目录继续遵循其上游许可证、版权声明和商标规则；本仓库不对第三方目录重新授权。

## 历史第三方包清单

Ubuntu 上旧 OpenWrt 24.10 构建的
`bin/packages/aarch64_generic/kiddin9/Packages` 索引包含 112 个二进制包。
这些二进制包已全部反向映射到 77 个 `kiddin9/kwrt-packages` 源码目录，
映射覆盖率为 112/112。源码快照提交：
`ba60167deab5c88bb5f77d39ed61a5c549f95e98`。

导入范围包括网络工具、存储工具、系统管理、主题和 LuCI 管理页，例如：

- SmartDNS、AdGuardHome、Netdata
- Clash、OpenClash、Nikki
- LinkEase、Lucky、Rclone、FileBrowser
- DiskMan、NFS、ZeroTier、Tailscale
- 网络测速、微信推送、定时任务、软件包同步
- Alpha、Argon、Design、Edge、iNAS、Kucat 等主题

`packages/` 中保留各上游目录原有许可证和来源信息，不导入上游仓库的
`.git` 元数据。

## `main` / OpenWrt 25.12

| 目录 | 说明 | 来源 |
| --- | --- | --- |
| `ClashSubscriber` | 本地订阅转换脚本 | 原 24.10/25.12 构建树 |
| `luci-app-nikki`、`nikki` | Nikki/Mihomo | 原 25.12 构建树，含 25.12 构建兼容修复 |
| `smartdns`、`luci-app-smartdns` | SmartDNS 后端和 LuCI | OpenWrt 25.12 对应 Feed 快照 |
| `adguardhome` | AdGuardHome 后端 | OpenWrt 25.12 对应 Feed 快照 |
| `luci-app-adguardhome-advanced` | 使用官方核心的高级 LuCI 管理界面 | 改编自 `rufengsuixing/luci-app-adguardhome`；适配说明见包内 `UPSTREAM.md` |
| 其余历史第三方目录 | 24.10 实际生成过的第三方包 | `kiddin9/kwrt-packages` 快照 |

OpenWrt 25.12 不再提供同名第三方 `luci-app-adguardhome`，避免覆盖官方包。
新增的 `luci-app-adguardhome-advanced` 只提供高级前端，依赖官方
`adguardhome` 核心；旧版大写 `/etc/init.d/AdGuardHome`、内置核心更新和
防火墙重定向脚本均未保留。

## `openwrt-24.10` / OpenWrt 24.10

以下包来自旧构建所使用的 `kiddin9/kwrt-packages` 快照：

- 仓库：`https://github.com/kiddin9/kwrt-packages.git`
- 快照提交：`ba60167deab5c88bb5f77d39ed61a5c549f95e98`

该分支保存上述 77 个历史源码目录及 SmartDNS、AdGuardHome、Netdata、
`ClashSubscriber`、`fullconenat-nft` 等原构建树包。SmartDNS 和 LuCI
使用 24.10 对应源码，以避免把 25.12 专用打包文件直接用于旧固件。

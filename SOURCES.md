# Package provenance

此仓库聚合多个上游项目的 OpenWrt 打包文件。各目录继续遵循其上游许可证、版权声明和商标规则；本仓库不对第三方目录重新授权。

## `main` / OpenWrt 25.12

| 目录 | 说明 | 来源 |
| --- | --- | --- |
| `ClashSubscriber` | 本地订阅转换脚本 | 原 24.10/25.12 构建树 |
| `luci-app-adguardhome` | AdGuardHome LuCI 管理页 | `rufengsuixing/luci-app-adguardhome` 衍生版本 |
| `luci-app-nikki` | Nikki LuCI 管理页 | 原 25.12 构建树 |
| `nikki` | Mihomo/Nikki 后端 | `MetaCubeX/mihomo`，含 25.12 构建兼容修复 |

## `openwrt-24.10` / OpenWrt 24.10

以下包来自旧构建所使用的 `kiddin9/kwrt-packages` 快照：

- 仓库：`https://github.com/kiddin9/kwrt-packages.git`
- 快照提交：`ba60167deab5c88bb5f77d39ed61a5c549f95e98`

包含：

- `ClashSubscriber`
- `fullconenat-nft`
- `linkease`
- `linkmount`
- `luci-app-clash`
- `luci-app-linkease`
- `luci-app-nikki`
- `luci-app-openclash`
- `luci-app-webadmin`
- `luci-theme-alpha`
- `luci-theme-argon`
- `nikki`

其中 `ClashSubscriber` 来自用户原构建树，其余目录保留原始 Makefile 中的上游信息。


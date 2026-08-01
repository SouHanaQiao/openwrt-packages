# SouHanaQiao OpenWrt Packages

个人维护的 OpenWrt 软件包 Feed 与二进制软件源，目标平台为 `armsr/armv8`，软件包架构为 `aarch64_generic`。

## 分支与格式

| 分支 | OpenWrt | 包格式 | 用途 |
| --- | --- | --- | --- |
| `main` | 25.12.5 | APK | 当前维护版本 |
| `openwrt-24.10` | 24.10.1 | IPK | 旧固件兼容版本 |

官方 OpenWrt 已提供的软件包不在此仓库重复维护。本仓库只保存自维护包和旧固件实际使用的第三方包。详细来源见 [SOURCES.md](SOURCES.md)。

## 源码 Feed

OpenWrt 25.12：

```text
src-git souhana https://github.com/SouHanaQiao/openwrt-packages.git;main
```

OpenWrt 24.10：

```text
src-git souhana https://github.com/SouHanaQiao/openwrt-packages.git;openwrt-24.10
```

加入 `feeds.conf.default` 后执行：

```sh
./scripts/feeds update souhana
./scripts/feeds install -a -f -p souhana
make menuconfig
```

## OpenWrt 25.12 APK 软件源

安装软件源公钥：

```sh
wget -O /etc/apk/keys/souhana-25.12.pem \
  https://souhanaqiao.github.io/openwrt-packages/25.12/aarch64_generic/public-key.pem
```

添加软件源：

```sh
echo 'https://souhanaqiao.github.io/openwrt-packages/25.12/aarch64_generic/packages.adb' \
  > /etc/apk/repositories.d/souhana.list
apk update
```

## OpenWrt 24.10 IPK 软件源

安装软件源公钥：

```sh
wget -O /etc/opkg/keys/a3d389f9a0aac186 \
  https://souhanaqiao.github.io/openwrt-packages/24.10/aarch64_generic/key-build.pub
```

添加软件源：

```sh
echo 'src/gz souhana https://souhanaqiao.github.io/openwrt-packages/24.10/aarch64_generic' \
  >> /etc/opkg/customfeeds.conf
opkg update
```

## 自动构建

GitHub Actions 使用与固件版本完全匹配的官方 SDK：

- OpenWrt 25.12.5 `armsr/armv8`，输出 APK 和 `packages.adb`。
- OpenWrt 24.10.1 `armsr/armv8`，输出 IPK、`Packages.gz` 和签名。

构建产物由 GitHub Pages 发布。推送到 `main`、`openwrt-24.10` 或手动运行工作流都会重新构建两个版本。

## 升级约定

升级软件包时必须提高 `PKG_VERSION` 或 `PKG_RELEASE`，否则路由器不会识别为新版本。

内核模块（例如 `kmod-nft-fullcone`）必须与运行中固件的内核 ABI 完全一致，不要跨固件版本安装。软件包在线升级也不能替代完整固件升级。


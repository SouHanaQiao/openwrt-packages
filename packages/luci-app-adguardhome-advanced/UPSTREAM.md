# Upstream and adaptation notes

This LuCI frontend was adapted from
[`rufengsuixing/luci-app-adguardhome`](https://github.com/rufengsuixing/luci-app-adguardhome)
as preserved in this repository's earlier package history.

The package intentionally does **not** ship, download, or update the AdGuard Home
binary. It depends on OpenWrt's official `adguardhome` package and controls only:

- `/usr/bin/AdGuardHome`
- `/etc/init.d/adguardhome`
- `/etc/config/adguardhome`
- `/etc/adguardhome/adguardhome.yaml`
- `/var/lib/adguardhome`

The legacy uppercase init service, bundled core updater, automatic binary
download, firewall redirect scripts, and obsolete configuration paths were
removed. `/etc/config/AdGuardHome` is retained only as a LuCI compatibility
configuration and is synchronized to the official lowercase UCI service.

The original upstream license and copyright notices remain applicable to the
adapted frontend files.

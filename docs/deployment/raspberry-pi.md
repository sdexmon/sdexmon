# Raspberry Pi 5 passive display

SDEXMON runs natively on 64-bit Ubuntu for Raspberry Pi 5. The release pipeline
builds a static `linux_arm64` binary, so no Go toolchain or desktop environment
is required on the device.

Ubuntu Server 24.04 LTS is recommended for a dedicated board. Connect the
display and network before booting, then install SDEXMON:

```bash
curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash
```

Test display mode interactively:

```bash
sdexmon display --pair XLM/USDC --rotate=0
```

`--rotate=30s` rotates through configured pairs. `--favorites-only` limits the
rotation to favorites; when no favorites exist, all configured pairs are used.
`--no-update-check` skips the GitHub release check, which is what you want on a
board with no one present to act on it.

Display mode never shows the upgrade notice, and the board retains its last good
snapshot. The footer shows data freshness instead of key hints, and appends
`reconnecting` while a data source is unavailable.

## Run on the HDMI console

Create an unprivileged service account and install the supplied unit:

```bash
sudo useradd --system --create-home --home-dir /var/lib/sdexmon sdexmon
sudo curl -fsSL \
  https://raw.githubusercontent.com/sdexmon/sdexmon/main/packaging/systemd/sdexmon.service \
  -o /etc/systemd/system/sdexmon.service
sudo systemctl daemon-reload
sudo systemctl disable --now getty@tty1.service
sudo systemctl enable --now sdexmon.service
```

The unit owns `/dev/tty1`, conflicts with `getty@tty1` so the two cannot fight
over the console, starts after `network-online.target`, and restarts 5 seconds
after an unexpected exit. It runs as the `sdexmon` user with `NoNewPrivileges`,
`PrivateTmp`, `ProtectHome` and `ProtectSystem=strict`.

Inspect non-TUI errors remotely with:

```bash
journalctl -u sdexmon.service -f
```

To prevent the Linux console from blanking, append `consoleblank=0` to the
kernel command line and reboot. Keep the kernel command line on one line.

For a fixed pair, edit `ExecStart` in the unit:

```ini
ExecStart=/usr/local/bin/sdexmon display --pair XLM/USDC --rotate=0 --no-update-check
```

An unknown `--pair` label is a fatal error, so a typo here stops the service
rather than silently showing the wrong market. Check `journalctl` if the unit
fails to start.

Run `sudo systemctl daemon-reload && sudo systemctl restart sdexmon` after
editing the unit.

## Configuration

The service sets `HOME=/var/lib/sdexmon`, so it reads
`/var/lib/sdexmon/.config/sdexmon/config.yaml`. Asset values must use `native`
or `CODE:ISSUER`; the issuer is part of the asset identity.

`ProtectSystem=strict` makes the filesystem read-only for the service, so the
board can read that config but not write it. That is intentional: display mode
never adds or removes pairs. Edit the file as root and restart the service, or
add a `ReadWritePaths=/var/lib/sdexmon` line to the unit if you need the service
itself to write.

Mark the pairs you want in the rotation:

```yaml
pairs:
  - name: XLM/USDC
    base: XLM:native
    quote: USDC:GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN
    favorite: true
```

See [Configuration](../guide/configuration.md) for the full schema.

## Hardware notes

Ethernet and SSD or NVMe storage are preferable for a continuously operating
board, although Wi-Fi and microSD are supported.

The layout targets 140x60 characters. On a smaller console the panels stack and
the row counts shrink automatically; on a larger one the content width is capped
at 180 columns.

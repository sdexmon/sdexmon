# Raspberry Pi 5 passive display

SDEXMON runs natively on 64-bit Ubuntu for Raspberry Pi 5. The release
pipeline builds a static `linux_arm64` binary, so no Go toolchain or desktop
environment is required on the device.

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
The board retains its last good snapshot and marks the footer as reconnecting
when a data source is temporarily unavailable.

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

The service owns `/dev/tty1`, starts after networking, and restarts after an
unexpected exit. Inspect non-TUI errors remotely with:

```bash
journalctl -u sdexmon.service -f
```

To prevent the Linux console from blanking, append `consoleblank=0` to the
kernel command line and reboot. Keep the kernel command line on one line.

For a fixed pair, edit `ExecStart` in the unit:

```ini
ExecStart=/usr/local/bin/sdexmon display --pair XLM/USDC --rotate=0 --no-update-check
```

Run `sudo systemctl daemon-reload && sudo systemctl restart sdexmon` after
editing the unit.

## Configuration

The service reads `/var/lib/sdexmon/.config/sdexmon/config.yaml`. Asset values
must use `native` or `CODE:ISSUER`; the issuer is part of the asset identity.

Ethernet and SSD/NVMe storage are preferable for a continuously operating
board, although Wi-Fi and microSD are supported.

# Upgrading

## Quick upgrade

Re-run the installer. It replaces the binary and the wrapper, and leaves
`~/.config/sdexmon/config.yaml` untouched.

```bash
curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash
```

## The startup update check

On startup SDEXMON asks the GitHub API for the latest release
(`https://api.github.com/repos/sdexmon/sdexmon/releases/latest`) with a 5 second
timeout. A failed check is logged as a warning and never blocks startup.

Version comparison is semantic. A `v` prefix and build metadata are ignored, and
a stable release is treated as newer than a prerelease of the same version. A
version that does not parse as `MAJOR.MINOR.PATCH` is never reported as
outdated, so development builds reporting `dev` are left alone.

When a newer release exists:

- The upgrade notice is shown before the landing screen.
- A `u: upgrade` shortcut appears in the footer on the landing, Pair Info and
  Pair Debug screens.

The notice is advisory. `esc` dismisses it, and `q` or `ctrl+c` always quit, so
an outdated build can never lock you out.

Display mode never shows the notice, because an unattended board must keep
rendering market data. Pass `--no-update-check` to skip the network call
entirely.

## Upgrading from inside the app

Press `enter` on the upgrade notice. SDEXMON hands the terminal to the installer
one-liner, and because the running binary is replaced in place, the app exits
when the installer finishes and prints:

```
Upgrade finished. Start sdexmon again to run the new version.
```

Start SDEXMON again to run the new build. If the installer fails, the error is
shown in the app and the current version keeps running.

## Manual upgrade

1. Check the current version:

   ```bash
   sdexmon --version
   ```

2. Download the release for your platform from
   <https://github.com/sdexmon/sdexmon/releases/latest> and verify it against
   `checksums.txt`.

3. Replace the binary. With a wrapper-based installation, the real binary is the
   hidden `.sdexmon-bin`; the wrapper itself should be left alone.

   ```bash
   sudo cp /usr/local/bin/.sdexmon-bin /usr/local/bin/.sdexmon-bin.backup
   sudo cp sdexmon /usr/local/bin/.sdexmon-bin
   sudo chmod 755 /usr/local/bin/.sdexmon-bin
   ```

4. Verify:

   ```bash
   sdexmon --version
   ```

## Troubleshooting

### Command not found after upgrading

Check that the wrapper survived:

```bash
ls -la /usr/local/bin/sdexmon
```

If it is missing, re-run the installer.

### Permission denied

The default install directory needs root. Install somewhere you own instead:

```bash
INSTALL_DIR=~/.local/bin \
  bash -c "$(curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh)"
export PATH="$HOME/.local/bin:$PATH"
```

### Still reporting the old version

Your shell has cached the old path:

```bash
hash -r
```

Or start a new shell.

### The version reads dev (build unknown)

`appVersion` and `gitCommit` are injected at build time. Plain `go build`,
`go run` and `go install` builds report the defaults. The `run` script injects
values from `git describe`, and release builds get them from GoReleaser.

## Version history notes

Releases v1.0.0 through v1.0.3 were published with an incorrect build
configuration and have been deprecated. The valid sequence runs from v0.1.0
upward. If you are on one of those builds, or installed before v0.1.1, see
[Migration](migration.md).

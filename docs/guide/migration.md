# Migration

This page is only relevant if you installed SDEXMON before v0.1.1.

## Symptom

The app opens on a "SCAR AQUILA" service selection screen instead of the
`sdexmon` banner.

## Background

Releases v1.0.0 through v1.0.3 were published with a `.goreleaser.yml` that
built from the wrong directory, and without the wrapper script. Those releases
have been deprecated and removed. The valid version sequence runs v0.1.0,
v0.1.1, and upward.

## Fix

Re-run the installer:

```bash
curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash
```

## What changed

Before, the installer placed the raw binary at `/usr/local/bin/sdexmon`.

Now it places the binary at `/usr/local/bin/.sdexmon-bin` and creates a small
wrapper at `/usr/local/bin/sdexmon` that sets two overridable defaults, sets the
terminal window title, and execs the binary:

```bash
export HORIZON_URL="${HORIZON_URL:-https://horizon.stellar.org}"
export DEBUG="${DEBUG:-false}"
```

The wrapper does not resize your terminal and does not force debug mode. If you
read an older version of this page that said otherwise, that documentation was
out of date.

## Overriding the defaults

Per invocation:

```bash
DEBUG=true sdexmon
HORIZON_URL=https://custom.horizon.endpoint sdexmon
DEBUG=true HORIZON_URL=https://custom.horizon.endpoint sdexmon
```

Or persistently, in `~/.bashrc` or `~/.zshrc`:

```bash
export DEBUG=true
export HORIZON_URL=https://custom.horizon.endpoint
```

See [Configuration](configuration.md) for every setting.

## Manual uninstall

```bash
sudo rm /usr/local/bin/sdexmon
sudo rm /usr/local/bin/.sdexmon-bin
```

`.sdexmon-bin` only exists on wrapper-based installations.

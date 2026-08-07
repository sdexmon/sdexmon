# Installation

## One-line installer (recommended)

```bash
curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash
```

The installer:

1. Detects your platform. Supported values are `linux`, `darwin` and `windows`
   (via MinGW, MSYS or Cygwin), on `amd64` or `arm64`.
2. Resolves the latest release tag from the GitHub API.
3. Downloads `sdexmon_<version>_<os>_<arch>.tar.gz` plus `checksums.txt` and
   verifies the SHA-256 digest before extracting anything.
4. Installs the binary as `<install dir>/.sdexmon-bin`.
5. Writes a small wrapper script at `<install dir>/sdexmon`.

`sha256sum` or `shasum` must be available; the installer refuses to continue
without one. `curl` and `tar` are also required.

### Choosing an install directory

`INSTALL_DIR` defaults to `/usr/local/bin`. The installer only escalates with
`sudo` when the target directory is not writable.

```bash
INSTALL_DIR=~/.local/bin \
  bash -c "$(curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh)"
```

If you install outside your `PATH`, add it:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### What the wrapper does

The wrapper is deliberately minimal. It sets two defaults that you can override
from your shell, sets the terminal window title, and execs the real binary:

```bash
export HORIZON_URL="${HORIZON_URL:-https://horizon.stellar.org}"
export DEBUG="${DEBUG:-false}"
```

It does not resize your terminal. See
[Configuration](configuration.md) for the full list of settings.

## Manual download

Release archives are published at
<https://github.com/sdexmon/sdexmon/releases/latest> for `linux`, `darwin` and
`windows` on `amd64` and `arm64`. Each archive contains the `sdexmon` binary,
`README.md`, `LICENSE`, the Raspberry Pi guide, and the systemd unit.

Verify the download against `checksums.txt`, then place the binary wherever you
like. If you are replacing a wrapper-based installation, overwrite
`.sdexmon-bin` rather than the wrapper. See [Upgrading](upgrading.md).

## Go install

Requires the Go toolchain declared in `go.mod`.

```bash
go install github.com/sdexmon/sdexmon/cmd/sdexmon@latest
```

Binaries built this way report `dev (build unknown)` for `--version`, because
`appVersion` and `gitCommit` are only injected by release builds.

## From source

```bash
git clone https://github.com/sdexmon/sdexmon.git
cd sdexmon
./run
```

The `run` script loads a local `.env` if present, sets `HORIZON_URL` and
`DEBUG=true`, requests a 140x60 terminal, and injects the version and commit
from `git describe`. To run without it:

```bash
go run ./cmd/sdexmon
```

## Verifying

```bash
sdexmon --version
```

This prints `<version> (build <commit>)` and exits without starting the TUI.

## Uninstalling

```bash
sudo rm /usr/local/bin/sdexmon
sudo rm /usr/local/bin/.sdexmon-bin
```

Your configuration at `~/.config/sdexmon/config.yaml` is left in place. Remove
it separately if you want a clean slate.

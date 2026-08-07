# Releasing

## Workflows

`.github/workflows/ci.yml` runs on pushes to `main` and on pull requests:

1. `go test -race ./...`
2. `go vet ./...`
3. A `linux/arm64` cross-build with `CGO_ENABLED=0`, which is the target the
   Raspberry Pi deployment depends on.

`.github/workflows/release.yml` runs on any `v*` tag. It repeats the test and
vet steps, then runs GoReleaser v2 with `release --clean` using the default
`GITHUB_TOKEN`.

Both workflows take the Go version from `go.mod`, so bumping the toolchain is a
single-file change.

## Cutting a release

```bash
git tag v0.2.3
git push origin v0.2.3
```

The release workflow does the rest. There is no manual GoReleaser step.

Versions must be `vMAJOR.MINOR.PATCH`. The in-app update check parses exactly
three numeric components, ignores a `v` prefix and build metadata, and treats a
stable release as newer than a prerelease of the same version. A tag that does
not parse will never be offered as an update to running clients.

Releases v1.0.0 through v1.0.3 were published with a broken build configuration
and have been deprecated. The valid sequence runs from v0.1.0 upward. Do not
reuse the v1.x range without accounting for clients that saw those tags.

## Build matrix

`.goreleaser.yml` builds `linux`, `darwin` and `windows` for `amd64` and
`arm64`, with:

```
-s -w
-X main.appVersion={{.Version}}
-X main.gitCommit={{.ShortCommit}}
```

This is the only path that produces a binary reporting a real version, so
`--version` output is a reliable signal of how a binary was built.

## Release assets

Each archive is named `sdexmon_{Version}_{Os}_{Arch}.tar.gz` and contains:

- the `sdexmon` binary
- `README.md`
- `LICENSE`
- `docs/deployment/raspberry-pi.md`
- `packaging/systemd/sdexmon.service`

A `checksums.txt` file is published alongside the archives.

If you move or rename any of the bundled documentation, update the `archives`
section of `.goreleaser.yml` in the same change. GoReleaser fails the release
when a listed file is missing.

## The installer contract

`install.sh` is served from `main`, not from a tag, so it always reflects the
current repository state. It depends on the following and will break if any of
it changes:

- The release tag is discoverable at
  `https://api.github.com/repos/sdexmon/sdexmon/releases/latest`.
- The archive is named `sdexmon_<version without v>_<os>_<arch>.tar.gz`.
- `checksums.txt` exists and lists that exact filename, with the digest first
  and the filename second.
- The archive extracts a top-level file called `sdexmon`.

The installer also defines the on-disk layout that the upgrade path relies on:
the real binary at `<install dir>/.sdexmon-bin` and a wrapper at
`<install dir>/sdexmon`.

The same one-liner is embedded in the binary as `upgradeCommand` and is what the
in-app upgrade executes, so the installer must remain safe to re-run over an
existing installation.

## After releasing

Verify the update path end to end, since it is the only part not covered by CI:

```bash
sdexmon --version
```

Then start an older build and confirm the upgrade notice appears with the new
version, and that `enter` runs the installer and exits cleanly.

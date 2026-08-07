# SDEXMON Documentation

SDEXMON is a terminal-native monitor for the Stellar Decentralized Exchange.
This folder holds the long-form documentation. The repository `README.md` is
the short landing page.

## Guide

For people running SDEXMON.

- [Installation](guide/installation.md) -- installer script, manual download,
  `go install`, and running from source.
- [Configuration](guide/configuration.md) -- environment variables, CLI flags,
  display mode, and the config file location.
- [Usage](guide/usage.md) -- every screen and every key binding.
- [Pair management](guide/pair-management.md) -- adding, removing and listing
  trading pairs, in the app and by hand.
- [Upgrading](guide/upgrading.md) -- the startup update check, the in-app
  upgrade, and manual upgrade steps.
- [Migration](guide/migration.md) -- for installations made before v0.1.1.

## Deployment

- [Raspberry Pi 5 passive display](deployment/raspberry-pi.md) -- unattended
  display mode on an HDMI console with systemd.

## Architecture

For people changing the code.

- [Overview](architecture/overview.md) -- package layout, data flow, polling
  intervals, and known technical debt.
- [Screens and routing](architecture/screens-and-routing.md) -- the screen
  state machine and the maintenance sub-state machine.
- [Configuration system](architecture/configuration-system.md) -- the YAML
  schema, the load path, fallbacks, and decimal precision resolution.

## Development

- [Building and testing](development/building-and-testing.md) -- Makefile
  targets, build flags, and the test suite.
- [Releasing](development/releasing.md) -- CI, GoReleaser, release assets, and
  the contract the installer depends on.

## Conventions

- Amounts use a space as the thousand separator, never a comma.
- Stellar amounts never exceed 7 decimal places. Displays show at least 2.
- These documents use plain ASCII so they render identically everywhere.

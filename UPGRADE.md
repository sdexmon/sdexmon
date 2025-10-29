# Upgrading sdexmon

## For Users with Older Versions

If you're running an older version of sdexmon that doesn't automatically check for updates, follow these steps to upgrade:

### Quick Upgrade (Recommended)

Run the following command to upgrade to the latest version:

```bash
curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash
```

This will:
- Download the latest release
- Replace your existing installation
- Preserve your existing configuration

### Manual Upgrade

If you prefer to manually upgrade:

1. **Check your current version:**
   ```bash
   sdexmon --version
   ```

2. **Download the latest release** from GitHub:
   - Visit: https://github.com/sdexmon/sdexmon/releases/latest
   - Download the appropriate binary for your platform

3. **Replace the existing binary:**
   ```bash
   # Backup your current version (optional)
   sudo cp /usr/local/bin/.sdexmon-bin /usr/local/bin/.sdexmon-bin.backup
   
   # Install new binary
   sudo cp sdexmon /usr/local/bin/.sdexmon-bin
   sudo chmod 755 /usr/local/bin/.sdexmon-bin
   ```

4. **Verify the upgrade:**
   ```bash
   sdexmon --version
   ```

### What's New

Starting from **v0.1.2**, sdexmon includes:
- **Automatic version checking** on startup
- **Forced upgrades** when critical updates are available
- You'll be prompted to upgrade if a newer version is detected

### Troubleshooting

**"Command not found" after upgrade:**
```bash
# Check if the wrapper script exists
ls -la /usr/local/bin/sdexmon

# If missing, reinstall:
curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash
```

**Permission denied:**
```bash
# The install script requires sudo to write to /usr/local/bin
# If you don't have sudo access, install to a local directory:
INSTALL_DIR=~/.local/bin bash -c "$(curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh)"

# Then add to your PATH if needed:
export PATH="$HOME/.local/bin:$PATH"
```

**Still showing old version:**
```bash
# Clear any cached binaries
hash -r

# Or restart your shell
exec $SHELL
```

## Version History

- **v0.1.2** - Added automatic version checking and forced upgrades
- **v0.1.1** - Fixed build configuration and wrapper script
- **v0.1.0** - Initial release with maintenance mode

## Need Help?

If you encounter any issues during the upgrade:
1. Check the [GitHub Issues](https://github.com/sdexmon/sdexmon/issues)
2. Open a new issue with details about your platform and the error message

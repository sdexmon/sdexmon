# Migration Guide

## For Users Who Installed Before Wrapper Script

If you installed `sdexmon` before version with the wrapper script and notice it looks different than expected, you need to reinstall:

### Quick Fix

```bash
# Reinstall with the updated installer
curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash
```

This will:
1. Download the latest binary as `.sdexmon-bin` (hidden)
2. Create a wrapper script at `sdexmon` that sets proper environment
3. Automatically enable debug mode and optimal terminal size

### What Changed?

**Before:** The installer placed the raw binary at `/usr/local/bin/sdexmon`

**After:** The installer now:
- Places the binary at `/usr/local/bin/.sdexmon-bin` (hidden)
- Creates a wrapper script at `/usr/local/bin/sdexmon` that:
  - Sets `DEBUG=true`
  - Sets optimal terminal size (140×60)
  - Sets default Horizon URL
  - Runs the actual binary

### Manual Uninstall (if needed)

```bash
sudo rm /usr/local/bin/sdexmon
sudo rm /usr/local/bin/.sdexmon-bin  # Only exists after wrapper installation
```

### Environment Variable Override

The wrapper sets defaults, but you can override them:

```bash
# Disable debug mode
DEBUG=false sdexmon

# Use custom Horizon endpoint
HORIZON_URL=https://custom.horizon.endpoint sdexmon

# Both
DEBUG=false HORIZON_URL=https://custom.horizon.endpoint sdexmon
```

Or export them in your shell config:

```bash
# Add to ~/.bashrc or ~/.zshrc
export DEBUG=false
export HORIZON_URL=https://custom.horizon.endpoint
```

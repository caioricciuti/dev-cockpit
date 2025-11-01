# Troubleshooting

This guide helps you resolve common issues when using Dev Cockpit.

## Installation Issues

### Command Not Found After Installation

**Problem:** Running `devcockpit` shows "command not found"

**Solution:**
1. Verify the binary is installed:
   ```bash
   ls -la /usr/local/bin/devcockpit
   ```

2. If missing, reinstall:
   ```bash
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/caioricciuti/dev-cockpit/main/install.sh)"
   ```

3. Check your PATH includes `/usr/local/bin`:
   ```bash
   echo $PATH
   ```

4. If not in PATH, add to your shell profile (`~/.zshrc` or `~/.bashrc`):
   ```bash
   export PATH="/usr/local/bin:$PATH"
   ```

### Permission Denied

**Problem:** Error "Permission denied" when running installer

**Solution:**
```bash
sudo /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/caioricciuti/dev-cockpit/main/install.sh)"
```

## Package Manager Issues

### npm Not Detected or "exit status 127"

**Problem:** Package module shows "Unknown" for npm or fails with "exit status 127"

**Solution:**

**If using NVM (Node Version Manager):**
1. Dev Cockpit automatically detects NVM installations in `~/.nvm/`
2. Ensure NVM's default version is set:
   ```bash
   nvm alias default node
   ```
3. Verify NVM directory exists:
   ```bash
   ls -la ~/.nvm/
   ```

**If using Homebrew Node:**
1. Verify npm is installed:
   ```bash
   which npm
   npm --version
   ```
2. Reinstall if needed:
   ```bash
   brew reinstall node
   ```

**If using system Node:**
1. Ensure `/usr/local/bin` is in your PATH
2. Restart Dev Cockpit after installing Node

### Homebrew Commands Fail

**Problem:** Homebrew operations fail or packages don't appear

**Solution:**
1. Verify Homebrew is installed:
   ```bash
   which brew
   brew --version
   ```

2. Update Homebrew:
   ```bash
   brew update
   ```

3. Check Homebrew paths (Apple Silicon uses `/opt/homebrew`):
   ```bash
   ls -la /opt/homebrew/bin/brew
   ```

4. Add Homebrew to PATH if needed:
   ```bash
   echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zshrc
   ```

### Docker Not Detected

**Problem:** Docker module shows "Not installed" but Docker Desktop is running

**Solution:**
1. Ensure Docker Desktop is running
2. Verify Docker CLI is accessible:
   ```bash
   which docker
   docker --version
   ```

3. If using Colima or another Docker alternative, ensure the socket is at:
   ```bash
   ls -la /var/run/docker.sock
   ```

4. Restart Docker Desktop and Dev Cockpit

## Cleanup Module Issues

### Operation Timeout

**Problem:** Cleanup operations timeout, especially for large directories

**Solution:**
- This is expected for very large caches (e.g., 50GB+ Xcode DerivedData)
- The operation continues in background even if timeout message appears
- Check `~/.devcockpit/debug.log` for actual completion status
- For manual cleanup of large directories:
  ```bash
  rm -rf ~/Library/Developer/Xcode/DerivedData/*
  ```

### Permission Denied During Cleanup

**Problem:** Some cleanup operations fail with permission errors

**Solution:**
1. Run Dev Cockpit with sudo for system-level cleanup:
   ```bash
   sudo devcockpit
   ```

2. For user-level cleanup, ensure you have write permissions:
   ```bash
   ls -la ~/Library/Caches/
   ```

3. Reset permissions if needed:
   ```bash
   chmod -R u+w ~/Library/Caches/
   ```

### Empty Trash Fails

**Problem:** "Empty Trash" operation fails or hangs

**Solution:**
1. Close any Finder windows
2. Run from command line:
   ```bash
   devcockpit cleanup empty-trash
   ```
3. If still failing, try macOS native command:
   ```bash
   rm -rf ~/.Trash/*
   ```

## Terminal and Display Issues

### Colors Not Displaying Correctly

**Problem:** Terminal shows garbled characters or wrong colors

**Solution:**
1. Use a modern terminal that supports 256 colors:
   - iTerm2 (recommended)
   - Terminal.app (macOS built-in)
   - Alacritty
   - Kitty

2. Verify TERM environment variable:
   ```bash
   echo $TERM
   ```
   Should be `xterm-256color` or similar

3. Set TERM if needed:
   ```bash
   export TERM=xterm-256color
   ```

### TUI Layout Broken

**Problem:** Interface looks broken or overlapping

**Solution:**
1. Resize terminal window (Dev Cockpit adapts automatically)
2. Minimum recommended size: 80x24 characters
3. Press `Ctrl+L` to force redraw
4. Restart Dev Cockpit

### Text Too Small/Large

**Problem:** Text is difficult to read

**Solution:**
1. Adjust terminal font size (usually `Cmd + +` or `Cmd + -`)
2. Use a monospace font (Fira Code, JetBrains Mono, etc.)
3. Ensure terminal window is at least 80 characters wide

## Performance Issues

### High CPU Usage

**Problem:** Dev Cockpit uses significant CPU

**Solution:**
1. This is normal during:
   - Initial system scan
   - Package listing operations
   - Large cleanup operations
2. CPU usage should drop after initial operations complete
3. Close other resource-intensive applications
4. Check Activity Monitor for actual usage

### Slow Response

**Problem:** Commands take a long time to respond

**Solution:**
1. Check system resources (CPU, memory, disk)
2. Ensure no other package managers are running:
   ```bash
   ps aux | grep -E "brew|npm|docker"
   ```
3. Restart Dev Cockpit
4. Reboot Mac if system is generally slow

## Navigation and Controls

### Stuck on Result Screen

**Problem:** Can't exit from cleanup or package operation results

**Solution:**
- Press **any key** to dismiss result screens
- Press **ESC** to return to module home
- Press **ESC** again to return to module switcher

### Modal Won't Close

**Problem:** Package list or other modal won't close

**Solution:**
1. Press **ESC** to close modal and return to module
2. Press **q** as alternative quit key
3. If frozen, press **Ctrl+C** to force quit Dev Cockpit

### Can't Navigate Between Modules

**Problem:** Tab switching doesn't work

**Solution:**
1. Ensure you're not in a module (press **ESC** first)
2. Use number keys (**1-9**) or **Tab** to switch
3. Use **←** and **→** arrow keys to navigate tabs
4. Make sure no modal is open (**ESC** to close)

## Debug Mode

### Enabling Debug Logs

To troubleshoot issues, enable debug logging:

```bash
devcockpit --debug
```

Logs are written to:
```bash
~/.devcockpit/debug.log
```

View logs:
```bash
tail -f ~/.devcockpit/debug.log
```

### Common Error Messages

**"Failed to get system info"**
- Usually temporary, caused by system API delays
- Restart Dev Cockpit
- Check Activity Monitor for system health

**"Sudo password required"**
- Some operations need elevated privileges
- Run with: `sudo devcockpit`
- Or grant sudo access when prompted

**"Command not found: brew/npm/docker"**
- Package manager not in PATH
- Restart terminal after installing package managers
- See "Package Manager Issues" section above

## Getting Help

If you're still experiencing issues:

1. **Check existing issues:**
   [GitHub Issues](https://github.com/caioricciuti/dev-cockpit/issues)

2. **Create a new issue:**
   Include:
   - macOS version and chip (M1/M2/M3)
   - Dev Cockpit version (`devcockpit --version`)
   - Terminal app being used
   - Steps to reproduce
   - Error messages or screenshots
   - Debug log if relevant

3. **System Information:**
   Helpful details to include:
   ```bash
   # macOS version
   sw_vers

   # Chip type
   uname -m  # Should show "arm64"

   # Package manager versions
   brew --version
   npm --version
   docker --version

   # Dev Cockpit version
   devcockpit --version
   ```

## Uninstalling Dev Cockpit

If you need to uninstall, use the built-in command:

```bash
devcockpit uninstall
```

This interactive command will:
- Check if Dev Cockpit is running and offer to stop it
- Remove the binary from `/usr/local/bin/devcockpit` (requests sudo if needed)
- Prompt to remove configuration directory (`~/.devcockpit/`)
- Clean up temporary files (`/tmp/devcockpit-*`)

For non-interactive uninstallation:
```bash
devcockpit uninstall --force
```

**Manual uninstallation** (fallback):
```bash
# Remove binary
sudo rm /usr/local/bin/devcockpit

# Remove config (optional)
rm -rf ~/.devcockpit
```

## Reporting Bugs

When reporting bugs, please include:
- Clear description of the issue
- Steps to reproduce
- Expected vs. actual behavior
- macOS version and chip type
- Error messages or screenshots
- Debug log excerpt if relevant

Create an issue at: [https://github.com/caioricciuti/dev-cockpit/issues](https://github.com/caioricciuti/dev-cockpit/issues)

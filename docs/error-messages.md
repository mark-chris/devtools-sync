# Error Messages Guide

The `devtools-sync` CLI provides clear, actionable error messages to help you quickly resolve issues.

## Error Message Principles

All error messages in devtools-sync follow these principles:

1. **Clear and Specific:** Explains exactly what went wrong
2. **Actionable:** Tells you what to do next
3. **Contextual:** Includes relevant details to help diagnose the issue

## Common Errors and Solutions

### Configuration Errors

#### Config Already Exists
```
Error: configuration already exists at /home/user/.devtools-sync/config.yaml

To reconfigure, either:
  1. Edit the file directly, or
  2. Delete it and run 'devtools-sync init' again, or
  3. Use 'devtools-sync config set <key> <value>' to update specific values
```

**Solution:** Choose one of the suggested approaches based on your needs.

#### Config Not Found
```
Error: failed to load config: open /home/user/.devtools-sync/config.yaml: no such file or directory

Run 'devtools-sync init' to create the configuration file
```

**Solution:** Run `devtools-sync init` to create the default configuration.

### Profile Errors

#### Profile Not Found (No Profiles Available)
```
Error: profile 'work-setup' not found

No profiles available. Create one with:
  devtools-sync profile save <name>
```

**Solution:** Save your first profile using `devtools-sync profile save <name>`.

#### Profile Not Found (Profiles Available)
```
Error: profile 'typo' not found

Available profiles: work-setup, personal, minimal

Use 'devtools-sync profile list' to see all profiles
```

**Solution:** Check the available profiles and use the correct name, or create a new profile.

#### VS Code Not Found (Save)
```
Error: failed to save profile: failed to list VS Code extensions: exec: "code": executable file not found in $PATH

Make sure:
  1. VS Code is installed
  2. The 'code' command is available in your PATH
  3. You can run 'code --version' successfully
```

**Solutions:**
1. Install VS Code from https://code.visualstudio.com/
2. Add VS Code to your PATH (usually done during installation)
3. On macOS, run Command Palette (Cmd+Shift+P) → "Shell Command: Install 'code' command in PATH"
4. On Linux, ensure `/usr/bin/code` or `/snap/bin/code` is in your PATH

#### VS Code Not Found (Load)
```
Error: failed to load profile: failed to install extension: exec: "code": executable file not found in $PATH

Make sure:
  1. VS Code is installed
  2. The 'code' command is available in your PATH
```

**Solution:** Same as above - ensure VS Code CLI is accessible.

### Sync Errors

#### Server Connection Failed
```
Error: failed to connect to server at http://localhost:8080: dial tcp [::1]:8080: connect: connection refused

Make sure:
  1. The server is running
  2. The server URL is correct (check with 'devtools-sync config show')
  3. You can reach the server from your network
```

**Solutions:**
1. Start the devtools-sync server: `cd server && go run main.go`
2. Verify server URL: `devtools-sync config show`
3. Update server URL if needed: `devtools-sync config set server.url http://correct-url:8080`
4. Test connectivity: `curl http://server-url:8080/health`

#### Server Connection Failed (DNS Error)
```
Error: failed to list server profiles: Get "http://nonexistent:8080/api/v1/profiles": dial tcp: lookup nonexistent: no such host

Check your server connection with:
  curl http://nonexistent:8080/health
```

**Solutions:**
1. Verify the hostname/domain is correct
2. Check DNS resolution: `nslookup nonexistent`
3. Update to correct server URL: `devtools-sync config set server.url http://correct-url:8080`

### Validation Errors

#### Invalid Config Key
```
Error: unknown config section: typo
```

**Solution:** Use tab completion or check valid keys:
- `server.url`
- `profiles.directory`
- `logging.level`

#### Invalid URL Scheme
```
Error: invalid configuration: server URL must use http or https scheme, got: ftp
```

**Solution:** Use HTTP or HTTPS scheme:
```bash
devtools-sync config set server.url https://server:8080
```

## Error Codes

The CLI uses standard exit codes:

- `0` - Success
- `1` - General error
- `2` - Misuse of command (invalid arguments)

## Debugging

### Enable Verbose Logging

While not yet implemented, future versions will support:
```bash
devtools-sync --log-level debug <command>
```

### Check Current Configuration

```bash
devtools-sync config show
```

### Test Server Connectivity

```bash
# Check if server is reachable
curl $(devtools-sync config show | grep URL | awk '{print $2}')/health

# Should return:
# {"status":"healthy","service":"devtools-sync-server"}
```

### List Available Profiles

```bash
devtools-sync profile list
```

## Getting Help

If you encounter an error not covered here:

1. Check the error message for actionable guidance
2. Run the suggested commands from the error message
3. Use `--help` on any command for usage information:
   ```bash
   devtools-sync <command> --help
   ```
4. Report issues at: https://github.com/mark-chris/devtools-sync/issues

## Best Practices

1. **Always run `init` first:** Before using the CLI, run `devtools-sync init`
2. **Verify config:** Use `devtools-sync config show` to check your configuration
3. **Test connectivity:** Before syncing, ensure you can reach the server
4. **Use completion:** Install shell completion for better UX (see shell-completion.md)
5. **Read error messages:** They contain specific guidance for resolution

## Examples

### First-Time Setup
```bash
# 1. Initialize configuration
devtools-sync init

# 2. Verify config (optional)
devtools-sync config show

# 3. Save current extensions
devtools-sync profile save my-setup

# 4. List profiles to verify
devtools-sync profile list
```

### Troubleshooting Sync Issues
```bash
# 1. Check current config
devtools-sync config show

# 2. Test server connectivity
curl $(devtools-sync config show | grep URL | awk '{print $2}')/health

# 3. If server URL is wrong, update it
devtools-sync config set server.url http://correct-url:8080

# 4. Try syncing again
devtools-sync sync push
```

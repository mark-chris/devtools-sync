# Extension Installation via Code CLI — Issue #14

## Overview

Enhance the agent's extension management to support install, uninstall, and bulk operations with progress reporting, failure handling, and dry-run support. No new CLI commands — functionality added to the `vscode` and `profile` packages for use by existing commands like `sync pull`.

## Changes

### 1. UninstallExtension

Add `UninstallExtension(id string) error` to `agent/internal/vscode/vscode.go`. Calls `code --uninstall-extension <id>`. Same pattern as existing `InstallExtension`.

### 2. BulkInstall / BulkUninstall

```go
type BulkResult struct {
    ID    string
    Error error
}

func BulkInstall(extensions []string, w io.Writer, dryRun bool) []BulkResult
func BulkUninstall(extensions []string, w io.Writer, dryRun bool) []BulkResult
```

- Sequential iteration (VS Code CLI doesn't support parallel installs)
- Progress to `w`: `[1/5] Installing publisher.extension... done` or `... FAILED: <error>`
- Dry-run: `[1/5] Would install publisher.extension` (no execution)
- Returns results for callers to summarize

### 3. Update profile.Load()

Accept `io.Writer` and `dryRun bool` parameters, delegate to `vscode.BulkInstall` instead of manual loop. Gives `sync pull` progress reporting and dry-run for free.

### 4. --dry-run flag on sync pull

Add `--dry-run` flag to `sync pull` command. Pass through to `profile.Load()`.

## Testing

- Make command execution injectable for unit testing (currently uses `exec.Command` directly)
- Test BulkInstall/BulkUninstall with mock executor
- Test dry-run produces correct output without executing
- Test failure handling (one fails, others continue)

## Not Included (YAGNI)

- No new CLI commands
- No progress bars
- No parallel installation
- No extension dependency resolution

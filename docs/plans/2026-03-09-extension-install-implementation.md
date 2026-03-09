# Extension Installation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add uninstall, bulk operations with progress reporting, failure handling, and dry-run support to the agent's extension management.

**Architecture:** Make command execution injectable in the `vscode` package for testability. Add `UninstallExtension`, `BulkInstall`, and `BulkUninstall` functions. Update `profile.Load()` to use bulk install with progress and dry-run. Add `--dry-run` flag to `sync pull`.

**Tech Stack:** Go 1.25, cobra CLI, os/exec

---

### Task 1: Make command execution injectable for testing

The `vscode` package calls `exec.Command` directly, making it impossible to unit test install/uninstall without actually running VS Code. Introduce a package-level `CommandRunner` that tests can replace.

**Files:**
- Modify: `agent/internal/vscode/vscode.go`
- Create: `agent/internal/vscode/bulk_test.go`

**Step 1: Add CommandRunner abstraction**

Add at the top of `agent/internal/vscode/vscode.go`, after the imports:

```go
// CommandRunner executes external commands. Override in tests.
var CommandRunner = func(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}
```

**Step 2: Update InstallExtension to use CommandRunner**

Replace the existing `InstallExtension` function body:

```go
func InstallExtension(extensionID string) error {
	if extensionID == "" {
		return errors.New("extension ID cannot be empty")
	}

	output, err := CommandRunner("code", "--install-extension", extensionID)
	if err != nil {
		return fmt.Errorf("failed to install extension %s: %w (output: %s)", extensionID, err, string(output))
	}

	return nil
}
```

**Step 3: Run existing tests to verify no regression**

Run: `cd agent && go test -v ./internal/vscode/ -run TestInstallExtension`
Expected: PASS

**Step 4: Commit**

```bash
git add agent/internal/vscode/vscode.go
git commit -m "refactor(agent): make command execution injectable for testing (issue #14)"
```

---

### Task 2: Add UninstallExtension

**Files:**
- Modify: `agent/internal/vscode/vscode.go`
- Create: `agent/internal/vscode/bulk_test.go`

**Step 1: Write the failing test**

Create `agent/internal/vscode/bulk_test.go`:

```go
package vscode

import (
	"bytes"
	"errors"
	"testing"
)

func TestUninstallExtension_EmptyID(t *testing.T) {
	err := UninstallExtension("")
	if err == nil {
		t.Error("expected error for empty extension ID")
	}
	if err.Error() != "extension ID cannot be empty" {
		t.Errorf("expected 'extension ID cannot be empty', got: %s", err.Error())
	}
}

func TestUninstallExtension_Success(t *testing.T) {
	// Save and restore original CommandRunner
	original := CommandRunner
	defer func() { CommandRunner = original }()

	var calledWith []string
	CommandRunner = func(name string, args ...string) ([]byte, error) {
		calledWith = append([]string{name}, args...)
		return []byte("Extension 'test.ext' was successfully uninstalled."), nil
	}

	err := UninstallExtension("test.ext")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if len(calledWith) != 3 || calledWith[0] != "code" || calledWith[1] != "--uninstall-extension" || calledWith[2] != "test.ext" {
		t.Errorf("expected 'code --uninstall-extension test.ext', got: %v", calledWith)
	}
}

func TestUninstallExtension_Failure(t *testing.T) {
	original := CommandRunner
	defer func() { CommandRunner = original }()

	CommandRunner = func(name string, args ...string) ([]byte, error) {
		return []byte("error output"), errors.New("exit status 1")
	}

	err := UninstallExtension("test.ext")
	if err == nil {
		t.Error("expected error on command failure")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd agent && go test -v ./internal/vscode/ -run TestUninstallExtension`
Expected: FAIL — `UninstallExtension` not defined

**Step 3: Write the implementation**

Add to `agent/internal/vscode/vscode.go` after `InstallExtension`:

```go
// UninstallExtension uninstalls a VS Code extension by ID
func UninstallExtension(extensionID string) error {
	if extensionID == "" {
		return errors.New("extension ID cannot be empty")
	}

	output, err := CommandRunner("code", "--uninstall-extension", extensionID)
	if err != nil {
		return fmt.Errorf("failed to uninstall extension %s: %w (output: %s)", extensionID, err, string(output))
	}

	return nil
}
```

**Step 4: Run tests to verify they pass**

Run: `cd agent && go test -v ./internal/vscode/ -run TestUninstallExtension`
Expected: PASS — all 3 tests

**Step 5: Commit**

```bash
git add agent/internal/vscode/vscode.go agent/internal/vscode/bulk_test.go
git commit -m "feat(agent): add UninstallExtension function (issue #14)"
```

---

### Task 3: Add BulkInstall and BulkUninstall

**Files:**
- Create: `agent/internal/vscode/bulk.go`
- Modify: `agent/internal/vscode/bulk_test.go`

**Step 1: Write the failing tests**

Append to `agent/internal/vscode/bulk_test.go`:

```go
func TestBulkInstall_Success(t *testing.T) {
	original := CommandRunner
	defer func() { CommandRunner = original }()

	installed := []string{}
	CommandRunner = func(name string, args ...string) ([]byte, error) {
		installed = append(installed, args[1]) // args[1] is the extension ID
		return []byte("ok"), nil
	}

	var buf bytes.Buffer
	results := BulkInstall([]string{"pub.ext1", "pub.ext2", "pub.ext3"}, &buf, false)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("expected no error for %s, got: %v", r.ID, r.Error)
		}
	}
	if len(installed) != 3 {
		t.Errorf("expected 3 installs, got %d", len(installed))
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("[1/3]")) {
		t.Errorf("expected progress output with [1/3], got: %s", output)
	}
	if !bytes.Contains([]byte(output), []byte("[3/3]")) {
		t.Errorf("expected progress output with [3/3], got: %s", output)
	}
}

func TestBulkInstall_PartialFailure(t *testing.T) {
	original := CommandRunner
	defer func() { CommandRunner = original }()

	CommandRunner = func(name string, args ...string) ([]byte, error) {
		if args[1] == "pub.ext2" {
			return []byte("error"), errors.New("exit status 1")
		}
		return []byte("ok"), nil
	}

	var buf bytes.Buffer
	results := BulkInstall([]string{"pub.ext1", "pub.ext2", "pub.ext3"}, &buf, false)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Error != nil {
		t.Error("expected ext1 to succeed")
	}
	if results[1].Error == nil {
		t.Error("expected ext2 to fail")
	}
	if results[2].Error != nil {
		t.Error("expected ext3 to succeed (continues after failure)")
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("FAILED")) {
		t.Errorf("expected FAILED in output, got: %s", output)
	}
}

func TestBulkInstall_DryRun(t *testing.T) {
	original := CommandRunner
	defer func() { CommandRunner = original }()

	called := false
	CommandRunner = func(name string, args ...string) ([]byte, error) {
		called = true
		return nil, nil
	}

	var buf bytes.Buffer
	results := BulkInstall([]string{"pub.ext1", "pub.ext2"}, &buf, true)

	if called {
		t.Error("dry-run should not execute commands")
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("dry-run should not produce errors, got: %v", r.Error)
		}
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Would install")) {
		t.Errorf("expected 'Would install' in dry-run output, got: %s", output)
	}
}

func TestBulkInstall_Empty(t *testing.T) {
	var buf bytes.Buffer
	results := BulkInstall([]string{}, &buf, false)
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty input, got %d", len(results))
	}
}

func TestBulkUninstall_Success(t *testing.T) {
	original := CommandRunner
	defer func() { CommandRunner = original }()

	CommandRunner = func(name string, args ...string) ([]byte, error) {
		return []byte("ok"), nil
	}

	var buf bytes.Buffer
	results := BulkUninstall([]string{"pub.ext1", "pub.ext2"}, &buf, false)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("expected no error for %s, got: %v", r.ID, r.Error)
		}
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Uninstalling")) {
		t.Errorf("expected 'Uninstalling' in output, got: %s", output)
	}
}

func TestBulkUninstall_DryRun(t *testing.T) {
	original := CommandRunner
	defer func() { CommandRunner = original }()

	called := false
	CommandRunner = func(name string, args ...string) ([]byte, error) {
		called = true
		return nil, nil
	}

	var buf bytes.Buffer
	results := BulkUninstall([]string{"pub.ext1"}, &buf, true)

	if called {
		t.Error("dry-run should not execute commands")
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Would uninstall")) {
		t.Errorf("expected 'Would uninstall' in dry-run output, got: %s", output)
	}

	if len(results) != 1 || results[0].Error != nil {
		t.Error("dry-run should return success results")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd agent && go test -v ./internal/vscode/ -run TestBulk`
Expected: FAIL — `BulkInstall`, `BulkUninstall`, `BulkResult` not defined

**Step 3: Write the implementation**

Create `agent/internal/vscode/bulk.go`:

```go
package vscode

import (
	"fmt"
	"io"
)

// BulkResult records the outcome of a single extension operation.
type BulkResult struct {
	ID    string
	Error error
}

// BulkInstall installs extensions sequentially with progress reporting.
// In dry-run mode, reports what would happen without executing.
// Continues on failure — all extensions are attempted regardless of individual errors.
func BulkInstall(extensions []string, w io.Writer, dryRun bool) []BulkResult {
	results := make([]BulkResult, 0, len(extensions))
	total := len(extensions)

	for i, id := range extensions {
		if dryRun {
			fmt.Fprintf(w, "[%d/%d] Would install %s\n", i+1, total, id)
			results = append(results, BulkResult{ID: id})
			continue
		}

		fmt.Fprintf(w, "[%d/%d] Installing %s... ", i+1, total, id)
		err := InstallExtension(id)
		if err != nil {
			fmt.Fprintf(w, "FAILED: %v\n", err)
			results = append(results, BulkResult{ID: id, Error: err})
		} else {
			fmt.Fprintf(w, "done\n")
			results = append(results, BulkResult{ID: id})
		}
	}

	return results
}

// BulkUninstall uninstalls extensions sequentially with progress reporting.
// In dry-run mode, reports what would happen without executing.
// Continues on failure — all extensions are attempted regardless of individual errors.
func BulkUninstall(extensions []string, w io.Writer, dryRun bool) []BulkResult {
	results := make([]BulkResult, 0, len(extensions))
	total := len(extensions)

	for i, id := range extensions {
		if dryRun {
			fmt.Fprintf(w, "[%d/%d] Would uninstall %s\n", i+1, total, id)
			results = append(results, BulkResult{ID: id})
			continue
		}

		fmt.Fprintf(w, "[%d/%d] Uninstalling %s... ", i+1, total, id)
		err := UninstallExtension(id)
		if err != nil {
			fmt.Fprintf(w, "FAILED: %v\n", err)
			results = append(results, BulkResult{ID: id, Error: err})
		} else {
			fmt.Fprintf(w, "done\n")
			results = append(results, BulkResult{ID: id})
		}
	}

	return results
}
```

**Step 4: Run tests to verify they pass**

Run: `cd agent && go test -v ./internal/vscode/ -run TestBulk`
Expected: PASS — all 7 tests

**Step 5: Commit**

```bash
git add agent/internal/vscode/bulk.go agent/internal/vscode/bulk_test.go
git commit -m "feat(agent): add BulkInstall and BulkUninstall with progress and dry-run (issue #14)"
```

---

### Task 4: Update profile.Load() to use BulkInstall

**Files:**
- Modify: `agent/internal/profile/profile.go`

**Step 1: Update Load() signature and implementation**

Change the `Load` function signature from:
```go
func Load(name string, profilesDir string) (*Profile, error) {
```
To:
```go
func Load(name string, profilesDir string, w io.Writer, dryRun bool) (*Profile, error) {
```

Add `"io"` to the imports.

Replace the install loop and reporting section (lines 248–267) with:

```go
	// Report skipped extensions
	if len(alreadyInstalled) > 0 {
		fmt.Fprintf(w, "Skipping %d already installed extension(s):\n", len(alreadyInstalled))
		for _, ext := range alreadyInstalled {
			fmt.Fprintf(w, "  - %s (already installed)\n", ext.ID)
		}
	}

	// Bulk install new extensions with progress
	ids := make([]string, len(toInstall))
	for i, ext := range toInstall {
		ids[i] = ext.ID
	}

	results := vscode.BulkInstall(ids, w, dryRun)

	// Count successes and failures
	var failed int
	for _, r := range results {
		if r.Error != nil {
			failed++
		}
	}

	// Report summary
	action := "Installed"
	if dryRun {
		action = "Would install"
	}
	fmt.Fprintf(w, "\nProfile '%s' summary:\n", profile.Name)
	fmt.Fprintf(w, "  - %s: %d extension(s)\n", action, len(toInstall)-failed)
	if failed > 0 {
		fmt.Fprintf(w, "  - Failed: %d extension(s)\n", failed)
	}
	fmt.Fprintf(w, "  - Skipped: %d extension(s)\n", len(alreadyInstalled))
	fmt.Fprintf(w, "  - Total: %d extension(s)\n", len(profile.Extensions))

	if failed > 0 {
		return &profile, fmt.Errorf("%d extension(s) failed to install", failed)
	}
```

**Step 2: Update all callers of Load()**

In `agent/internal/profile/profile_test.go`, update the `TestLoad` calls from:
```go
Load("nonexistent", tempDir)
Load("", tempDir)
```
To:
```go
Load("nonexistent", tempDir, io.Discard, false)
Load("", tempDir, io.Discard, false)
```

Add `"io"` to the test file imports.

Search for any other callers of `profile.Load` in the codebase. Currently none exist in `sync.go` (sync pull saves to disk but doesn't call Load).

**Step 3: Run tests**

Run: `cd agent && go test -v ./...`
Expected: All tests pass

**Step 4: Commit**

```bash
git add agent/internal/profile/profile.go agent/internal/profile/profile_test.go
git commit -m "feat(agent): update profile.Load() to use BulkInstall with progress (issue #14)"
```

---

### Task 5: Add --dry-run flag to sync pull

**Files:**
- Modify: `agent/cmd/sync.go`

**Step 1: Add the flag**

In `agent/cmd/sync.go`, add to the `init()` function:

```go
func init() {
	syncPullCmd.Flags().BoolVar(&syncPullDryRun, "dry-run", false, "Show what would be synced without making changes")
	syncCmd.AddCommand(syncPushCmd)
	syncCmd.AddCommand(syncPullCmd)
	rootCmd.AddCommand(syncCmd)
}
```

Add the package-level variable before `syncCmd`:

```go
var syncPullDryRun bool
```

**Step 2: Use the flag in syncPullCmd**

In the `syncPullCmd.RunE` function, after the `saveProfile` call (line 137), add a call to `profile.Load` if the profile was pulled and `--dry-run` was not used. But actually, `sync pull` currently just saves the profile JSON to disk — it doesn't install extensions. The dry-run flag should control whether the profile is saved to disk.

Update the `syncPullCmd.RunE` to check `syncPullDryRun`:

Before `saveProfile` (around line 134), add:
```go
			if syncPullDryRun {
				cmd.Printf("[dry-run] Would pull profile '%s' (%d extensions)\n", name, len(apiProfile.Extensions))
				pulled = append(pulled, name)
				continue
			}
```

**Step 3: Run tests**

Run: `cd agent && go build ./... && go test -v ./...`
Expected: Build succeeds, all tests pass

**Step 4: Commit**

```bash
git add agent/cmd/sync.go
git commit -m "feat(agent): add --dry-run flag to sync pull (issue #14)"
```

# VS Code Extension Detection Design

**Date:** 2026-02-04
**Issue:** #11 - Implement VS Code extension detection and listing
**Status:** Design Complete

## Overview

This design describes a comprehensive VS Code extension detection system that uses a CLI-first approach with directory parsing fallback. The system will detect extensions from both VS Code stable and Insiders installations, merge them intelligently, and extract rich metadata for user-facing features.

## Architecture & Strategy

The extension detection system uses a **CLI-first with directory fallback** approach:

### Primary Method (CLI)

- Execute `code --list-extensions --show-versions`
- Fast, reliable, officially supported
- Already implemented in current codebase

### Fallback Method (Directory Parsing)

Triggered when:
- `code` CLI is not found in PATH
- CLI command fails or returns an error
- User explicitly requests directory-based detection (via flag)

Benefits:
- Directly reads extension directories and parses manifests
- Provides resilience when CLI is unavailable
- Enables advanced features like state detection

### Execution Flow

```
┌─────────────────────────┐
│ ListExtensions() called │
└───────────┬─────────────┘
            │
            ▼
    ┌───────────────┐
    │ Try CLI first │
    └───┬───────────┘
        │
        ├─ Success ──────────► Return extensions
        │
        └─ Failure
            │
            ▼
    ┌─────────────────────┐
    │ Fall back to direct │
    │ directory parsing   │
    └──────────┬──────────┘
               │
               ▼
        Return extensions
        (with warnings logged)
```

## Extension Directory Discovery

### Extension Directory Paths

The system searches for extensions in platform-specific locations for both VS Code and Insiders:

**macOS:**
- VS Code: `~/Library/Application Support/Code/extensions`
- Insiders: `~/Library/Application Support/Code - Insiders/extensions`

**Windows:**
- VS Code: `%USERPROFILE%\.vscode\extensions`
- Insiders: `%USERPROFILE%\.vscode-insiders\extensions`

**Linux:**
- VS Code: `~/.vscode/extensions`
- Insiders: `~/.vscode-insiders/extensions`

### Extension Directory Structure

Each extension lives in a subdirectory named `publisher.name-version` (e.g., `ms-python.python-2024.0.0`) containing:
- `package.json` - Extension manifest with metadata
- Other extension files (code, resources, etc.)

### Discovery Process

1. Check all applicable paths for the current OS
2. For each existing directory, scan subdirectories
3. Each subdirectory represents a potential extension installation
4. Parse the `package.json` in each subdirectory
5. Collect extensions from all found installations
6. Skip directories without valid `package.json` files (log warning)

The discovery process runs against all paths in parallel for performance, then merges results.

## Extension Manifest Parsing

### package.json Structure

Extension manifests are JSON files containing metadata. We extract these fields:

```json
{
  "name": "python",
  "displayName": "Python",
  "description": "IntelliSense, linting, debugging...",
  "version": "2024.0.0",
  "publisher": "ms-python",
  ...
}
```

### Updated Extension Data Structure

```go
type Extension struct {
    ID          string  // "ms-python.python"
    Version     string  // "2024.0.0"
    Enabled     bool    // Global enabled state
    DisplayName string  // "Python"
    Description string  // "IntelliSense, linting..."
    Publisher   string  // "ms-python"
}
```

### Parsing Logic

1. **Read package.json** from extension directory
2. **Parse JSON** into a struct
3. **Extract ID** - Use directory name pattern `publisher.name-version` as source of truth, validate against JSON fields
4. **Extract metadata** - Get displayName, description, publisher from JSON
5. **Handle missing fields** - Use sensible defaults (empty strings for optional fields)
6. **Validation** - Ensure required fields (name, version, publisher) exist

### Error Handling

- Invalid JSON → Log warning, skip extension
- Missing package.json → Log warning, skip extension
- Missing required fields → Log warning, skip extension
- Continue processing remaining extensions

This "skip and continue" approach ensures that a few problematic extensions don't prevent detection of valid ones.

## Enabled/Disabled State Detection

### State Storage Locations

VS Code stores global extension state in platform-specific locations:

**macOS:**
- VS Code: `~/Library/Application Support/Code/User/globalStorage/storage.json`
- Insiders: `~/Library/Application Support/Code - Insiders/User/globalStorage/storage.json`

**Windows:**
- VS Code: `%APPDATA%\Code\User\globalStorage\storage.json`
- Insiders: `%APPDATA%\Code - Insiders\User\globalStorage\storage.json`

**Linux:**
- VS Code: `~/.config/Code/User/globalStorage/storage.json`
- Insiders: `~/.config/Code - Insiders/User/globalStorage/storage.json`

### State Detection Logic

1. **Locate state file** for each VS Code installation
2. **Parse JSON** to find disabled extensions list (typically under key like `extensionsIdentifiers/disabled`)
3. **Check extension ID** against disabled list
4. **Default to enabled** - If extension not in disabled list, it's enabled

### Fallback Behavior

If state file is missing, corrupted, or unreadable:
- **Assume all installed extensions are enabled**
- Log a warning that state couldn't be determined
- This is a reasonable default since most extensions are enabled

## De-duplication & Merging Logic

### Merging Strategy

When extensions are found in multiple installations (e.g., both VS Code stable and Insiders), we merge them into a single list.

### De-duplication Algorithm

```
1. Collect extensions from all installations into a map[extensionID][]Extension
2. For each extension ID with multiple entries:
   a. Compare versions using semantic version parsing
   b. Keep the extension with the highest version number
   c. Log which installation/version was chosen (for debugging)
3. Return the de-duplicated list
```

### Version Comparison

Use semantic version comparison (e.g., `2024.1.0` > `2024.0.5`):
- Parse versions into major.minor.patch components
- Compare numerically (not lexicographically)
- Handle pre-release versions appropriately (e.g., `2.0.0-beta` < `2.0.0`)

### Example Scenario

```
VS Code Stable:     ms-python.python@2024.0.0
VS Code Insiders:   ms-python.python@2024.1.0-insider
                    ms-vscode.go@0.40.0

Result (merged):    ms-python.python@2024.1.0-insider (newer)
                    ms-vscode.go@0.40.0 (only in Insiders)
```

### Implementation

We'll use Go's `golang.org/x/mod/semver` package for reliable semantic version comparison, ensuring correct ordering of version strings.

## Testing Strategy

### Unit Tests

Test individual components in isolation:

1. **Directory Discovery**
   - Path generation for each OS
   - Handling missing directories
   - Permission errors

2. **Manifest Parsing**
   - Valid package.json files
   - Invalid/corrupted JSON
   - Missing required fields
   - Missing optional fields (displayName, description)

3. **State Detection**
   - Parsing state files with disabled extensions
   - Missing state files (default to enabled)
   - Corrupted state files

4. **Version Comparison**
   - Semantic version ordering
   - Pre-release versions
   - Invalid version strings

5. **De-duplication Logic**
   - Multiple installations with same extension
   - Different versions of same extension
   - Extensions unique to one installation

### Integration Tests

1. **End-to-end with test fixtures**
   - Create mock extension directories
   - Populate with test package.json files
   - Verify correct extensions are detected

2. **CLI fallback scenario**
   - Mock CLI unavailable
   - Verify directory parsing activates

### Cross-Platform Testing

- Use build tags for OS-specific tests
- CI/CD runs tests on Windows, macOS, Linux
- Mock filesystem operations for consistent testing

### Edge Cases

- Empty extension directories
- Symlinked extension directories
- Extensions with unusual naming
- Very large numbers of extensions (performance)

## Implementation Phases

### Phase 1: Core Directory Parsing
- Implement directory discovery for all platforms
- Parse extension manifests
- Extract basic metadata (ID, version, displayName, description, publisher)

### Phase 2: State Detection
- Implement enabled/disabled state parsing
- Handle missing/corrupted state files gracefully

### Phase 3: Merging & De-duplication
- Implement version comparison
- Merge extensions from multiple installations
- De-duplicate by keeping newest versions

### Phase 4: CLI Integration
- Modify existing ListExtensions() to try CLI first
- Fall back to directory parsing on CLI failure
- Add configuration option for method preference

### Phase 5: Testing & Polish
- Comprehensive unit tests
- Integration tests with fixtures
- Cross-platform validation
- Documentation updates

## Success Criteria

- All installed extensions are detected from both VS Code and Insiders
- Extension metadata is accurate and complete
- Works on Windows, macOS, and Linux
- Gracefully handles edge cases (corrupted files, permission errors)
- Falls back to directory parsing when CLI unavailable
- Test coverage > 80%
- No performance regression compared to CLI-only approach

## Open Questions

None - design is complete and validated.

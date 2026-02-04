# Extension Profile Enhancements Design

**Date:** 2026-02-04
**Issue:** #12 - Implement extension profile save/load functionality
**Status:** Design Complete

## Overview

This design enhances the existing profile system with validation, conflict detection, and diff functionality. The core save/load features already exist; this adds safety, visibility, and user control.

## Current State

**Existing functionality:**
- ✅ Profile save command (`profile save`)
- ✅ Profile load command (`profile load`)
- ✅ Profile list command (`profile list`)
- ✅ JSON-based profile format
- ✅ Local storage in profiles directory

**Missing functionality (to be added):**
- ❌ Profile validation
- ❌ Conflict detection and handling
- ❌ Profile diff functionality

## Architecture & Approach

The enhancement adds three key capabilities to the existing profile system:

### 1. Profile Validation
- Validate profile JSON structure (required fields, correct types)
- Validate extension ID format (must be `publisher.name`)
- Return detailed error messages for validation failures
- Can be called explicitly or automatically during load

### 2. Conflict Detection & Handling
- Before installing, check which extensions are already installed
- Skip extensions that exist (no version checking to avoid downgrades)
- Report skipped extensions as warnings to the user
- Continue with remaining installations

### 3. Profile Diff
- Compare profile contents with currently installed extensions
- Show two lists: "Will install" (new) and "Already installed" (existing)
- Available as standalone `profile diff` command
- Automatically shown before load (optional flag to skip)

### Design Philosophy
- **Non-destructive**: Never uninstall or downgrade existing extensions
- **Informative**: Always tell users what will happen before doing it
- **Fail-safe**: Validate before taking action
- **Backward compatible**: Existing profiles work without changes

## Profile Validation Implementation

### Validation Function

```go
func Validate(profile *Profile) error
```

### Validation Rules

**1. Profile-level validation:**
- Name must not be empty
- Name must be valid filename (no `/`, `\`, etc.)
- Extensions array must exist (can be empty for new profiles)
- CreatedAt and UpdatedAt should be valid timestamps (warn if zero)

**2. Extension-level validation:**
- Each extension must have non-empty ID
- ID must match format: `publisher.name` (one dot, no spaces)
- Version should be non-empty (warn if missing, don't fail)
- Enabled field should be boolean (already enforced by Go type)

**3. Validation errors:**
- Return first critical error encountered
- Include field name and reason in error message
- Example: `"invalid extension ID 'bad_id': must be in format 'publisher.name'"`

### Usage
- Called automatically in `Load()` before installation
- Available as standalone function for external validation
- Can be added to `Save()` as defensive check (shouldn't fail if vscode.ListExtensions works)

### Error Handling
- Critical errors (invalid ID format) → fail fast, return error
- Warnings (missing version) → log but continue
- Empty extensions list → valid profile, just installs nothing

## Conflict Detection & Handling

### Conflict Detection Function

```go
func detectConflicts(profileExtensions []Extension, installedExtensions []vscode.Extension) (toInstall, alreadyInstalled []Extension)
```

### Algorithm

1. Get currently installed extensions via `vscode.ListExtensions()`
2. Create a map of installed extension IDs for O(1) lookup
3. For each extension in profile:
   - If ID exists in installed map → add to `alreadyInstalled` list
   - If ID not found → add to `toInstall` list
4. Return both lists

### Integration into Load()

```go
func Load(name string, profilesDir string) (*Profile, error) {
    // 1. Read and parse profile (existing)
    // 2. Validate profile (new)
    // 3. Get currently installed extensions
    // 4. Detect conflicts
    // 5. Report what will be skipped (warnings)
    // 6. Install only extensions in toInstall list
    // 7. Return summary of what was done
}
```

### User Feedback
- Before installation: "Skipping 3 already installed extensions: ext1, ext2, ext3"
- After installation: "Installed 5 extensions, skipped 3 already installed"
- Log level: INFO (not error) since this is expected behavior

### Benefits
- Avoids redundant installation attempts
- Prevents potential version downgrades
- Gives users visibility into what's happening
- Idempotent: running load multiple times is safe

## Profile Diff Implementation

### Diff Function

```go
func Diff(profileName string, profilesDir string) (*DiffResult, error)

type DiffResult struct {
    ProfileName      string
    ToInstall        []Extension  // Extensions not currently installed
    AlreadyInstalled []Extension  // Extensions already present
    TotalInProfile   int
}
```

### Implementation Steps

1. Load profile from file
2. Validate profile structure
3. Get currently installed extensions
4. Use same conflict detection logic as Load()
5. Return structured diff result

### CLI Command

```bash
devtools-sync profile diff <name>
```

### Output Format

```
Profile: work-setup (15 extensions)

Will install (10):
  - ms-python.python@2024.0.0
  - golang.go@0.40.0
  - rust-lang.rust@1.0.0
  ...

Already installed (5):
  - dbaeumer.vscode-eslint@2.4.0
  - esbenp.prettier-vscode@10.1.0
  ...

Run 'devtools-sync profile load work-setup' to install.
```

### Integration with Load

Options for integration:
1. Add `--dry-run` flag to `profile load` command - shows diff without installing
2. Show diff automatically before load, add `--yes` flag to skip confirmation
3. Keep as separate command only

**Recommendation:** Implement as separate command initially (simplest), add `--dry-run` later if needed.

### Error Cases
- Profile not found → show available profiles
- Profile invalid → show validation errors
- Can't detect installed extensions → error with helpful message

## Testing Strategy

### Unit Tests

**1. Validation Tests** (`profile_test.go`)
- Valid profile passes validation
- Invalid profile name fails
- Invalid extension ID format fails (no dot, multiple dots, spaces)
- Empty extensions list is valid
- Missing required fields fail

**2. Conflict Detection Tests**
- No conflicts when all extensions are new
- Detects all conflicts when all extensions exist
- Mixed scenario with some new, some existing
- Empty profile returns empty results

**3. Diff Tests**
- Profile not found returns error
- Invalid profile returns validation error
- Correct categorization of new vs existing extensions
- DiffResult structure is populated correctly

### Integration Tests

**1. Load with Conflicts**
- Create profile with known extensions
- Pre-install some extensions
- Verify Load skips existing and installs new ones
- Verify warnings are logged

**2. End-to-End Workflow**
- Save profile → Diff shows all will install
- Load profile → Extensions installed
- Diff again → All show as existing
- Load again → All skipped

### Test Fixtures
- Use `t.TempDir()` for test profiles directory
- Mock `vscode.ListExtensions()` to control "currently installed" state
- Sample valid and invalid profile JSON files

### Coverage Target
>80% for new functions

## Implementation Plan

### Phase 1: Validation
1. Add `Validate()` function to `profile` package
2. Add validation tests
3. Integrate into `Load()` function
4. Test with invalid profiles

### Phase 2: Conflict Detection
1. Add `detectConflicts()` helper function
2. Add conflict detection tests
3. Update `Load()` to skip existing extensions
4. Add user feedback for skipped extensions

### Phase 3: Diff Command
1. Add `Diff()` function to `profile` package
2. Add diff tests
3. Create `profile diff` CLI command
4. Format and display diff output

### Phase 4: Integration & Polish
1. Add integration tests
2. Update documentation
3. Add examples to README
4. Verify backward compatibility

## Success Criteria

- ✅ Invalid profiles are rejected with clear error messages
- ✅ Loading a profile skips already installed extensions
- ✅ Users can preview what will be installed before loading
- ✅ All tests pass with >80% coverage
- ✅ Existing profiles continue to work without modification
- ✅ CLI commands provide clear, actionable feedback

## Backward Compatibility

All changes are additive and backward compatible:
- Existing profile JSON files work without modification
- Existing `profile save` and `profile load` commands continue to work
- New validation only adds safety, doesn't change behavior for valid profiles
- Conflict handling improves UX without breaking existing workflows

## Future Enhancements (Out of Scope)

These are explicitly NOT included in this design but could be added later:
- Version conflict resolution (update/downgrade)
- Extension marketplace availability checking
- Interactive conflict resolution prompts
- Profile merging or composition
- Remote profile storage/sync

## Open Questions

None - design is complete and validated.

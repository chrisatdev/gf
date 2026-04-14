# Design: gf CLI Enhancements

## Overview

This document specifies the technical design for implementing 8 enhancements to the gf CLI tool. The changes are primarily focused on improving workflow efficiency, modernizing the configuration directory structure, and adding new capabilities.

## Architecture Decisions

### 1. Configuration Directory Change (.gf → .git/gf)

**Decision**: Update the `gfDir` constant from `.gf` to `.git/gf`

**Rationale**: 
- `.git/gf` follows the same pattern as other Git tool extensions (e.g., `.git/hooks`, `.git/info`)
- Avoids confusion with user-created directories in project root
- More professional and standardized approach

**Implementation**:
```go
// Line 61 in internal/cli/root.go
const gfDir = ".git/gf"
```

**Files to update**:
- `internal/cli/root.go`: Update constant and all path constructions

**Migration strategy**:
- On first run, check if `.gf` directory exists but `.git/gf` does not
- Copy existing configuration files to new location
- Optionally delete old `.gf` directory after successful migration

### 2. Init Enhancement (gf -i includes gf -c)

**Decision**: Modify `runInit()` to call `runConfigure()` after successful initialization

**Rationale**:
- Eliminates two-step process: `gf -i` then `gf -c`
- Provides immediate setup of main branch detection
- Better user experience for new repositories

**Implementation**:
```go
// In runInit(), after successful initialization
func runInit(runner *git.Runner) error {
    // Existing init code...

    // NEW: Call configure after init
    fmt.Printf("%s🔧 Running configuration...%s\n", cyan, nc)
    if err := runConfigure(runner); err != nil {
        fmt.Printf("%s⚠️  Warning: Configuration failed: %v%s\n", yellow, err, nc)
        // Continue - don't fail entirely
    }

    return nil
}
```

**Current behavior analysis**:
Looking at lines 516-520, `runInit()` already:
- Creates initial commit
- Detects main branch (`detectMainBranch()`)
- Saves main branch (`saveMainBranch()`)

This means init already "configures" but doesn't call `runConfigure()`. The spec requirement is to explicitly call `runConfigure()` which shows more detailed output and provides better detection logic.

### 3. Branch Creation Enhancement (Pull Before Branch)

**Decision**: Modify `runBranch()` to pull current branch before creating new branch

**Rationale**:
- Ensures new branch is based on latest code
- Reduces merge conflicts downstream
- Follows Git flow best practices

**Current behavior**:
Lines 639-650 show that `runBranch()` pulls the *main* branch, not the *current* branch.

**Implementation**:
```go
// In runBranch(), before creating new branch
func runBranch(runner *git.Runner, branchType, name string) error {
    // Existing validation...

    // NEW: Pull current branch first
    currentBranch, _ := runner.CurrentBranch()
    if currentBranch != "" && currentBranch != GetMainBranch() {
        fmt.Printf("%s🔄 Pulling current branch (%s)...%s\n", green, currentBranch, nc)
        runner.Fetch("origin")
        if _, err := runner.Command("pull", "origin", currentBranch); err != nil {
            fmt.Printf("%s⚠️  Couldn't pull from origin/%s. Continuing with local branch%s\n", 
                yellow, currentBranch, nc)
        }
    }

    // Then pull main branch (existing behavior)
    // ... existing main branch update code ...
}
```

**Error handling**:
- If pull fails, show warning but continue with branch creation
- Never fail entire operation due to pull failure

### 4. Help Unification

**Decision**: Use Kong's native help, removing custom `showHelp()` function

**Rationale**:
- Eliminates code duplication
- Kong's help is automatically updated when flags change
- Consistent behavior between `-h` and `--help`

**Implementation**:
1. Remove custom `showHelp()` function (lines 187-229)
2. Remove conditional call to `showHelp()` in `Execute()` (line 88)
3. Rely on Kong's built-in help via `kong.HelpOptions{}`

```go
// Remove or simplify showHelp()
// Delete the entire showHelp() function

// In Execute(), replace:
// if CLI.Help || len(os.Args) == 1 {
//     showHelp()
//     return nil
// }

// With Kong's default behavior which already shows help when:
// - No flags provided
// - -h or --help used
```

**Alternative if custom help is required**:
Keep `showHelp()` but make it generate output programmatically that matches Kong's format.

### 5. Main Branch Storage Path

**Decision**: Already covered by Change #1 - when `gfDir` changes to `.git/gf`, `GetMainBranch()` automatically reads from `.git/gf/main_branch`

**Implementation**: No additional code needed beyondChange #1

```go
// GetMainBranch() - already uses gfDir constant
func GetMainBranch() string {
    path := filepath.Join(gfDir, mainBranchFile)  // Now reads from .git/gf/main_branch
    data, err := os.ReadFile(path)
    if err != nil {
        return "main"
    }
    return strings.TrimSpace(string(data))
}
```

### 6. CHANGELOG Automation

**Decision**: Remove user prompt in `runAdd()`, always update CHANGELOG automatically

**Rationale**:
- Streamlines workflow
- Ensures consistent changelog entries
- No user interaction required

**Current behavior** (lines 703-736):
- Prompts user with "Update CHANGELOG.md with these changes? (y/N)"
- Only updates if user responds "y"

**Implementation**:
```go
// In runAdd(), replace prompt section
func runAdd(runner *git.Runner, files []string) error {
    // Existing staging code...

    // NEW: Always update CHANGELOG without prompting
    fmt.Printf("%s📝 Updating CHANGELOG...%s\n", cyan, nc)

    // Ensure CHANGELOG exists
    changelog.EnsureExists()

    // Get staged files info
    newFiles, modFiles, delFiles, _, _ := runner.StagedFilesByStatus()

    var allFiles []string
    allFiles = append(allFiles, newFiles...)
    allFiles = append(allFiles, modFiles...)
    allFiles = append(allFiles, delFiles...)

    if len(allFiles) > 0 {
        gen := commit.NewGenerator()
        info := gen.Generate(allFiles, "")
        shortMsg := fmt.Sprintf("%s %s: %s", info.Emoji, info.Type, info.Description)

        commitType := info.Type
        if err := changelog.AddEntry(commitType, shortMsg); err != nil {
            fmt.Printf("%s⚠️  Warning: Could not update CHANGELOG: %v%s\n", yellow, err, nc)
            // Don't fail - continue with staging
        } else {
            fmt.Printf("%s✅ CHANGELOG.md updated%s\n", green, nc)

            // Stage the changelog
            runner.Add("CHANGELOG.md")
            fmt.Printf("%s✅ CHANGELOG staged%s\n", green, nc)
        }
    } else {
        fmt.Printf("%s⚠️  No staged files to add to CHANGELOG%s\n", yellow, nc)
    }

    return nil
}
```

### 7. Push-Only Flag

**Decision**: Add `--push-only` (`-P`) flag to skip MR creation after push

**Rationale**:
- Some workflows don't need MRs (internal branches, WIP branches)
- Reduces friction for simple push workflows
- Maintains backward compatibility

**Implementation**:
1. Add new field to CLI struct:
```go
// In CLI struct, add new field
var CLI struct {
    // Existing fields...
    PushOnly bool `name:"push-only" short:"P" help:"Commit and push, but don't create MR"`
}
```

2. Modify `runCommit()` to check flag:
```go
// In runCommit(), around line 799
// Modify the MR creation section
func runCommit(runner *git.Runner, message string) error {
    // Existing code...

    // Push
    fmt.Printf("%s📤 Pushing to %s%s%s\n", green, cyan, currentBranch, nc)
    if err := runner.Push(currentBranch, true); err != nil {
        fmt.Printf("%s❌ Push failed: %v%s\n", red, err, nc)
        fmt.Printf("%s⚠️  If conflicts exist, run: %sgf -m%s\n", yellow, cyan, nc)
        return nil
    }
    fmt.Printf("%s✅ Push successful%s\n", green, nc)

    // NEW: Only create MR if push-only flag is NOT set
    if CLI.PushOnly {
        fmt.Printf("%s🔹 Push-only mode: MR not created%s\n", cyan, nc)
    } else {
        // Existing MR creation logic
        if !isMain {
            url, _ := mr.GetMRURL(runner, currentBranch)
            if url != "" {
                fmt.Printf("%s🔗 Opening Merge Request...%s\n", cyan, nc)
                openBrowser(url)
            }
        } else {
            fmt.Printf("%s⚠️  No MR will be created for %s branch%s\n", yellow, mainBranch, nc)
        }
    }

    return nil
}
```

3. Add flag to detection in Execute():
```go
// In Execute(), detect push-only flag
if CLI.Commit || CLI.PushOnly || flagProvided("-p") {
    msg := ""
    if len(CLI.Args) > 0 {
        msg = CLI.Args[0]
    }
    return runCommit(runner, msg)
}
```

### 8. Error Handling Verification

**Decision**: Review and enhance error handling across all modified functions

**Areas to verify**:

| Function | Current Behavior | Required Fix |
|---------|-----------------|-------------|
| `runInit()` | Returns error on failure | Already handles - ensure clear messages |
| `runConfigure()` | Returns error on failure | Already handles |
| `runBranch()` | Returns error on failure | Already handles |
| `runAdd()` | Returns error on staging only | Add CHANGELOG error handling (see Change #6) |
| `runCommit()` | Returns error on failure | Already handles |
| Path operations | Use `os.MkdirAll` | Verify proper error propagation |

**Specific fixes needed**:
1. Empty branch name validation (line 635-636) - already exists
2. No staged files check (lines 754-762) - already exists
3. CHANGELOG write failures - will be handled by Change #6

## Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| github.com/alecthomas/kong | ^0.6.0 | CLI parsing |
| github.com/chrisatdev/gf/internal/changelog | local | CHANGELOG updates |
| github.com/chrisatdev/gf/internal/commit | local | Commit message generation |
| github.com/chrisatdev/gf/internal/git | local | Git operations |
| github.com/chrisatdev/gf/internal/mr | local | MR URL generation |

**No new external dependencies required**

## File Changes Summary

| File | Change Type | Lines Affected |
|------|------------|----------------|
| `internal/cli/root.go` | Modified | Multiple |

### Detailed Line Changes

| Line(s) | Change | Description |
|---------|--------|-------------|
| 61 | Modified | `gfDir` constant changed from `.gf` to `.git/gf` |
| 48 | Modified | Add `PushOnly bool` field to CLI struct |
| 88-90 | Modified | Remove custom help call, rely on Kong |
| 187-229 | Deleted | Custom `showHelp()` function |
| 490-525 | Modified | Add call to `runConfigure()` in `runInit()` |
| 527-565 | Modified | Ensure uses new `gfDir` |
| 595-605 | Modified | `saveMainBranch()` uses new `gfDir` |
| 607-615 | Modified | `GetMainBranch()` uses new `gfDir` |
| 627-664 | Modified | `runBranch()` pulls current branch before creating new branch |
| 681-739 | Modified | Remove CHANGELOG prompt in `runAdd()`, auto-update |
| 741-810 | Modified | `runCommit()` checks `CLI.PushOnly` for MR skip |

## Testing Strategy

### Unit Tests
- Test `GetMainBranch()` reads from correct path
- Test `saveMainBranch()` writes to correct path
- Test `runAdd()` updates CHANGELOG without prompt

### Integration Tests
- Test `gf -i` creates `.git/gf/main_branch` (not `.gf`)
- Test `gf -s -f name` pulls current branch before creating feature branch
- Test `gf -p -P "msg"` commits and pushes without MR
- Test `gf -a` auto-updates CHANGELOG

### Manual Verification
- Test help output matches between `-h` and `--help`
- Test error messages are clear and helpful

## Rollback Plan

1. Revert `gfDir` constant to `.gf`
2. Remove `runConfigure()` call from `runInit()`
3. Remove pull current branch logic from `runBranch()`
4. Restore custom `showHelp()` function
5. Restore user prompt in `runAdd()`
6. Remove `PushOnly` field and associated logic
7. Verify with existing test suite

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Breaking existing users who have `.gf` config | Medium | Low | Implement automatic migration |
| Help output different after unification | Low | Medium | Keep custom help if needed |
| CHANGELOG update failures | Low | Low | Warning only, don't fail entire operation |
| Push-only flag not recognized | Low | Medium | Test both `-P` and `--push-only` |

## Success Criteria

- [ ] `gf -i` successfully initializes and configures repository
- [ ] All configuration stored in `.git/gf/` not `.gf/`
- [ ] `gf -s -f name` pulls current branch before creating new branch
- [ ] `gf -h` and `gf --help` produce identical output
- [ ] `gf -a` automatically updates CHANGELOG without prompting
- [ ] `gf -p -P "msg"` commits and pushes without opening MR
- [ ] `gf -p "msg"` continues to create MR as before
- [ ] Error messages are clear and actionable
- [ ] No regressions in existing functionality
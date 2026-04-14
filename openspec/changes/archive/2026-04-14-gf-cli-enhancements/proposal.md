# Proposal: gf CLI Enhancements

## Intent

Enhance the gf CLI tool to improve user experience and workflow efficiency by:
1. Making `gf -i` automatically run configuration like `gf -c`
2. Moving configuration from `.gf` to `.git/gf` for better Git integration
3. Having branch creation commands (`gf -s -f/-h/-b/-r`) pull the current branch first
4. Unifying help output between custom `showHelp()` and Kong's native help
5. Using `.git/gf/main_branch` for storing main branch in `gf -m` operations
6. Automating CHANGELOG updates in `gf -a` without user prompts
7. Adding a `--push-only` flag to commit/push without creating MRs
8. Fixing identified bugs and inconsistencies

## Scope

### In Scope
- Modify `runInit()` to call `runConfigure()` after initialization
- Change `gfDir` constant from ".gf" to ".git/gf"
- Update `runBranch()` to pull current branch before creating new branches
- Unify help text in `showHelp()` to match Kong's help output
- Modify `GetMainBranch()` to read from `.git/gf/main_branch`
- Update `runAdd()` to automatically update CHANGELOG without prompting
- Add `--push-only` (`-P`) flag to `runCommit()` to skip MR creation
- Identify and fix bugs in CLI functionality

### Out of Scope
- Major architectural changes to the CLI framework
- Adding new version control system support (beyond Git)
- Changing the underlying Git command execution approach
- Adding interactive TUI features beyond current scope

## Capabilities

### New Capabilities
- `push-only`: Commit and push changes without creating merge requests

### Modified Capabilities
- `init`: Now includes automatic configuration (was initialization only)
- `configure`: Now uses `.git/gf` directory instead of `.gf`
- `branch`: Now pulls current branch before creating new branches
- `help`: Now provides unified help matching Kong's output
- `main-branch`: Now reads from `.git/gf/main_branch` instead of `.gf/main_branch`
- `add`: Now automatically updates CHANGELOG (was optional prompt)
- `commit`: Now supports `--push-only` flag to skip MR creation

## Approach

1. **Configuration Directory Change**: Update `gfDir` constant and all references to use `.git/gf` instead of `.gf`
2. **Init Enhancement**: Modify `runInit()` to call `runConfigure()` after successful initialization
3. **Branch Enhancement**: Update `runBranch()` to fetch and pull the current branch before creating new branches
4. **Help Unification**: Align `showHelp()` output with Kong's automatic help generation
5. **Main Branch Storage**: Update path handling in `saveMainBranch()` and `GetMainBranch()` functions
6. **CHANGELOG Automation**: Remove user prompt in `runAdd()` and always update CHANGELOG when changes are staged
7. **Push-only Flag**: Add new CLI flag and modify `runCommit()` to conditionally skip MR creation
8. **Bug Fixes**: Review and fix any identified issues in the codebase

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/cli/root.go` | Modified | Contains all CLI command implementations and constants |
| `gfDir` constant (line 61) | Modified | Change from ".gf" to ".git/gf" |
| `mainBranchFile` constant (line 64) | Modified | Update path construction for new directory |
| `runInit()` function (lines 490-525) | Modified | Add call to `runConfigure()` |
| `runConfigure()` function (lines 527-565) | Modified | Update directory path usage |
| `saveMainBranch()` function (lines 595-605) | Modified | Update directory path usage |
| `GetMainBranch()` function (lines 607-615) | Modified | Update directory path usage |
| `runBranch()` function (lines 627-664) | Modified | Add pull of current branch |
| `showHelp()` function (lines 187-229) | Modified | Unify with Kong help output |
| `runAdd()` function (lines 681-739) | Modified | Remove prompt, always update CHANGELOG |
| CLI struct (lines 27-51) | Modified | Add `--push-only`/`-P` flag |
| `runCommit()` function (lines 741-810) | Modified | Add logic to skip MR when push-only flag is set |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Breaking existing user workflows | Medium | Maintain backward compatibility where possible; provide clear documentation |
| Issues with `.git/gf` directory permissions | Low | Ensure proper directory creation with appropriate permissions |
| Conflicts with Git's internal `.git` directory usage | Low | Use `.git/gf` which is a standard practice for Git tool extensions |
| Forgetting to update all references to old `.gf` directory | Medium | Use comprehensive search and replace; verify with tests |
| Help text divergence between custom and Kong versions | Low | Implement unified help function that serves both purposes |

## Rollback Plan

1. Revert `gfDir` constant to ".gf"
2. Remove call to `runConfigure()` from `runInit()`
3. Revert `runBranch()` to not pull current branch
4. Restore original `showHelp()` implementation
5. Revert path changes in main branch functions to use `.gf` directory
6. Restore user prompt in `runAdd()` for CHANGELOG updates
7. Remove `--push-only` flag and associated logic
8. Verify all changes by comparing with backup or git history

## Dependencies

- None beyond standard Go libraries and existing dependencies

## Success Criteria

- [ ] `gf -i` successfully initializes repository and configures main branch
- [ ] All gf configuration is stored in `.git/gf` directory instead of `.gf`
- [ ] `gf -s -f/<name>` (and similar) pulls current branch before creating new branch
- [ ] `gf -h` produces identical output to `gf --help` (Kong's help)
- [ ] `gf -m` reads main branch from `.git/gf/main_branch`
- [ ] `gf -a` automatically updates CHANGELOG without prompting
- [ ] `gf -p -P` (or `gf -p --push-only`) commits and pushes without creating MR
- [ ] All existing functionality continues to work as expected
- [ ] No regressions in existing CLI command behavior
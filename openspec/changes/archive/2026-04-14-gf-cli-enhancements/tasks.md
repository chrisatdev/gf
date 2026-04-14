# Tasks: gf CLI 8 Changes

## Phase 1: Infrastructure & Configuration

- [ ] 1.1 Update `internal/cli/root.go` line 61: change `gfDir` constant from `.gf` to `.git/gf`
- [ ] 1.2 Add migration logic in `initConfig()` to copy from `.gf` to `.git/gf` if needed
- [ ] 1.3 Add `PushOnly bool` field to CLI struct (line ~48) with `name:"push-only" short:"P"`

## Phase 2: Init Enhancement

- [ ] 2.1 Modify `runInit()` (line ~490) to call `runConfigure()` after successful init
- [ ] 2.2 Add warning handling - don't fail entire init if configure fails

## Phase 3: Branch Creation Enhancement

- [ ] 3.1 Modify `runBranch()` (line ~627) to pull current branch before creating new branch
- [ ] 3.2 Add fetch and pull for current branch (not main) before branch creation
- [ ] 3.3 Handle pull failure gracefully - warn but continue

## Phase 4: Help Unification

- [ ] 4.1 Remove custom `showHelp()` function (lines 187-229)
- [ ] 4.2 Remove conditional help call in `Execute()` (line 88-90)
- [ ] 4.3 Configure Kong's built-in help via `kong.HelpOptions{}`

## Phase 5: CHANGELOG Automation

- [ ] 5.1 Modify `runAdd()` (lines 681-739) to remove user prompt
- [ ] 5.2 Add automatic CHANGELOG update without prompting
- [ ] 5.3 Ensure CHANGELOG.md is staged after update

## Phase 6: Push-Only Flag

- [ ] 6.1 Modify `runCommit()` (line ~741) to check `CLI.PushOnly` flag
- [ ] 6.2 Skip MR creation when PushOnly is true
- [ ] 6.3 Add flag detection in `Execute()` around line 269

## Phase 7: Error Handling

- [ ] 7.1 Review and enhance error messages in all modified functions
- [ ] 7.2 Add empty branch name validation in `runBranch()`
- [ ] 7.3 Add no staged files check in `runAdd()`

## Phase 8: Testing

- [ ] 8.1 Write unit test for `GetMainBranch()` reading from `.git/gf/main_branch`
- [ ] 8.2 Write integration test for `gf -i` creating `.git/gf/` config
- [ ] 8.3 Write integration test for `gf -s -f name` pulling current branch
- [ ] 8.4 Write integration test for `gf -p -P "msg"` push without MR
- [ ] 8.5 Write integration test for `gf -a` auto-updating CHANGELOG
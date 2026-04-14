# Spec: gf CLI Enhancements

## 1. Make `gf -i` automatically run configuration like `gf -c`

### Requirements:
- When `gf -i` is executed, after successful repository initialization, it should automatically call the configuration function
- The configuration should detect and save the main branch (main/master/develop/dev) to `.git/gf/main_branch`
- User should not need to run `gf -c` separately after `gf -i`

### Scenarios:
#### 1.1 Successful initialization with configuration
**Given** a directory that is not a git repository
**When** the user runs `gf -i`
**Then** the repository is initialized with an initial commit
**And** the main branch is detected and saved to `.git/gf/main_branch`
**And** the user sees success messages for both initialization and configuration

#### 1.2 Initialization in existing repository
**Given** a directory that is already a git repository
**When** the user runs `gf -i`
**Then** the command shows a warning that it's already a git repository
**And** no configuration is performed

#### 1.3 Configuration failure handling
**Given** a newly initialized repository where main branch detection fails
**When** the user runs `gf -i`
**Then** the repository is initialized successfully
**And** a warning is shown about configuration failure
**And** the command returns successfully (does not fail entirely)

## 2. Change configuration directory from `.gf` to `.git/gf`

### Requirements:
- All references to the `.gf` directory should be changed to `.git/gf`
- The `gfDir` constant should be updated from ".gf" to ".git/gf"
- All functions that create, read, or write to the gf directory should use the new path
- Existing `.gf` directories should be migrated to `.git/gf` (optional enhancement)

### Scenarios:
#### 2.1 Directory path usage
**Given** any gf command that uses the configuration directory
**When** the command executes
**Then** all file operations use `.git/gf` instead of `.gf`
**Example**: `saveMainBranch()` writes to `.git/gf/main_branch`

#### 2.2 Directory creation
**Given** a repository without a `.git/gf` directory
**When** a command requires the configuration directory (e.g., `gf -c`)
**Then** the `.git/gf` directory is created with appropriate permissions (0755)

#### 2.3 File operations
**Given** a repository with a `.git/gf` directory containing configuration
**When** commands read configuration (e.g., `GetMainBranch()`)
**Then** files are read from `.git/gf` instead of `.gf`

## 3. Pull current branch in branch creation commands (`gf -s [-f,-h,-b,-r]`)

### Requirements:
- Before creating a new branch, the current branch should be pulled from origin
- This ensures the local branch is up-to-date before branching
- Should handle cases where pull fails gracefully (show warning but continue)

### Scenarios:
#### 3.1 Successful pull before branch creation
**Given** a repository where the current branch can be pulled from origin
**When** the user runs `gf -s -f feature-name`
**Then** the current branch is checked out
**And** `git fetch origin` is executed
**And** `git pull origin current-branch` is executed
**And** the new feature branch is created from the updated current branch

#### 3.2 Pull failure handling
**Given** a repository where pulling from origin fails (e.g., no upstream)
**When** the user runs `gf -s -f feature-name`
**Then** a warning is shown about the pull failure
**And** the branch creation continues using the local branch
**And** the new branch is still created successfully

#### 3.3 All branch types
**Given** a repository with any current branch
**When** the user runs `gf -s -f name`, `gf -s -h name`, `gf -s -b name`, or `gf -s -r name`
**Then** the current branch is pulled before creating the respective branch type

## 4. Unify help output between custom `showHelp()` and Kong's native help

### Requirements:
- The custom `showHelp()` function should produce identical output to `gf --help` (Kong's help)
- Remove duplication of help text maintenance
- Ensure all options and examples are accurately represented

### Scenarios:
#### 4.1 Identical help output
**Given** the gf CLI
**When** the user runs `gf -h`
**Then** the output is identical to running `gf --help`
**When** the user runs `gf --help`
**Then** the output is identical to running `gf -h`

#### 4.2 Complete option coverage
**Given** the unified help function
**When** help is displayed
**Then** all CLI flags are documented: `-v`, `--update`, `-i`, `-c`, `-s`, `-f`, `-h`, `-b`, `-r`, `-a`, `-p`, `-m`, `-F`, `-M`, `-t`, `-w`, `-?`
**Then** all positional args are documented
**Then** all examples are included and accurate

#### 4.3 Help when no arguments
**Given** the gf CLI invoked with no arguments
**When** the user runs `gf` (no flags)
**Then** the unified help is displayed (same as `gf -h`)

## 5. Use `.git/gf/main_branch` for storing main branch in `gf -m` operations

### Requirements:
- The `GetMainBranch()` function should read from `.git/gf/main_branch`
- The `saveMainBranch()` function should write to `.git/gf/main_branch`
- All references to the old `.gf/main_branch` path should be updated

### Scenarios:
#### 5.1 Reading main branch
**Given** a repository with `.git/gf/main_branch` containing "main"
**When** `GetMainBranch()` is called (e.g., during `gf -m`)
**Then** the function returns "main"

#### 5.2 Saving main branch
**Given** a repository without `.git/gf` directory
**When** `saveMainBranch("develop")` is called (e.g., during `gf -c`)
**Then** the `.git/gf` directory is created
**And** `.git/gf/main_branch` is created with content "develop"

#### 5.3 Fallback behavior
**Given** a repository without `.git/gf/main_branch`
**When** `GetMainBranch()` is called
**Then** the function returns the default "main" branch
**And** no error is returned

## 6. Automate CHANGELOG updates in `gf -a` without user prompts

### Requirements:
- Remove the interactive prompt asking if user wants to update CHANGELOG
- Always update CHANGELOG when changes are staged via `gf -a`
- Maintain the same CHANGELOG update logic (generate commit message from staged files)

### Scenarios:
#### 6.1 Automatic CHANGELOG update
**Given** a repository with staged changes
**When** the user runs `gf -a`
**Then** the changes are staged
**And** the CHANGELOG is automatically updated with a generated entry
**And** the CHANGELOG file is staged for commit
**And** no user prompt is displayed

#### 6.2 CHANGELOG update with specific files
**Given** a repository with specific files staged via `gf -a file1.txt file2.txt`
**When** the command executes
**Then** only the specified files are staged
**And** the CHANGELOG is updated based on those files
**And** the CHANGELOG file is staged

#### 6.3 CHANGELOG update failure handling
**Given** a repository where CHANGELOG update fails (e.g., permission issues)
**When** the user runs `gf -a`
**Then** the changes are staged successfully
**And** a warning is shown about the CHANGELOG update failure
**And** the command returns successfully (does not fail entirely)

## 7. Add `--push-only` flag to commit/push without creating MRs

### Requirements:
- Add a new `--push-only` (`-P`) flag to the CLI structure
- Modify `runCommit()` to conditionally skip MR creation when the flag is set
- When flag is set, the command should commit, push, but not open MR URL
- When flag is not set, existing behavior is preserved (create MR for non-main branches)

### Scenarios:
#### 7.1 Push-only flag usage
**Given** a repository with staged changes on a feature branch
**When** the user runs `gf -p -P "commit message"` or `gf -p --push-only "commit message"`
**Then** the changes are committed with the provided message
**And** the changes are pushed to the remote
**And** no MR is created or opened
**And** the command completes successfully

#### 7.2 Push-only without commit message
**Given** a repository with staged changes
**When** the user runs `gf -p -P` (no message)
**Then** an auto-generated commit message is used
**And** the changes are committed and pushed
**And** no MR is created or opened

#### 7.3 Default behavior preserved
**Given** a repository with staged changes on a feature branch
**When** the user runs `gf -p "commit message"` (without push-only flag)
**Then** the changes are committed and pushed
**And** an MR is created and opened in the browser
**When** the user runs `gf -p` on main branch
**Then** the changes are committed and pushed
**And** no MR is created (existing behavior for main branch)

#### 7.4 Flag alias recognition
**Given** the push-only functionality
**When** the user uses either `-P` or `--push-only`
**Then** both forms work identically

## 8. Verify and fix errors

### Requirements:
- Review the entire codebase for potential bugs and inconsistencies
- Fix any identified issues related to the changes being made
- Ensure all modified functions handle errors appropriately
- Verify that error messages are clear and helpful

### Scenarios:
#### 8.1 Error handling in init
**Given** a directory where git init fails
**When** the user runs `gf -i`
**Then** an appropriate error message is shown
**And** the command returns a non-zero exit code

#### 8.2 Error handling in branch creation
**Given** invalid parameters to branch creation
**When** the user runs `gf -s -f ""` (empty name)
**Then** an error is returned indicating branch name is required

#### 8.3 Error handling in CHANGELOG update
**Given** a repository where CHANGELOG file cannot be written
**When** the user runs `gf -a`
**Then** changes are staged successfully
**And** a warning is shown about CHANGELOG failure
**And** command completes successfully

#### 8.4 Error handling in push-only
**Given** a repository where push fails (e.g., no remote)
**When** the user runs `gf -p -P`
**Then** an appropriate error message is shown about push failure
**And** command returns error (does not continue to MR step since it's skipped anyway)

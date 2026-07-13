# gf

A streamlined Git workflow CLI for GitHub and GitLab projects.

`gf` wraps the day-to-day git operations — branching, committing, syncing, pushing, and conflict resolution — behind a consistent set of short flags and subcommands.

## Install

**Linux and macOS:**

```bash
curl -fsSL https://raw.githubusercontent.com/chrisatdev/gf/main/install.sh | bash
```

Or download a specific release manually from the [releases page](https://github.com/chrisatdev/gf/releases).

> **Windows**: use [WSL 2](https://learn.microsoft.com/en-us/windows/wsl/install). Native Windows binaries are not provided.

## Usage

### Initialize a repo

```bash
gf -i
```

Detects the remote (GitHub or GitLab), prompts for the main branch name, and writes `.git/gf.toml`.

### Start a branch

```bash
gf -s feat user-auth        # creates feat/user-auth
gf -s fix  login-redirect   # creates fix/login-redirect
gf -s                       # list available branch types
```

### Stage files

```bash
gf -a                       # git add --all
gf -a src/handler.go        # stage specific file
```

### Commit, update changelog, and push

```bash
gf -p                       # auto-infer commit message from branch name
gf -p -M "feat(auth): add JWT validation"
gf -p --only-push           # push without creating a PR/MR
```

### Interactive commit wizard

```bash
gf commit
```

Steps through type → scope → breaking change → description → preview, then runs `git commit`.

### Sync with main

```bash
gf sync
```

Fetches origin, merges the main branch into the current branch, and prints ahead/behind counts. On conflicts, lists affected files and suggests `gf resolve`.

### Resolve merge conflicts

```bash
gf resolve
```

Detects conflicted files. `CHANGELOG.md` is resolved automatically. Other files are presented one at a time with `[o]urs / [t]heirs / [e]dit / [s]kip` options.

### Other flags

```bash
gf -S          # git status
gf -v          # print version
gf -c          # print current gf config
gf -w          # interactively switch branches
gf -m          # merge origin/main into current branch (--no-commit --no-ff)
gf -f          # delete current branch locally and remotely
gf -u          # self-update to the latest release
```

## Platform support

| Platform | Support |
|---|---|
| Linux (amd64, arm64) | ✅ |
| macOS (amd64, arm64) | ✅ |
| Windows | WSL 2 recommended |

Both **GitHub** and **GitLab** are supported. `gf -p` creates a GitHub compare URL (opens browser) or a GitLab MR via push options, depending on the configured platform.

## Configuration

`gf` stores its config at `.git/gf.toml` inside the repository (not tracked by git). Run `gf -i` to generate it, or edit it directly:

```toml
[repo]
platform     = "github"   # github | gitlab
main_branch  = "main"
project_path = "owner/repo"

[flow]
mfa_active = false
```

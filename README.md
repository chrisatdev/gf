# Git Flow Enhanced (gf) 🚀

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Version](https://img.shields.io/badge/Version-1.1.0-blue)
![Bash](https://img.shields.io/badge/Bash-5.0%2B-brightgreen)

A powerful Git workflow automation tool with semantic commits, Gitmoji support, and advanced features.

**Author**: [Christian Benítez](https://github.com/chrisatdev)

## ✨ Features

- 🆕 **Project Initialization**: `gf -i`
- 🌿 **Branch Management**:
  - Create feature branches: `gf -s -f feature-name`
  - Create hotfix branches: `gf -s -h hotfix-name`
  - Create release branches: `gf -s -r release-name`
- 💾 **Smart Commits**:
  - Auto-generated semantic commit messages
  - Gitmoji support
  - Detailed file change tracking
- 📦 **Staging Changes**: `gf -a` or `gf -a file1 file2`
- 📤 **Push & Create MR/PR**: `gf -p "commit message"`
- 🔀 **Merge Handling**: `gf -m`
- 🗑️ **Branch Cleanup**: `gf -f`
- 🔄 **Cross-branch MR Creation** (GitLab): `gf -r source target`
- 📜 **Automatic CHANGELOG.md** generation and updates

## 🚀 Installation

1. Download the script:

```bash
curl -o gf https://raw.githubusercontent.com/chrisatdev/gf/main/gf.sh
```

2. Make it executable:

```bash
chmod +x gf
```

3. Move to your PATH:

```bash
sudo mv gf /usr/local/bin/
```

## 📚 Basic Usage

### Initialize a new repository

```bash
gf -i
```

### Create a feature branch

```bash
gf -s -f awesome-feature
```

### Stage all changes

```bash
gf -a
```

### Commit with auto-generated message

```bash
gf -p
```

### Or with custom message

```bash
gf -p "feat: add new payment processor"
```

### Merge main into current branch

```bash
gf -m
```

### Delete current branch

```bash
gf -f
```

### Create MR from main to dev (GitLab)

```bash
gf -r main dev
```

## 🌈 Gitmoji Support

Automatically adds appropriate emojis based on commit type:

| Type     | Emoji | Description              |
| -------- | ----- | ------------------------ |
| feat     | ✨    | New feature              |
| fix      | 🐛    | Bug fix                  |
| docs     | 📝    | Documentation            |
| style    | 💄    | Code style               |
| refactor | ♻️    | Refactoring              |
| test     | ✅    | Testing                  |
| chore    | 🔧    | Maintenance              |
| build    | 👷    | Build system             |
| ci       | ⚙️    | CI configuration         |
| perf     | ⚡    | Performance improvements |
| revert   | ⏪    | Revert changes           |

## 📜 CHANGELOG.md Automation

The tool automatically:

- Creates `CHANGELOG.md` on first commit
- Updates it with each new commit
- Organizes changes by type (Added, Fixed, Changed)

Example structure:

```markdown
# CHANGELOG

## [Unreleased]

### Added

- New payment processor integration
- User authentication service

### Fixed

- Login timeout issue
```

## 🛠️ Requirements

- Bash 5.0+
- Git 2.20+
- GitLab repository (for MR features)

## 📝 License

MIT License - Copyright (c) 2023 Christian Benítez

## 🤝 Contributing

Feel free to submit issues or PRs at:  
[https://github.com/chrisatdev/gf](https://github.com/chrisatdev/gf)

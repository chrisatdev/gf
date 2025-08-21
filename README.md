# Git Flow Enhanced (gf) 🚀

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Version](https://img.shields.io/badge/Version-1.3.0-blue)
![Bash](https://img.shields.io/badge/Bash-5.0%2B-brightgreen)

A powerful Git workflow automation tool with semantic commits, Gitmoji support, automatic changelog management, and advanced features.

**Author**: [Christian Benítez](https://github.com/chrisatdev)

## ✨ Features

- 🆕 **Project Initialization**: `gf -i`
- 🌿 **Branch Management**:
  - Create feature branches: `gf -s -f feature-name`
  - Create hotfix branches: `gf -s -h hotfix-name`
  - Create bugfix branches: `gf -s -b bugfix-name`
  - Create release branches: `gf -s -r release-name`
- 💾 **Smart Commits**:
  - Auto-generated semantic commit messages with file count
  - Gitmoji support
  - Detailed file change tracking (new, modified, deleted, renamed)
  - English commit messages with GitLab user attribution
- 📦 **Staging Changes**: `gf -a` or `gf -a file1 file2`
- 📤 **Push & Create MR/PR**: `gf -p "commit message"`
- 🔀 **Merge Handling**: `gf -m`
- 🗑️ **Branch Cleanup**: `gf -f`
- 🔄 **Cross-branch MR Creation** (GitLab): `gf -r source target`
- 📜 **Advanced CHANGELOG.md Management**:
  - Automatic generation and updates
  - Monthly rotation and archiving
  - Auto-cleanup of old archives (6+ months)
  - Enhanced format with user attribution

## 🚀 Installation

### Linux (Ubuntu/Debian)

```bash
# Method 1: Direct download
curl -o gf https://raw.githubusercontent.com/chrisatdev/gf/main/gf
chmod +x gf
sudo mv gf /usr/local/bin/

# Method 2: Using wget
wget https://raw.githubusercontent.com/chrisatdev/gf/main/gf -O gf
chmod +x gf
sudo mv gf /usr/local/bin/
```

### Linux (RHEL/CentOS/Fedora)

```bash
# Download and install
curl -o gf https://raw.githubusercontent.com/chrisatdev/gf/main/gf
chmod +x gf
sudo mv gf /usr/local/bin/

# Alternative: Install to user directory
mkdir -p ~/.local/bin
mv gf ~/.local/bin/
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### macOS

```bash
# Method 1: Using curl
curl -o gf https://raw.githubusercontent.com/chrisatdev/gf/main/gf
chmod +x gf
sudo mv gf /usr/local/bin/

# Method 2: Install to user directory
mkdir -p ~/bin
mv gf ~/bin/
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc  # or ~/.bash_profile
source ~/.zshrc  # or source ~/.bash_profile
```

### Verify Installation

```bash
gf -h
```

You should see the help message with version 1.3.0.

## 📚 Basic Usage

### Initialize a new repository

```bash
gf -i
```

### Create different types of branches

```bash
# Feature branch
gf -s -f awesome-feature

# Hotfix branch
gf -s -h critical-bug-fix

# Bugfix branch
gf -s -b login-issue

# Release branch
gf -s -r v2.1.0
```

### Stage changes

```bash
# Stage all changes
gf -a

# Stage specific files
gf -a src/app.js package.json
```

### Commit with enhanced messages

```bash
# Auto-generated semantic message with file count and details
gf -p

# Custom message (gitmoji added automatically)
gf -p "feat: add new payment processor"
```

**Example auto-generated commit:**

```
✨ feat: add new features (3 files)

Changes: 2 new, 1 modified

**New files:**
- src/components/PaymentForm.tsx
- src/utils/payment.ts

**Modified files:**
- package.json
```

### Merge and cleanup

```bash
# Merge main into current branch
gf -m

# Delete current branch and return to main
gf -f
```

### Create Merge Requests (GitLab)

```bash
# Create MR from current branch to main
gf -p "feat: new feature"  # Automatically opens MR

# Create MR between specific branches
gf -r feature/payment main
```

## ✨ Gitmoji Support

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

## 📜 Advanced CHANGELOG.md Management

### Features:

- **Monthly Rotation**: Automatically archives changelogs from previous months
- **Smart Archiving**: Creates `changelogs/` directory and archives with format `CHANGELOG-YYYY-MM.md`
- **Auto-cleanup**: Removes changelog archives older than 6 months
- **Enhanced Format**: Includes GitLab user attribution

### Automatic Structure:

```
project/
├── CHANGELOG.md              # Current month
├── changelogs/              # Historical archives
│   ├── CHANGELOG-2024-07.md # July 2024
│   └── CHANGELOG-2024-08.md # August 2024
└── ...
```

### Example CHANGELOG.md Format:

```markdown
# CHANGELOG

## [Unreleased] - September 2024

### Added

- [feat] add new payment processor (3 files) by @yourusername
- [feat] implement user authentication service (5 files) by @yourusername

### Fixed

- [fix] resolve login timeout issue (2 files) by @yourusername

### Documentation

- [docs] update API documentation (1 file) by @yourusername
```

### Monthly Rotation Behavior:

1. **Script detects** if current `CHANGELOG.md` is from previous month
2. **Archives automatically** to `changelogs/CHANGELOG-YYYY-MM.md`
3. **Creates new** `CHANGELOG.md` for current month
4. **Cleans up** archives older than 6 months

## 🎯 Workflow Examples

### Feature Development Workflow

```bash
# 1. Create feature branch
gf -s -f user-profile

# 2. Make changes, then stage
gf -a

# 3. Commit and push (opens MR automatically)
gf -p

# 4. Merge main updates if needed
gf -m

# 5. When done, cleanup branch
gf -f
```

### Hotfix Workflow

```bash
# 1. Create hotfix branch
gf -s -h critical-security-fix

# 2. Make fixes, stage and commit
gf -a
gf -p "fix: resolve security vulnerability in auth"

# 3. Cleanup
gf -f
```

## 🛠️ Requirements

- **Bash**: 5.0+
- **Git**: 2.20+
- **System**: Linux/macOS
- **Remote**: GitLab or GitHub repository
- **Optional**: `xdg-open` (Linux) or `open` (macOS) for automatic MR/PR opening

## 🔧 Configuration

The script automatically detects:

- **GitLab username** from git config or remote URL
- **Repository type** (GitLab/GitHub) for MR/PR creation
- **Operating system** for proper date handling

### Manual Git Configuration (recommended):

```bash
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

## 🐛 Troubleshooting

### Common Issues:

**Permission denied:**

```bash
chmod +x gf
```

**Command not found:**

```bash
# Ensure the binary is in your PATH
echo $PATH
which gf
```

**Date command issues (macOS):**

```bash
# Install GNU coreutils if needed
brew install coreutils
```

## 📝 License

MIT License - Copyright (c) 2025 Christian Benítez

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.

**Repository**: [https://github.com/chrisatdev/gf](https://github.com/chrisatdev/gf)

### Development:

1. Fork the repository
2. Create a feature branch: `gf -s -f new-feature`
3. Make changes and test thoroughly
4. Submit a pull request

## 🔄 Version History

- **v1.3.0**: Monthly changelog rotation, enhanced commit messages, auto-cleanup
- **v1.2.0**: Improved changelog format, GitLab user attribution
- **v1.1.2**: Enhanced semantic commits and Gitmoji support
- **v1.0.0**: Initial release with basic Git workflow automation

---

**Made with ❤️ by [Christian Benítez](https://github.com/chrisatdev)**

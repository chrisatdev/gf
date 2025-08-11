#!/bin/bash
# -------------------------------------------------------------------
# Git Flow Enhanced (gf)
# Version: 1.1.3
# Author: Christian Benítez
# GitHub: https://github.com/chrisatdev
# Description: Advanced Git workflow automation tool
# -------------------------------------------------------------------

# Colors for better output representation
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Emojis for gitmoji integration
EMOJI_FEAT="✨"
EMOJI_FIX="🐛"
EMOJI_DOCS="📝"
EMOJI_STYLE="💄"
EMOJI_REFACTOR="♻️"
EMOJI_TEST="✅"
EMOJI_CHORE="🔧"
EMOJI_HOTFIX="🚨"
EMOJI_INIT="🎉"
EMOJI_MERGE="🔀"
EMOJI_DELETE="🗑️"
EMOJI_BUILD="👷"
EMOJI_PERF="⚡"

# Gitmoji mapping
declare -A GITMOJI=(
    ["feat"]="✨"
    ["fix"]="🐛"
    ["docs"]="📝"
    ["style"]="💄"
    ["refactor"]="♻️"
    ["test"]="✅"
    ["chore"]="🔧"
    ["build"]="👷"
    ["ci"]="⚙️"
    ["perf"]="⚡"
    ["revert"]="⏪"
    ["hotfix"]="🚨"
    ["init"]="🎉"
    ["merge"]="🔀"
)

# Function to print colored output
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to print success message
print_success() {
    print_color $GREEN "✅ $1"
}

# Function to print error message
print_error() {
    print_color $RED "❌ $1"
}

# Function to print warning message
print_warning() {
    print_color $YELLOW "⚠️  $1"
}

# Function to print info message
print_info() {
    print_color $BLUE "ℹ️  $1"
}

# Function to get current git user
get_git_user() {
    git config user.name 2>/dev/null || echo "unknown"
}

# Function to get current branch
get_current_branch() {
    git branch --show-current 2>/dev/null
}

# Function to check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir >/dev/null 2>&1; then
        print_error "Not a git repository. Please run 'gf -i' to initialize or navigate to a git project."
        exit 1
    fi
}

# Initialize CHANGELOG.md if it doesn't exist
init_changelog() {
    if [ ! -f "CHANGELOG.md" ]; then
        cat >CHANGELOG.md <<'EOF'
# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project setup
- Basic functionality implemented

EOF
        git add CHANGELOG.md 2>/dev/null
        print_success "CHANGELOG.md created"
    fi
}

# Function to generate detailed file status information
generate_file_status() {
    local status_info=""

    # New files (excluding CHANGELOG.md)
    local new_files=$(git diff --name-only --cached --diff-filter=A 2>/dev/null | grep -v "^CHANGELOG.md$")
    if [ -n "$new_files" ]; then
        status_info+="\n### 🆕 New files\n"
        status_info+=$(echo "$new_files" | sed 's/^/- /')
        status_info+="\n"
    fi

    # Modified files (excluding CHANGELOG.md)
    local modified_files=$(git diff --name-only --cached --diff-filter=M 2>/dev/null | grep -v "^CHANGELOG.md$")
    if [ -n "$modified_files" ]; then
        status_info+="\n### ✏️ Modified files\n"
        status_info+=$(echo "$modified_files" | sed 's/^/- /')
        status_info+="\n"
    fi

    # Deleted files (excluding CHANGELOG.md)
    local deleted_files=$(git diff --name-only --cached --diff-filter=D 2>/dev/null | grep -v "^CHANGELOG.md$")
    if [ -n "$deleted_files" ]; then
        status_info+="\n### 🗑️ Deleted files\n"
        status_info+=$(echo "$deleted_files" | sed 's/^/- /')
        status_info+="\n"
    fi

    # Renamed files (excluding CHANGELOG.md)
    local renamed_files=$(git diff --name-only --cached --diff-filter=R 2>/dev/null | grep -v "^CHANGELOG.md$")
    if [ -n "$renamed_files" ]; then
        status_info+="\n### 🏷️ Renamed files\n"
        status_info+=$(echo "$renamed_files" | sed 's/^/- /')
        status_info+="\n"
    fi

    echo -e "$status_info"
}

# Function to detect commit type based on file changes
detect_commit_type() {
    local staged_files=$(git diff --name-only --cached 2>/dev/null)
    local change_types=$(git diff --name-only --cached 2>/dev/null | xargs -I {} git diff --cached --name-status {} 2>/dev/null | cut -f1 | sort | uniq)

    # Filter out CHANGELOG.md from detection to avoid affecting commit type
    staged_files=$(echo "$staged_files" | grep -v "^CHANGELOG.md$")

    # Check for new features (new files with significant code)
    if echo "$staged_files" | grep -q -E '(src/|lib/|app/|components/|pages/|api/).*\.(js|ts|jsx|tsx|py|php|java|rb|go|c|cpp|cs|kt|swift)$'; then
        if echo "$change_types" | grep -q '^A'; then
            echo "feat"
            return
        fi
    fi

    # Check for documentation changes
    if echo "$staged_files" | grep -q -E '(README|docs/|\.md$|\.txt$|\.rst$)'; then
        echo "docs"
        return
    fi

    # Check for test changes
    if echo "$staged_files" | grep -q -E '(test/|spec/|__tests__|\.test\.|\.spec\.|cypress/|jest\.)'; then
        echo "test"
        return
    fi

    # Check for style changes
    if echo "$staged_files" | grep -q -E '\.(css|scss|sass|less|styl)$'; then
        echo "style"
        return
    fi

    # Check for build/config changes
    if echo "$staged_files" | grep -q -E '(package\.json|composer\.json|Gemfile|requirements\.txt|Dockerfile|\.yml$|\.yaml$|webpack|gulpfile|Makefile)'; then
        echo "chore"
        return
    fi

    # Check for bug fixes (modifications with keywords)
    if echo "$staged_files" | grep -q -E '(fix|bug|error|issue|patch)' && echo "$change_types" | grep -q '^M'; then
        echo "fix"
        return
    fi

    # Check for refactoring (modifications to code files without bug keywords)
    if echo "$change_types" | grep -q '^M' && echo "$staged_files" | grep -q -E '\.(js|ts|jsx|tsx|py|php|java|rb|go|c|cpp|cs|kt|swift)$'; then
        echo "refactor"
        return
    fi

    # Default to chore
    echo "chore"
}

# Function to generate intelligent commit message
generate_commit_message() {
    local staged_files=$(git diff --name-only --cached 2>/dev/null)
    local num_changes=$(echo "$staged_files" | grep -c '^' 2>/dev/null || echo "0")
    local user=$(get_git_user)

    if [ $num_changes -eq 0 ]; then
        print_error "No staged changes found"
        return 1
    fi

    # Skip if only CHANGELOG.md is staged (avoid recursive commits)
    if [ $num_changes -eq 1 ] && echo "$staged_files" | grep -q "^CHANGELOG.md$"; then
        print_error "Only CHANGELOG.md staged - this would create recursive commits"
        return 1
    fi

    local commit_type=$(detect_commit_type)
    local emoji=${GITMOJI[$commit_type]}

    # Generate intelligent short description
    local short_desc=""
    case $commit_type in
    "feat")
        local main_file=$(echo "$staged_files" | grep -E '\.(js|ts|jsx|tsx|py|php|java|rb|go)$' | head -n1)
        if [ -n "$main_file" ]; then
            local filename=$(basename "$main_file" | sed 's/\.[^.]*$//' | sed 's/_/ /g' | sed 's/-/ /g')
            short_desc="add $filename functionality"
        else
            short_desc="add new functionality"
        fi
        ;;
    "fix")
        local main_file=$(echo "$staged_files" | head -n1)
        if [ -n "$main_file" ]; then
            local filename=$(basename "$main_file")
            short_desc="resolve issues in $filename"
        else
            short_desc="resolve issues and bugs"
        fi
        ;;
    "docs") short_desc="update documentation" ;;
    "style") short_desc="improve code formatting and styles" ;;
    "test") short_desc="add or update tests" ;;
    "refactor") short_desc="refactor code structure" ;;
    "chore") short_desc="update configuration and dependencies" ;;
    *) short_desc="update files" ;;
    esac

    # Generate detailed file status
    local file_status=$(generate_file_status)

    # Combine messages in the requested format
    echo -e "$emoji $commit_type: $short_desc by @$user$file_status"
}

# Function to update CHANGELOG.md automatically
update_changelog() {
    local commit_message="$1"
    local changelog_file="CHANGELOG.md"

    if [ ! -f "$changelog_file" ]; then
        return
    fi

    # Extract commit info
    local commit_type=$(echo "$commit_message" | grep -o -E '(feat|fix|docs|style|refactor|test|chore|build|ci|perf|revert):' | sed 's/://')
    local commit_desc=$(echo "$commit_message" | sed -E 's/^[^:]+: (.+) by @.*$/\1/')
    local user=$(echo "$commit_message" | grep -o 'by @[^[:space:]]*' | sed 's/by @//')
    local short_hash=$(git rev-parse --short HEAD 2>/dev/null || echo "pending")

    # Don't update changelog for changelog-related commits to avoid recursion
    if [[ "$commit_desc" =~ "update changelog"i ]] || [[ "$commit_desc" =~ "changelog"i && "$commit_type" == "chore" ]]; then
        return
    fi

    # Map commit type to changelog section
    case $commit_type in
    "feat") local section="### Added" ;;
    "fix") local section="### Fixed" ;;
    "docs") local section="### Documentation" ;;
    "style") local section="### Style" ;;
    "refactor") local section="### Refactored" ;;
    "test") local section="### Testing" ;;
    "chore") local section="### Maintenance" ;;
    "perf") local section="### Performance" ;;
    "build") local section="### Build" ;;
    *) local section="### Changed" ;;
    esac

    # Create changelog entry
    local changelog_entry="- \`$short_hash\`: $commit_desc by @$user"

    # Check if this exact entry already exists to avoid duplicates
    if grep -q "$changelog_entry" "$changelog_file"; then
        return
    fi

    # Create a backup of the original file to compare later
    local original_content=$(cat "$changelog_file")

    # Update CHANGELOG.md
    if grep -q "## \[Unreleased\]" "$changelog_file"; then
        if ! grep -q "$section" "$changelog_file"; then
            # Section doesn't exist, add it
            sed -i "/## \[Unreleased\]/a\\$section\n$changelog_entry" "$changelog_file"
        else
            # Section exists, append to it
            sed -i "/$section/a\\$changelog_entry" "$changelog_file"
        fi
    else
        # Create new Unreleased section
        echo -e "## [Unreleased]\n$section\n$changelog_entry\n\n$(cat "$changelog_file")" >"$changelog_file"
    fi

    # Only report update if file actually changed
    local new_content=$(cat "$changelog_file")
    if [ "$original_content" != "$new_content" ]; then
        print_info "CHANGELOG.md updated automatically"
    fi
}

# Function to initialize git flow
init_git_flow() {
    print_info "Initializing Git Flow in current directory..."

    if git rev-parse --git-dir >/dev/null 2>&1; then
        print_warning "Git repository already exists."
    else
        git init
        print_success "Git repository initialized"
    fi

    # Initialize CHANGELOG.md
    init_changelog

    # Create initial commit if none exists
    if ! git rev-parse HEAD >/dev/null 2>&1; then
        local user=$(get_git_user)
        git commit --allow-empty -m "$EMOJI_INIT init: initial repository setup by @$user

Project initialization
- Created empty repository
- Added CHANGELOG.md
- Ready for development"
        print_success "Initial commit created"
    fi

    # Ensure main branch exists
    if ! git show-ref --verify --quiet refs/heads/main; then
        git checkout -b main 2>/dev/null || git branch -M main
        print_success "Main branch configured"
    fi

    print_success "Git Flow initialized successfully!"
}

# Function to start a new branch
start_branch() {
    check_git_repo

    local branch_type=$1
    local branch_name=$2

    if [[ -z "$branch_name" ]]; then
        print_error "Branch name is required. Usage: gf -s -f <branch-name>"
        exit 1
    fi

    # Update repository
    print_info "Updating repository..."
    git pull origin main 2>/dev/null || print_warning "Could not pull from origin/main"

    # Switch to main branch
    git checkout main 2>/dev/null || {
        print_error "Could not switch to main branch"
        exit 1
    }
    echo $branch_type
    # Create and switch to new branch
    local full_branch_name=""
    case $branch_type in
    "--feature" | "-f")
        full_branch_name="feature/$branch_name"
        ;;
    "--hotfix" | "-h")
        full_branch_name="hotfix/$branch_name"
        ;;
    "--bugfix" | "-b")
        full_branch_name="bugfix/$branch_name"
        ;;
    "--release" | "-r")
        full_branch_name="release/$branch_name"
        ;;
    *)
        print_error "Invalid branch type. Use: -f (feature), -h (hotfix), -b (bugfix), -r (release)"
        exit 1
        ;;
    esac

    git checkout -b "$full_branch_name"
    print_success "Created and switched to branch: $full_branch_name"
}

# Function to add files
add_files() {
    check_git_repo

    if [[ $# -eq 0 ]]; then
        git add .
        print_success "All changes added to staging area"
    else
        git add "$@"
        print_success "Specified files added to staging area"
    fi

    # Show status
    print_info "Current status:"
    git status --short
}

# Function to push with automatic commit and MR
push_changes() {
    check_git_repo

    local custom_message="$*"

    # Get current branch early for main branch check
    local current_branch=$(git rev-parse --abbrev-ref HEAD)

    if [[ -z "$current_branch" ]]; then
        print_error "Could not determine current branch"
        exit 1
    fi

    # Skip MR creation for main branch
    if [ "$current_branch" = "main" ]; then
        print_warning "No MR will be created for main branch"
    fi

    # Initialize changelog if it doesn't exist
    init_changelog

    # Check if there are any changes to commit
    if git diff --cached --quiet && git diff --quiet; then
        print_warning "No changes to commit"
        return 0
    fi

    # Check if there are staged changes
    if ! git diff --cached --quiet; then
        local commit_message=""

        if [[ -n "$custom_message" ]]; then
            local user=$(get_git_user)
            local commit_type=$(detect_commit_type)
            local emoji=${GITMOJI[$commit_type]}

            # Check if message already has type prefix
            if [[ "$custom_message" =~ ^(feat|fix|docs|style|refactor|test|chore|build|ci|perf|revert): ]]; then
                commit_message="$emoji $custom_message by @$user"
            else
                commit_message="$emoji $commit_type: $custom_message by @$user"
            fi
        else
            commit_message=$(generate_commit_message)
            if [[ $? -ne 0 ]]; then
                print_error "No changes to commit"
                exit 1
            fi
        fi

        # Extract main commit line and body
        local commit_title=$(echo "$commit_message" | head -n1)
        local commit_body=$(echo "$commit_message" | tail -n +2 | sed 's/^###/**/g')

        # Create commit
        if [[ -n "$commit_body" && "$commit_body" != " " ]]; then
            git commit -m "$commit_title" -m "$commit_body"
        else
            git commit -m "$commit_title"
        fi

        print_success "Commit created with semantic message"

        # Update changelog after commit to get correct hash
        update_changelog "$commit_title"

        # Check if CHANGELOG.md was actually modified and add it to the same commit
        if ! git diff --quiet CHANGELOG.md 2>/dev/null; then
            git add CHANGELOG.md
            git commit --amend --no-edit
            print_info "CHANGELOG.md updated and included in commit"
        fi

    elif ! git diff --quiet; then
        print_warning "You have unstaged changes. Run 'gf -a' first to stage them."
        exit 1
    else
        print_warning "No changes to commit"
        return 0
    fi

    # Push to remote
    print_info "Pushing to remote repository..."
    local push_output=$(git push -u origin "$current_branch" 2>&1)

    if [[ $? -eq 0 ]]; then
        print_success "Successfully pushed to origin/$current_branch"

        # Skip MR creation for main branch
        if [ "$current_branch" = "main" ]; then
            print_info "Push to main completed successfully"
            return 0
        fi

        # Generate MR/PR URL for non-main branches
        local remote_url=$(git remote get-url origin 2>/dev/null | sed 's/\.git$//' | sed 's/git@\(.*\):/https:\/\/\1\//' | sed 's/ssh:\/\/git@/https:\/\//')
        local mr_url=""

        if [[ "$remote_url" == *"gitlab"* ]]; then
            mr_url="${remote_url}/-/merge_requests/new?merge_request[source_branch]=${current_branch}&merge_request[target_branch]=main"
            print_info "Opening GitLab Merge Request..."
        elif [[ "$remote_url" == *"github"* ]]; then
            mr_url="${remote_url}/compare/main...${current_branch}?expand=1"
            print_info "Opening GitHub Pull Request..."
        fi

        if [[ -n "$mr_url" ]]; then
            if command -v xdg-open >/dev/null; then
                xdg-open "$mr_url" 2>/dev/null &
            elif command -v open >/dev/null; then
                open "$mr_url" 2>/dev/null &
            else
                print_info "Merge/Pull Request URL: $mr_url"
            fi
        fi
    else
        if echo "$push_output" | grep -q "rejected"; then
            print_error "Push rejected due to conflicts. Run 'gf -m' to merge latest changes."
        else
            print_error "Push failed: $push_output"
        fi
        exit 1
    fi
}

# Function to merge from main
merge_from_main() {
    check_git_repo

    local current_branch=$(get_current_branch)

    if [[ "$current_branch" == "main" ]]; then
        print_warning "Already on main branch. Nothing to merge."
        exit 0
    fi

    print_info "Merging latest changes from main..."

    # Fetch latest changes
    git fetch origin main

    # Merge with no-ff and no-commit to allow conflict resolution
    git merge origin/main --no-ff --no-commit

    if [[ $? -eq 0 ]]; then
        # Check if there are changes to commit
        if ! git diff --cached --quiet; then
            local user=$(get_git_user)
            git commit -m "$EMOJI_MERGE merge: sync with main branch by @$user

Merged latest changes from main
- Updated branch: $current_branch
- Synchronized with remote changes"
            print_success "Successfully merged changes from main"
        else
            print_info "No changes to merge from main"
        fi
    else
        print_warning "Merge conflicts detected. Please resolve conflicts and run 'git commit' when ready."
        print_info "After resolving conflicts, you can continue with 'gf -p' to push your changes."
    fi
}

# Function to finish/delete branch
finish_branch() {
    check_git_repo

    local current_branch=$(get_current_branch)

    if [[ "$current_branch" == "main" ]]; then
        print_error "Cannot delete main branch"
        exit 1
    fi

    if [[ -z "$current_branch" ]]; then
        print_error "Could not determine current branch"
        exit 1
    fi

    # Switch to main branch
    git checkout main

    # Delete local branch
    git branch -D "$current_branch"
    print_success "Deleted local branch: $current_branch"

    # Delete remote branch (suppress error if it doesn't exist)
    git push origin --delete "$current_branch" 2>/dev/null &&
        print_success "Deleted remote branch: $current_branch" ||
        print_info "Remote branch did not exist or could not be deleted"
}

# Function to view changelog
view_changelog() {
    check_git_repo

    local lines=${1:-20}
    local format=${2:-"file"}

    if [[ "$format" == "git" ]]; then
        print_info "Git history (last $lines commits):"
        git log --oneline -n "$lines" --pretty=format:"%C(yellow)%h%C(reset): %s" --all
        echo ""
    else
        if [[ -f "CHANGELOG.md" ]]; then
            print_info "Project CHANGELOG.md:"
            head -n $lines CHANGELOG.md | while IFS= read -r line; do
                print_color $CYAN "$line"
            done
        else
            print_warning "CHANGELOG.md not found. Initialize with 'gf -i' or create manually."
            print_info "Showing git log instead:"
            git log --oneline -n 10 --pretty=format:"%C(yellow)%h%C(reset): %s"
        fi
        echo ""
    fi

    print_success "Changelog displayed successfully!"
}

# Function to show help
show_help() {
    print_color $CYAN "🚀 GF - Advanced Git Flow Command v1.1.3"
    echo ""
    print_color $WHITE "USAGE:"
    echo "  gf [OPTION] [ARGUMENTS]"
    echo ""
    print_color $WHITE "OPTIONS:"
    echo ""
    print_color $GREEN "  -i, --init"
    echo "    Initialize git flow in current project"
    echo "    Creates git repository, CHANGELOG.md and initial commit"
    echo ""
    print_color $GREEN "  -s, --start [TYPE] [NAME]"
    echo "    Start a new branch from main with automatic pull"
    echo "    Types:"
    echo "      -f, --feature   : Feature branch (feature/name)"
    echo "      -h, --hotfix    : Hotfix branch (hotfix/name)"
    echo "      -b, --bugfix    : Bugfix branch (bugfix/name)"
    echo "      -r, --release   : Release branch (release/name)"
    echo ""
    print_color $GREEN "  -a, --add [FILES...]"
    echo "    Add files to staging area"
    echo "    Without arguments: adds all changes (git add .)"
    echo "    With arguments: adds specified files"
    echo ""
    print_color $GREEN "  -p, --push [MESSAGE]"
    echo "    Create semantic commit and push to remote"
    echo "    Auto-generates commit message or uses provided message"
    echo "    Automatically updates CHANGELOG.md and opens MR URL"
    echo "    Note: No MR created when on main branch"
    echo ""
    print_color $GREEN "  -m, --merge"
    echo "    Merge latest changes from main branch"
    echo "    Uses --no-ff --no-commit for conflict resolution"
    echo ""
    print_color $GREEN "  -f, --finish"
    echo "    Delete current branch (local and remote)"
    echo "    Switches to main before deletion"
    echo ""
    print_color $GREEN "  -c, --changelog [LINES] [FORMAT]"
    echo "    View project changelog"
    echo "    LINES: number of lines to show (default: 20)"
    echo "    FORMAT: 'file' (CHANGELOG.md) or 'git' (git log) (default: file)"
    echo ""
    print_color $GREEN "  -h, --help"
    echo "    Show this help message"
    echo ""
    print_color $WHITE "EXAMPLES:"
    echo ""
    print_color $YELLOW "  gf -i"
    echo "    Initialize git flow with CHANGELOG.md"
    echo ""
    print_color $YELLOW "  gf -s -f ticket-1000"
    echo "    Create feature branch 'feature/ticket-1000'"
    echo ""
    print_color $YELLOW "  gf -a"
    echo "    Add all changes to staging"
    echo ""
    print_color $YELLOW "  gf -p"
    echo "    Auto-commit with semantic message and push"
    echo ""
    print_color $YELLOW "  gf -p \"implement user authentication\""
    echo "    Commit with custom message and push"
    echo ""
    print_color $YELLOW "  gf -m"
    echo "    Merge latest changes from main"
    echo ""
    print_color $YELLOW "  gf -f"
    echo "    Delete current branch"
    echo ""
    print_color $YELLOW "  gf -c"
    echo "    View CHANGELOG.md (20 lines)"
    echo ""
    print_color $YELLOW "  gf -c 50 file"
    echo "    View CHANGELOG.md (50 lines)"
    echo ""
    print_color $YELLOW "  gf -c 10 git"
    echo "    View git history (10 commits)"
    echo ""
}

# Main script logic
main() {
    if [[ $# -eq 0 ]]; then
        show_help
        exit 0
    fi

    case $1 in
    -i | --init)
        init_git_flow
        ;;
    -s | --start)
        if [[ $# -lt 3 ]]; then
            # print_error "Usage: gf -s [TYPE] [NAME]"
            git status
            exit 0
        fi
        start_branch "$2" "$3"
        ;;
    -a | --add)
        shift
        add_files "$@"
        ;;
    -p | --push)
        shift
        push_changes "$@"
        ;;
    -m | --merge)
        merge_from_main
        ;;
    -f | --finish)
        finish_branch
        ;;
    -c | --changelog)
        shift
        view_changelog "$@"
        ;;
    -h | --help)
        show_help
        ;;
    *)
        print_error "Unknown option: $1"
        echo "Use 'gf -h' for help"
        exit 1
        ;;
    esac
}

# Execute main function with all arguments
main "$@"

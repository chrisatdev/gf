#!/bin/bash
# -------------------------------------------------------------------
# Git Flow Enhanced (gf)
# Version: 1.1.2
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

# Function to detect commit type based on file changes
detect_commit_type() {
    local modified_files=$(git diff --cached --name-only)
    local new_files=$(git diff --cached --diff-filter=A --name-only)
    local deleted_files=$(git diff --cached --diff-filter=D --name-only)

    # Check for specific patterns
    if echo "$modified_files" | grep -qE "\.(md|txt|rst)$|README|CHANGELOG|docs/"; then
        echo "docs"
    elif echo "$modified_files" | grep -qE "test|spec|__tests__|\.test\.|\.spec\."; then
        echo "test"
    elif echo "$modified_files" | grep -qE "package\.json|Gemfile|requirements\.txt|composer\.json"; then
        echo "chore"
    elif echo "$modified_files" | grep -qE "\.css$|\.scss$|\.sass$|\.less$|\.styl$"; then
        echo "style"
    elif [[ -n "$new_files" ]]; then
        echo "feat"
    elif [[ -n "$deleted_files" ]]; then
        echo "chore"
    else
        # Default to feat for new functionality or fix for modifications
        if git diff --cached | grep -q "^+.*function\|^+.*class\|^+.*const\|^+.*let\|^+.*var"; then
            echo "feat"
        else
            echo "fix"
        fi
    fi
}

# Function to get emoji for commit type
get_emoji_for_type() {
    case $1 in
    "feat") echo $EMOJI_FEAT ;;
    "fix") echo $EMOJI_FIX ;;
    "docs") echo $EMOJI_DOCS ;;
    "style") echo $EMOJI_STYLE ;;
    "refactor") echo $EMOJI_REFACTOR ;;
    "test") echo $EMOJI_TEST ;;
    "chore") echo $EMOJI_CHORE ;;
    "hotfix") echo $EMOJI_HOTFIX ;;
    *) echo $EMOJI_FEAT ;;
    esac
}

# Function to generate commit message
generate_commit_message() {
    local commit_type=$(detect_commit_type)
    local emoji=$(get_emoji_for_type $commit_type)
    local modified_files=$(git diff --cached --name-only)
    local files_count=$(echo "$modified_files" | wc -l)

    # Generate short message
    local short_message=""
    case $commit_type in
    "feat")
        short_message="add new functionality"
        ;;
    "fix")
        short_message="resolve issues and bugs"
        ;;
    "docs")
        short_message="update documentation"
        ;;
    "style")
        short_message="improve code formatting"
        ;;
    "refactor")
        short_message="refactor code structure"
        ;;
    "test")
        short_message="add or update tests"
        ;;
    "chore")
        short_message="update dependencies and tools"
        ;;
    esac

    # Generate long message with file list
    local long_message="Modified files:"
    while IFS= read -r file; do
        [[ -n "$file" ]] && long_message="$long_message\n- $file"
    done <<<"$modified_files"

    local user=$(get_git_user)
    local branch=$(get_current_branch)
    long_message="$long_message\n\nBranch: $branch\nAuthor: @$user"

    echo -e "$emoji $commit_type: $short_message\n\n$long_message"
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

    # Create initial commit if none exists
    if ! git rev-parse HEAD >/dev/null 2>&1; then
        git commit --allow-empty -m "$EMOJI_INIT init: initial commit

Repository initialization
- Created empty repository
- Ready for development

Author: @$(get_git_user)"
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

    # Create and switch to new branch
    local full_branch_name=""
    case $branch_type in
    "feature" | "f")
        full_branch_name="feature/$branch_name"
        ;;
    "hotfix" | "h")
        full_branch_name="hotfix/$branch_name"
        ;;
    "bugfix" | "b")
        full_branch_name="bugfix/$branch_name"
        ;;
    "release" | "r")
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
    local current_branch=$(get_current_branch)

    if [[ -z "$current_branch" ]]; then
        print_error "Could not determine current branch"
        exit 1
    fi

    # Check if there are staged changes
    if ! git diff --cached --quiet; then
        # Generate or use custom commit message
        local commit_message=""
        if [[ -n "$custom_message" ]]; then
            local commit_type=$(detect_commit_type)
            local emoji=$(get_emoji_for_type $commit_type)
            commit_message="$emoji $commit_type: $custom_message"
        else
            commit_message=$(generate_commit_message)
        fi

        # Create commit
        git commit -m "$commit_message"
        print_success "Commit created with semantic message"
    elif ! git diff --quiet; then
        print_warning "You have unstaged changes. Run 'gf -a' first to stage them."
        exit 1
    fi

    # Push to remote
    print_info "Pushing to remote repository..."
    local push_output=$(git push -u origin "$current_branch" 2>&1)

    if [[ $? -eq 0 ]]; then
        print_success "Successfully pushed to origin/$current_branch"

        # Extract MR URL from push output
        local mr_url=$(echo "$push_output" | grep -o 'https://[^[:space:]]*/-/merge_requests/new[^[:space:]]*' | head -1)

        if [[ -n "$mr_url" ]]; then
            print_info "Opening Merge Request in browser..."
            if command -v xdg-open >/dev/null; then
                xdg-open "$mr_url" 2>/dev/null &
            elif command -v open >/dev/null; then
                open "$mr_url" 2>/dev/null &
            else
                print_info "Merge Request URL: $mr_url"
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
            git commit -m "$EMOJI_MERGE merge: sync with main branch

Merged latest changes from main
- Updated branch: $current_branch
- Synchronized with remote changes

Author: @$user"
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

# Function to show help
show_help() {
    print_color $CYAN "🚀 GF - Advanced Git Flow Command v2.0.0"
    echo ""
    print_color $WHITE "USAGE:"
    echo "  gf [OPTION] [ARGUMENTS]"
    echo ""
    print_color $WHITE "OPTIONS:"
    echo ""
    print_color $GREEN "  -i, --init"
    echo "    Initialize git flow in current project"
    echo "    Creates git repository and initial commit"
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
    echo "    Automatically opens Merge Request URL"
    echo ""
    print_color $GREEN "  -m, --merge"
    echo "    Merge latest changes from main branch"
    echo "    Uses --no-ff --no-commit for conflict resolution"
    echo ""
    print_color $GREEN "  -f, --finish"
    echo "    Delete current branch (local and remote)"
    echo "    Switches to main before deletion"
    echo ""
    print_color $GREEN "  -h, --help"
    echo "    Show this help message"
    echo ""
    print_color $WHITE "EXAMPLES:"
    echo ""
    print_color $YELLOW "  gf -i"
    echo "    Initialize git flow"
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
    print_color $WHITE "FEATURES:"
    echo "  • Semantic commits with gitmoji integration"
    echo "  • Automatic commit message generation"
    echo "  • Conflict detection and resolution guidance"
    echo "  • Automatic Merge Request creation"
    echo "  • Branch naming conventions"
    echo "  • Colored output for better UX"
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
            print_error "Usage: gf -s [TYPE] [NAME]"
            exit 1
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

#!/bin/bash
# -------------------------------------------------------------------
# Git Flow Enhanced (gf)
# Version: 1.1.0
# Author: Christian Benítez
# GitHub: https://github.com/chrisatdev
# Description: Advanced Git workflow automation tool
# -------------------------------------------------------------------

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

#Initial cursor setup
trap 'tput cnorm; exit' INT TERM EXIT
tput civis

# Gitmoji mapping
declare -A GITMOJI=(
    ["feat"]="✨"      # New feature
    ["fix"]="🐛"       # Bug fix
    ["docs"]="📝"      # Documentation
    ["style"]="💄"     # Code style
    ["refactor"]="♻️" # Refactoring
    ["test"]="✅"      # Testing
    ["chore"]="🔧"     # Chores
    ["build"]="👷"     # Build system
    ["ci"]="⚙️"       # CI configuration
    ["perf"]="⚡"      # Performance
    ["revert"]="⏪"    # Revert changes
)

# Initialize CHANGELOG.md if it doesn't exist
init_changelog() {
    if [ ! -f "CHANGELOG.md" ]; then
        echo -e "# CHANGELOG\n\n## [Unreleased]\n### Added\n- Initial version" >CHANGELOG.md
        git add CHANGELOG.md 2>/dev/null
    fi
}

# Function to show help
show_help() {
    echo -e "${GREEN}🚀 Git Flow Enhanced (gf)${NC}"
    echo -e "${GREEN} Version: 1.1.0 - by Christian Benítez${NC}"
    echo -e "${GREEN} GitHub: https://github.com/chrisatdev${NC}"
    echo -e "${GREEN}   Usage:${NC}"
    echo -e "  ${CYAN}gf -i${NC}                        ${GREEN}🆕${NC} Initialize new Git repository"
    echo -e "  ${CYAN}gf -s -f [name]${NC}              ${GREEN}✨${NC} Create feature branch (feature/name)"
    echo -e "  ${CYAN}gf -s -h [name]${NC}              ${RED}🐛${NC} Create hotfix branch (hotfix/name)"
    echo -e "  ${CYAN}gf -s -b [name]${NC}              ${YELLOW}🚑${NC} Create bugfix branch (bugfix/name)"
    echo -e "  ${CYAN}gf -s -r [name]${NC}              ${BLUE}🚀${NC} Create release branch (release/name)"
    echo -e "  ${CYAN}gf -a [files]${NC}                ${GREEN}📦${NC} Stage changes (stage all if no files specified)"
    echo -e "  ${CYAN}gf -p \"[msg]\"${NC}              ${GREEN}💾${NC} Commit (with message) and push, then open MR/PR"
    echo -e "  ${CYAN}gf -m${NC}                        ${GREEN}🔀${NC} Merge main into current branch (handle conflicts)"
    echo -e "  ${CYAN}gf -f${NC}                        ${RED}🗑️${NC} Finish and delete current branch (local & remote)"
    echo -e "  ${CYAN}gf -r [source] [target]${NC}      ${PURPLE}🔄${NC} Create MR from source to target branch (GitLab)"
    echo -e "  ${CYAN}gf -h${NC}                        ${BLUE}ℹ️${NC} Show this help"
    echo -e "\n${PURPLE}📚 Examples:${NC}"
    echo -e "  ${CYAN}gf -i${NC}"
    echo -e "  ${CYAN}gf -s -f ticket-1000${NC}"
    echo -e "  ${CYAN}gf -a${NC}"
    echo -e "  ${CYAN}gf -p \"feat: add new API endpoint\"${NC}"
    echo -e "  ${CYAN}gf -m${NC}"
    echo -e "  ${CYAN}gf -f${NC}"
    echo -e "  ${CYAN}gf -r main dev${NC}"
}

# Function to generate detailed file status information
generate_file_status() {
    local status_info=""

    # New files
    local new_files=$(git diff --name-only --cached --diff-filter=A)
    if [ -n "$new_files" ]; then
        status_info+="\n### 🆕 New files\n"
        status_info+=$(echo "$new_files" | sed 's/^/- /')
        status_info+="\n"
    fi

    # Modified files
    local modified_files=$(git diff --name-only --cached --diff-filter=M)
    if [ -n "$modified_files" ]; then
        status_info+="\n### ✏️ Modified files\n"
        status_info+=$(echo "$modified_files" | sed 's/^/- /')
        status_info+="\n"
    fi

    # Deleted files
    local deleted_files=$(git diff --name-only --cached --diff-filter=D)
    if [ -n "$deleted_files" ]; then
        status_info+="\n### 🗑️ Deleted files\n"
        status_info+=$(echo "$deleted_files" | sed 's/^/- /')
        status_info+="\n"
    fi

    # Renamed files
    local renamed_files=$(git diff --name-only --cached --diff-filter=R)
    if [ -n "$renamed_files" ]; then
        status_info+="\n### 🏷️ Renamed files\n"
        status_info+=$(echo "$renamed_files" | sed 's/^/- /')
        status_info+="\n"
    fi

    echo -e "$status_info"
}

# Function to detect change type and generate semantic message
generate_semantic_message() {
    local staged_files=$(git diff --name-only --cached)
    local num_changes=$(echo "$staged_files" | grep -c '^')

    if [ $num_changes -eq 0 ]; then
        echo "🔧 chore: update repository"
        return 1
    fi

    # Analyze changes
    local change_types=$(git diff --name-only --cached | xargs -I {} git diff --cached --name-status {} | cut -f1 | sort | uniq)

    # Determine semantic type
    local semantic_type="chore"
    local emoji="🔧"

    # Check for new features (new files with significant code)
    if echo "$staged_files" | grep -q -E 'src/|lib/|app/|main/'; then
        if echo "$change_types" | grep -q '^A'; then
            semantic_type="feat"
            emoji=${GITMOJI["feat"]}
        fi
    fi

    # Check for bug fixes (changes to existing files)
    if echo "$change_types" | grep -q '^M'; then
        if echo "$staged_files" | grep -q -E 'fix|bug|error|issue'; then
            semantic_type="fix"
            emoji=${GITMOJI["fix"]}
        fi
    fi

    # Check for documentation changes
    if echo "$staged_files" | grep -q -E 'README|docs/|\.md$'; then
        semantic_type="docs"
        emoji=${GITMOJI["docs"]}
    fi

    # Check for style changes
    if echo "$staged_files" | grep -q -E '\.css$|\.scss$|\.less$|style'; then
        semantic_type="style"
        emoji=${GITMOJI["style"]}
    fi

    # Check for test changes
    if echo "$staged_files" | grep -q -E 'test/|spec/|__tests__|\.test\.|\.spec\.'; then
        semantic_type="test"
        emoji=${GITMOJI["test"]}
    fi

    # Generate short description
    local short_desc=""
    case $semantic_type in
    "feat") short_desc="Add $(echo "$staged_files" | head -n1 | xargs basename)" ;;
    "fix") short_desc="Fix issue in $(echo "$staged_files" | head -n1 | xargs basename)" ;;
    "docs") short_desc="Update documentation" ;;
    "style") short_desc="Improve code style" ;;
    "test") short_desc="Add tests" ;;
    *) short_desc="Update files" ;;
    esac

    # Generate detailed file status
    local file_status=$(generate_file_status)

    # Combine messages
    echo -e "${emoji} ${semantic_type}: ${short_desc}\n\n${file_status}"
}

# Function to update CHANGELOG.md
update_changelog() {
    local commit_message="$1"
    local changelog_file="CHANGELOG.md"

    if [ ! -f "$changelog_file" ]; then
        return
    fi

    # Extract commit type and message
    local commit_type=$(echo "$commit_message" | grep -o -E '^(feat|fix|docs|style|refactor|test|chore|build|ci|perf|revert)')
    local commit_desc=$(echo "$commit_message" | sed -E 's/^[^:]+: //' | head -n1)

    # Map commit type to changelog section
    case $commit_type in
    "feat") local section="### Added" ;;
    "fix") local section="### Fixed" ;;
    *) local section="### Changed" ;;
    esac

    # Update CHANGELOG.md
    if grep -q "## \[Unreleased\]" "$changelog_file"; then
        # Add to existing Unreleased section
        if ! grep -q "$section" "$changelog_file"; then
            # Section doesn't exist yet, add it
            sed -i "/## \[Unreleased\]/a $section\n- $commit_desc" "$changelog_file"
        else
            # Section exists, append to it
            sed -i "/$section/a - $commit_desc" "$changelog_file"
        fi
    else
        # Create new Unreleased section
        echo -e "## [Unreleased]\n$section\n- $commit_desc\n\n$(cat "$changelog_file")" >"$changelog_file"
    fi

    # Stage CHANGELOG.md changes
    git add "$changelog_file" 2>/dev/null
}

# Initialize repository
init_repo() {
    echo -e "${GREEN}🆕 Initializing new Git repository...${NC}"
    git init
    if [ $? -eq 0 ]; then
        init_changelog
        git commit --allow-empty -m "${GITMOJI["chore"]} chore: Initial commit"
        echo -e "${GREEN}✅ Repository initialized with empty commit${NC}"
    else
        echo -e "${RED}❌ Error initializing repository${NC}"
        exit 1
    fi
}

# Create new branch
start_branch() {
    local branch_type=""
    local branch_name=""
    local emoji=""

    case $1 in
    -f)
        branch_type="feature"
        emoji="✨"
        ;;
    -h)
        branch_type="hotfix"
        emoji="🐛"
        ;;
    -b)
        branch_type="bugfix"
        emoji="🚑"
        ;;
    -r)
        branch_type="release"
        emoji="🚀"
        ;;
    *)
        echo -e "${RED}❌ Invalid branch type. Use -f, -h, -b or -r${NC}"
        exit 1
        ;;
    esac

    branch_name="$2"

    if [ -z "$branch_name" ]; then
        echo -e "${RED}❌ Branch name is required${NC}"
        show_help
        exit 1
    fi

    full_branch_name="$branch_type/$branch_name"

    echo -e "${GREEN}🔄 Updating main branch...${NC}"
    git checkout main 2>/dev/null || git checkout -b main
    git pull origin main

    if [ $? -ne 0 ]; then
        echo -e "${YELLOW}⚠️ Couldn't pull from origin/main. Using local main branch${NC}"
    fi

    echo -e "${GREEN}🌱 Creating branch: ${CYAN}$full_branch_name ${emoji}${NC}"
    git checkout -b "$full_branch_name"

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Branch ${CYAN}$full_branch_name ${GREEN}created${NC}"
    else
        echo -e "${RED}❌ Error creating branch${NC}"
        exit 1
    fi
}

# Stage changes
add_changes() {
    if [ -z "$1" ]; then
        echo -e "${GREEN}📦 Staging all changes...${NC}"
        git add .
    else
        echo -e "${GREEN}📦 Staging specified files...${NC}"
        git add "$@"
    fi

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Changes staged${NC}"
    else
        echo -e "${RED}❌ Error staging changes${NC}"
        exit 1
    fi
}

# Commit and push
commit_and_push() {
    local commit_message="$1"

    if [ -z "$commit_message" ]; then
        commit_message=$(generate_semantic_message)
        if [ $? -ne 0 ]; then
            echo -e "${RED}❌ No changes to commit${NC}"
            exit 1
        fi
        short_msg=$(echo "$commit_message" | head -n1)
        emoji=$(echo "$short_msg" | grep -o -E '✨|🐛|📝|💄|♻️|✅|🔧|👷|⚙️|⚡|⏪')
        echo -e "${YELLOW}📝 Auto-generated commit message: ${PURPLE}$short_msg ${emoji}${NC}"
    else
        # Add gitmoji if not present in custom message
        if ! grep -q -E '✨|🐛|📝|💄|♻️|✅|🔧|👷|⚙️|⚡|⏪' <<<"$commit_message"; then
            # Try to detect type from message
            if [[ "$commit_message" =~ ^feat ]]; then
                commit_message="${GITMOJI["feat"]} $commit_message"
            elif [[ "$commit_message" =~ ^fix ]]; then
                commit_message="${GITMOJI["fix"]} $commit_message"
            elif [[ "$commit_message" =~ ^docs ]]; then
                commit_message="${GITMOJI["docs"]} $commit_message"
            else
                commit_message="${GITMOJI["chore"]} $commit_message"
            fi
        fi
    fi

    # Update CHANGELOG.md before committing
    update_changelog "$(echo "$commit_message" | head -n1)"

    echo -e "${GREEN}💾 Creating commit...${NC}"
    local md_body=$(echo "$commit_message" | tail -n +3 | sed 's/^• /- /g')
    git commit -m "$(echo "$commit_message" | head -n1)" -m "$md_body"

    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Error creating commit${NC}"
        exit 1
    fi

    current_branch=$(git rev-parse --abbrev-ref HEAD)
    echo -e "${GREEN}📤 Pushing to ${CYAN}$current_branch${GREEN}...${NC}"
    git push -u origin "$current_branch"

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Push successful${NC}"
        # Open MR/PR URL
        remote_url=$(git remote get-url origin | sed 's/\.git$//' | sed 's/git@\(.*\):/\1\//' | sed 's/https:\/\///')
        if [[ $remote_url == *"gitlab"* ]]; then
            mr_url="https://${remote_url}/-/merge_requests/new?merge_request[source_branch]=${current_branch}"
            echo -e "${CYAN}🔗 Opening Merge Request...${NC}"
            xdg-open "$mr_url" 2>/dev/null || open "$mr_url" 2>/dev/null || start "$mr_url" 2>/dev/null
        elif [[ $remote_url == *"github"* ]]; then
            pr_url="https://${remote_url}/compare/${current_branch}?expand=1"
            echo -e "${CYAN}🔗 Opening Pull Request...${NC}"
            xdg-open "$pr_url" 2>/dev/null || open "$pr_url" 2>/dev/null || start "$pr_url" 2>/dev/null
        fi
    else
        echo -e "${RED}❌ Error pushing changes${NC}"
        echo -e "${YELLOW}⚠️ If conflicts exist, run: ${CYAN}gf -m${NC}"
        exit 1
    fi
}

# Merge main into current branch
merge_main() {
    current_branch=$(git rev-parse --abbrev-ref HEAD)

    if [ "$current_branch" = "main" ]; then
        echo -e "${RED}❌ Cannot merge main into itself${NC}"
        exit 1
    fi

    echo -e "${GREEN}🔄 Updating main branch...${NC}"
    git fetch origin main

    echo -e "${GREEN}🔀 Merging main into ${CYAN}$current_branch${GREEN}...${NC}"
    git merge --no-ff --no-commit origin/main

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Merge successful${NC}"
        echo -e "${YELLOW}📝 Review changes and commit when ready${NC}"
    else
        echo -e "${RED}❌ Merge conflicts detected${NC}"
        echo -e "${YELLOW}✏️ Resolve conflicts and commit manually${NC}"
        exit 1
    fi
}

# Finish and delete branch
finish_branch() {
    current_branch=$(git rev-parse --abbrev-ref HEAD)

    if [ "$current_branch" = "main" ]; then
        echo -e "${RED}❌ Cannot delete main branch${NC}"
        exit 1
    fi

    echo -e "${GREEN}🔄 Switching to main branch...${NC}"
    git checkout main

    echo -e "${GREEN}📥 Pulling latest changes...${NC}"
    git pull origin main

    echo -e "${GREEN}🗑️ Deleting local branch ${CYAN}$current_branch${GREEN}...${NC}"
    git branch -D "$current_branch"

    echo -e "${GREEN}♻️ Attempting to delete remote branch...${NC}"
    git push origin --delete "$current_branch" 2>/dev/null

    echo -e "${GREEN}✅ Branch ${CYAN}$current_branch ${GREEN}cleaned up${NC}"
}

# Create MR between branches (GitLab specific)
create_mr() {
    local source_branch=$1
    local target_branch=$2

    if [ -z "$source_branch" ] || [ -z "$target_branch" ]; then
        echo -e "${RED}❌ Both source and target branches are required${NC}"
        echo -e "${YELLOW}Usage: gf -r source target${NC}"
        exit 1
    fi

    # Verify we're using GitLab
    remote_url=$(git remote get-url origin | sed 's/\.git$//' | sed 's/git@\(.*\):/\1\//' | sed 's/https:\/\///')
    if [[ ! $remote_url == *"gitlab"* ]]; then
        echo -e "${RED}❌ MR creation is only supported for GitLab repositories${NC}"
        exit 1
    fi

    echo -e "${GREEN}🔄 Creating MR from ${CYAN}$source_branch${GREEN} to ${CYAN}$target_branch${GREEN}...${NC}"

    # Get current branch to return later
    local current_branch=$(git rev-parse --abbrev-ref HEAD)

    # Checkout source branch and push if needed
    git checkout "$source_branch"
    git push --set-upstream origin "$source_branch" 2>/dev/null

    # Create MR URL
    mr_url="https://${remote_url}/-/merge_requests/new?merge_request[source_branch]=${source_branch}&merge_request[target_branch]=${target_branch}"

    echo -e "${GREEN}✅ MR created${NC}"
    echo -e "${CYAN}🔗 Opening Merge Request...${NC}"
    xdg-open "$mr_url" 2>/dev/null || open "$mr_url" 2>/dev/null || start "$mr_url" 2>/dev/null

    # Return to original branch
    git checkout "$current_branch" 2>/dev/null
}

# Main execution
init_changelog

if [ $# -eq 0 ]; then
    show_help
    exit 0
fi

case $1 in
-i)
    init_repo
    exit 0
    ;;
-s)
    shift
    start_branch "$@"
    exit 0
    ;;
-a)
    shift
    add_changes "$@"
    exit 0
    ;;
-p)
    shift
    commit_and_push "$*"
    exit 0
    ;;
-m)
    merge_main
    exit 0
    ;;
-f)
    finish_branch
    exit 0
    ;;
-r)
    shift
    create_mr "$@"
    exit 0
    ;;
-h)
    show_help
    exit 0
    ;;
*)
    echo -e "${RED}❌ Invalid option. Use -h for help.${NC}"
    exit 1
    ;;
esac

# Restore cursor
tput cnorm

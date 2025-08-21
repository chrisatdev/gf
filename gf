#!/bin/bash
# -------------------------------------------------------------------
# Git Flow Enhanced (gf)
# Version: 1.3.0
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

# Function to get current month-year for archiving
get_current_month_year() {
    date +"%Y-%m"
}

# Function to get month-year from 6 months ago
get_six_months_ago() {
    date -d "6 months ago" +"%Y-%m" 2>/dev/null || date -j -v-6m +"%Y-%m" 2>/dev/null || echo "2025-01"
}

# Function to check if current CHANGELOG.md is from previous month
is_changelog_from_previous_month() {
    if [ ! -f "CHANGELOG.md" ]; then
        return 1
    fi

    # Get the last modification date of CHANGELOG.md
    local last_modified
    if command -v stat >/dev/null 2>&1; then
        # Linux/GNU stat
        last_modified=$(stat -c %Y "CHANGELOG.md" 2>/dev/null)
        if [ -z "$last_modified" ]; then
            # macOS/BSD stat
            last_modified=$(stat -f %m "CHANGELOG.md" 2>/dev/null)
        fi
    else
        return 1
    fi

    if [ -z "$last_modified" ]; then
        return 1
    fi

    # Convert to month-year format
    local file_month_year
    if command -v date >/dev/null 2>&1; then
        file_month_year=$(date -d "@$last_modified" +"%Y-%m" 2>/dev/null || date -r "$last_modified" +"%Y-%m" 2>/dev/null)
    else
        return 1
    fi

    local current_month_year=$(get_current_month_year)

    if [ "$file_month_year" != "$current_month_year" ]; then
        return 0 # Is from previous month
    else
        return 1 # Is from current month
    fi
}

# Function to archive current changelog
archive_changelog() {
    if [ ! -f "CHANGELOG.md" ]; then
        return
    fi

    # Create changelogs directory if it doesn't exist
    if [ ! -d "changelogs" ]; then
        mkdir -p "changelogs"
        echo -e "${GREEN}📁 Created changelogs directory${NC}"
    fi

    # Get the last modification date for naming
    local last_modified
    if command -v stat >/dev/null 2>&1; then
        last_modified=$(stat -c %Y "CHANGELOG.md" 2>/dev/null)
        if [ -z "$last_modified" ]; then
            last_modified=$(stat -f %m "CHANGELOG.md" 2>/dev/null)
        fi
    fi

    local archive_name
    if [ -n "$last_modified" ]; then
        local file_month_year=$(date -d "@$last_modified" +"%Y-%m" 2>/dev/null || date -r "$last_modified" +"%Y-%m" 2>/dev/null)
        archive_name="CHANGELOG-${file_month_year}.md"
    else
        # Fallback to previous month if we can't get the date
        local prev_month=$(date -d "1 month ago" +"%Y-%m" 2>/dev/null || date -j -v-1m +"%Y-%m" 2>/dev/null || echo "2024-01")
        archive_name="CHANGELOG-${prev_month}.md"
    fi

    # Move current changelog to archive
    mv "CHANGELOG.md" "changelogs/$archive_name"
    echo -e "${GREEN}📦 Archived changelog as ${CYAN}changelogs/$archive_name${NC}"

    # Stage the archived file
    git add "changelogs/$archive_name" 2>/dev/null
}

# Function to clean old changelog archives (6+ months)
clean_old_changelogs() {
    if [ ! -d "changelogs" ]; then
        return
    fi

    local six_months_ago=$(get_six_months_ago)
    local files_deleted=0

    # Find and remove old changelog files
    for file in changelogs/CHANGELOG-*.md; do
        if [ -f "$file" ]; then
            # Extract date from filename (CHANGELOG-YYYY-MM.md)
            local file_date=$(echo "$file" | sed -n 's/.*CHANGELOG-\([0-9]\{4\}-[0-9]\{2\}\)\.md/\1/p')

            if [ -n "$file_date" ]; then
                # Compare dates (simple string comparison works for YYYY-MM format)
                if [ "$file_date" \< "$six_months_ago" ]; then
                    echo -e "${YELLOW}🗑️  Removing old changelog: ${CYAN}$file${NC}"
                    git rm "$file" 2>/dev/null || rm "$file"
                    files_deleted=$((files_deleted + 1))
                fi
            fi
        fi
    done

    if [ $files_deleted -gt 0 ]; then
        echo -e "${GREEN}✅ Cleaned $files_deleted old changelog(s) (6+ months old)${NC}"
    fi
}

# Function to create new changelog for current month
create_new_changelog() {
    local current_date=$(date +"%B %Y")
    echo -e "# CHANGELOG\n\n## [Unreleased] - $current_date\n### Added\n- New changelog for $current_date" >CHANGELOG.md
    git add CHANGELOG.md 2>/dev/null
    echo -e "${GREEN}📝 Created new changelog for ${CYAN}$current_date${NC}"
}

# Initialize CHANGELOG.md with monthly rotation logic
init_changelog() {
    # Clean old changelogs first
    clean_old_changelogs

    # Check if current changelog exists and is from previous month
    if [ -f "CHANGELOG.md" ] && is_changelog_from_previous_month; then
        echo -e "${YELLOW}📅 Changelog is from previous month, archiving...${NC}"
        archive_changelog
        create_new_changelog
    elif [ ! -f "CHANGELOG.md" ]; then
        # No changelog exists, create new one
        create_new_changelog
    fi
}

# Function to get GitLab username
get_gitlab_user() {
    # Try to get from git config first
    local git_user=$(git config user.name 2>/dev/null)
    if [ -n "$git_user" ]; then
        echo "$git_user"
        return
    fi

    # Try to extract from remote URL
    local remote_url=$(git remote get-url origin 2>/dev/null)
    if [[ $remote_url == *"gitlab"* ]]; then
        if [[ $remote_url == *"@"* ]]; then
            # SSH format: git@gitlab.com:username/repo.git
            echo "$remote_url" | sed -n 's/.*@.*:\([^/]*\)\/.*/\1/p'
        else
            # HTTPS format: https://gitlab.com/username/repo.git
            echo "$remote_url" | sed -n 's/.*\/\([^/]*\)\/[^/]*\.git.*/\1/p'
        fi
    else
        echo "developer"
    fi
}

# Function to show help
show_help() {
    echo -e "${GREEN}🚀 Git Flow Enhanced (gf)${NC}"
    echo -e "${GREEN} Version: 1.3.0 - by Christian Benítez${NC}"
    echo -e "${GREEN} GitHub: https://github.com/chrisatdev${NC}"
    echo -e "${GREEN}   Usage:${NC}"
    echo -e "  ${CYAN}gf -i${NC}                        ${GREEN}🆕${NC} Initialize new Git repository"
    echo -e "  ${CYAN}gf -s${NC}                        ${GREEN}✅${NC} Alias to git status"
    echo -e "  ${CYAN}gf -s -f [name]${NC}              ${GREEN}✨${NC} Create feature branch (feature/name)"
    echo -e "  ${CYAN}gf -s -h [name]${NC}              ${RED}🐛${NC} Create hotfix branch (hotfix/name)"
    echo -e "  ${CYAN}gf -s -b [name]${NC}              ${YELLOW}🚑${NC} Create bugfix branch (bugfix/name)"
    echo -e "  ${CYAN}gf -s -r [name]${NC}              ${BLUE}🚀${NC} Create release branch (release/name)"
    echo -e "  ${CYAN}gf -a [files]${NC}                ${GREEN}📦${NC} Stage changes (stage all if no files specified)"
    echo -e "  ${CYAN}gf -p \"[msg]\"${NC}                ${GREEN}💾${NC} Commit (with message) and push, then open MR/PR"
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

# Function to count and categorize file changes
get_file_changes_summary() {
    local new_count=$(git diff --name-only --cached --diff-filter=A | wc -l)
    local modified_count=$(git diff --name-only --cached --diff-filter=M | wc -l)
    local deleted_count=$(git diff --name-only --cached --diff-filter=D | wc -l)
    local renamed_count=$(git diff --name-only --cached --diff-filter=R | wc -l)

    local total_count=$((new_count + modified_count + deleted_count + renamed_count))

    echo "$total_count $new_count $modified_count $deleted_count $renamed_count"
}

# Function to generate detailed file status information
generate_file_status() {
    local status_info=""

    # New files
    local new_files=$(git diff --name-only --cached --diff-filter=A)
    if [ -n "$new_files" ]; then
        status_info+="\n**New files:**\n"
        status_info+=$(echo "$new_files" | sed 's/^/- /')
        status_info+="\n"
    fi

    # Modified files
    local modified_files=$(git diff --name-only --cached --diff-filter=M)
    if [ -n "$modified_files" ]; then
        status_info+="\n**Modified files:**\n"
        status_info+=$(echo "$modified_files" | sed 's/^/- /')
        status_info+="\n"
    fi

    # Deleted files
    local deleted_files=$(git diff --name-only --cached --diff-filter=D)
    if [ -n "$deleted_files" ]; then
        status_info+="\n**Deleted files:**\n"
        status_info+=$(echo "$deleted_files" | sed 's/^/- /')
        status_info+="\n"
    fi

    # Renamed files
    local renamed_files=$(git diff --name-only --cached --diff-filter=R)
    if [ -n "$renamed_files" ]; then
        status_info+="\n**Renamed files:**\n"
        status_info+=$(echo "$renamed_files" | sed 's/^/- /')
        status_info+="\n"
    fi

    echo -e "$status_info"
}

# Function to detect change type and generate semantic message
generate_semantic_message() {
    local staged_files=$(git diff --name-only --cached)
    local changes_summary=($(get_file_changes_summary))
    local total_count=${changes_summary[0]}
    local new_count=${changes_summary[1]}
    local modified_count=${changes_summary[2]}
    local deleted_count=${changes_summary[3]}
    local renamed_count=${changes_summary[4]}

    if [ $total_count -eq 0 ]; then
        echo "🔧 chore: update repository"
        return 1
    fi

    # Analyze changes
    local change_types=$(git diff --name-only --cached | xargs -I {} git diff --cached --name-status {} | cut -f1 | sort | uniq)

    # Determine semantic type
    local semantic_type="chore"
    local emoji="🔧"

    # Check for new features (new files with significant code)
    if echo "$staged_files" | grep -q -E 'src/|lib/|app/|main/|components/|pages/|views/'; then
        if echo "$change_types" | grep -q '^A'; then
            semantic_type="feat"
            emoji=${GITMOJI["feat"]}
        fi
    fi

    # Check for bug fixes (changes to existing files with fix-related patterns)
    if echo "$change_types" | grep -q '^M'; then
        if echo "$staged_files" | grep -q -E 'fix|bug|error|issue' || git diff --cached | grep -q -i -E 'fix|bug|error|issue'; then
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
    if echo "$staged_files" | grep -q -E '\.css$|\.scss$|\.less$|\.styl$|style'; then
        semantic_type="style"
        emoji=${GITMOJI["style"]}
    fi

    # Check for test changes
    if echo "$staged_files" | grep -q -E 'test/|spec/|__tests__|\.test\.|\.spec\.'; then
        semantic_type="test"
        emoji=${GITMOJI["test"]}
    fi

    # Check for build/config changes
    if echo "$staged_files" | grep -q -E 'webpack|rollup|vite|package\.json|yarn\.lock|package-lock\.json|tsconfig|babel|eslint'; then
        semantic_type="build"
        emoji=${GITMOJI["build"]}
    fi

    # Generate files summary for short description
    local files_summary=""
    if [ $total_count -eq 1 ]; then
        files_summary="1 file"
    else
        files_summary="$total_count files"
    fi

    # Generate detailed summary
    local summary_parts=()
    [ $new_count -gt 0 ] && summary_parts+=("$new_count new")
    [ $modified_count -gt 0 ] && summary_parts+=("$modified_count modified")
    [ $deleted_count -gt 0 ] && summary_parts+=("$deleted_count deleted")
    [ $renamed_count -gt 0 ] && summary_parts+=("$renamed_count renamed")

    local detailed_summary=$(
        IFS=', '
        echo "${summary_parts[*]}"
    )

    # Generate short description based on type
    local short_desc=""
    case $semantic_type in
    "feat")
        if [ $new_count -gt 0 ]; then
            short_desc="add new features ($files_summary)"
        else
            short_desc="implement new functionality ($files_summary)"
        fi
        ;;
    "fix")
        short_desc="resolve issues ($files_summary)"
        ;;
    "docs")
        short_desc="update documentation ($files_summary)"
        ;;
    "style")
        short_desc="improve code styling ($files_summary)"
        ;;
    "test")
        short_desc="update tests ($files_summary)"
        ;;
    "build")
        short_desc="update build configuration ($files_summary)"
        ;;
    *)
        short_desc="update codebase ($files_summary)"
        ;;
    esac

    # Generate detailed file status
    local file_status=$(generate_file_status)

    # Combine messages with enhanced format
    local commit_header="${emoji} ${semantic_type}: ${short_desc}"
    local commit_details=""

    if [ -n "$detailed_summary" ]; then
        commit_details="Changes: $detailed_summary"
    fi

    echo -e "${commit_header}\n\n${commit_details}${file_status}"
}

# Function to update CHANGELOG.md with enhanced format
update_changelog() {
    local commit_message="$1"
    local changelog_file="CHANGELOG.md"

    if [ ! -f "$changelog_file" ]; then
        return
    fi

    # Extract commit type and message
    local commit_line=$(echo "$commit_message" | head -n1)
    local commit_type=$(echo "$commit_line" | grep -o -E '^[^:]*: ' | sed 's/: $//' | sed 's/^[^a-z]*//')
    local commit_desc=$(echo "$commit_line" | sed -E 's/^[^:]+: //')

    # Get GitLab user
    local gitlab_user=$(get_gitlab_user)

    # Create changelog entry with new format
    local changelog_entry="[$commit_type] $commit_desc by @$gitlab_user"

    # Map commit type to changelog section
    case $commit_type in
    "feat") local section="### Added" ;;
    "fix") local section="### Fixed" ;;
    "docs") local section="### Documentation" ;;
    "style") local section="### Style" ;;
    "refactor") local section="### Refactored" ;;
    "test") local section="### Testing" ;;
    "build") local section="### Build" ;;
    "ci") local section="### CI_CD" ;;
    "perf") local section="### Performance" ;;
    *) local section="### Changed" ;;
    esac

    # Update CHANGELOG.md
    if grep -q "## \[Unreleased\]" "$changelog_file"; then
        # Check if section exists using fixed string search
        if ! grep -A 20 "## \[Unreleased\]" "$changelog_file" | grep -Fq "$section"; then
            # Section doesn't exist, add it after Unreleased
            sed -i "/## \[Unreleased\]/a\\$section\n- $changelog_entry" "$changelog_file"
        else
            # Section exists, append to it using fixed string search
            sed -i "/$(echo "$section" | sed 's/[[\.*^$()+?{|]/\\&/g')/a\\- $changelog_entry" "$changelog_file"
        fi
    else
        # Create new Unreleased section with current month
        local current_date=$(date +"%B %Y")
        echo -e "## [Unreleased] - $current_date\n$section\n- $changelog_entry\n\n$(cat "$changelog_file")" >"$changelog_file"
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
        git commit --allow-empty -m "${GITMOJI["chore"]} chore: initial commit"
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

    # Get current branch early for main branch check
    current_branch=$(git rev-parse --abbrev-ref HEAD)

    # Skip MR creation for main branch
    if [ "$current_branch" = "main" ]; then
        echo -e "${YELLOW}⚠️  No MR will be created for main branch${NC}"

        if [ -z "$commit_message" ]; then
            commit_message=$(generate_semantic_message)
            if [ $? -ne 0 ]; then
                echo -e "${RED}❌ No changes to commit${NC}"
                exit 1
            fi
            short_msg=$(echo "$commit_message" | head -n1)
            emoji=$(echo "$short_msg" | grep -o -E '✨|🐛|📝|💄|♻️|✅|🔧|👷|⚙️|⚡|⏪')
            echo -e "${YELLOW}📝 Auto-generated commit message: ${PURPLE}$short_msg ${emoji}${NC}"
        fi

        # Resto del proceso normal de commit/push sin MR
        update_changelog "$(echo "$commit_message" | head -n1)"
        echo -e "${GREEN}💾 Creating commit...${NC}"
        local md_body=$(echo "$commit_message" | tail -n +3 | sed 's/^\*\*/*/g')
        git commit -m "$(echo "$commit_message" | head -n1)" -m "$md_body"

        echo -e "${GREEN}📤 Pushing to ${CYAN}main${GREEN}...${NC}"
        git push origin main
        return # Exit early for main branch
    fi

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
    local md_body=$(echo "$commit_message" | tail -n +3 | sed 's/^\*\*/*/g')
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
            exit 0
        elif [[ $remote_url == *"github"* ]]; then
            pr_url="https://${remote_url}/compare/${current_branch}?expand=1"
            echo -e "${CYAN}🔗 Opening Pull Request...${NC}"
            xdg-open "$pr_url" 2>/dev/null || open "$pr_url" 2>/dev/null || start "$pr_url" 2>/dev/null
            exit 0
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

    # Check for uncommitted changes
    if ! git diff --quiet || ! git diff --cached --quiet; then
        echo -e "${YELLOW}⚠️  Uncommitted changes detected${NC}"
        read -p "Do you want to commit these changes before merging? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git add .
            commit_message="${GITMOJI["chore"]} chore: auto-commit before merge"
            git commit -m "$commit_message"
        fi
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

    echo -e "${GREEN}🔥 Pulling latest changes...${NC}"
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
    exit 0
}

# Main execution
if [ "$1" != "-h" ]; then
    # Show archive status message only if there are files to process
    if [ -f "CHANGELOG.md" ] && is_changelog_from_previous_month; then
        echo -e "${BLUE}📅 Checking changelog rotation...${NC}"
    fi
    init_changelog
fi

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
    if [[ -z "$2" ]]; then
        git status
    else
        shift
        start_branch "$@"
    fi
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

#!/usr/bin/env bash
set -euo pipefail

# Conventional Commits -> Semantic Versioning release script
#
# Commit types:
#   feat:     -> minor version bump
#   fix:      -> patch version bump
#   refactor: -> patch version bump
#   docs:     -> patch version bump
#   style:    -> patch version bump
#   perf:     -> patch version bump
#   test:     -> patch version bump
#   chore:    -> patch version bump
#   BREAKING CHANGE: or feat!:/fix!: -> major version bump

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the latest tag
get_latest_tag() {
    git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"
}

# Parse version components from tag
parse_version() {
    local version="${1#v}"
    IFS='.' read -r major minor patch <<< "$version"
    echo "$major $minor $patch"
}

# Determine bump type from commits since last tag
determine_bump() {
    local last_tag="$1"
    local commits

    if [[ "$last_tag" == "v0.0.0" ]]; then
        commits=$(git log --format="%s%n%b" HEAD)
    else
        commits=$(git log --format="%s%n%b" "${last_tag}..HEAD")
    fi

    # Check for breaking changes (major bump)
    if echo "$commits" | grep -qE "^BREAKING CHANGE:|^[a-z]+(\(.+\))?!:"; then
        echo "major"
        return
    fi

    # Check for features (minor bump)
    if echo "$commits" | grep -qE "^feat(\(.+\))?:"; then
        echo "minor"
        return
    fi

    # Check for any conventional commit (patch bump)
    if echo "$commits" | grep -qE "^(fix|refactor|docs|style|perf|test|chore)(\(.+\))?:"; then
        echo "patch"
        return
    fi

    echo "none"
}

# Calculate new version
calculate_new_version() {
    local major="$1"
    local minor="$2"
    local patch="$3"
    local bump="$4"

    case "$bump" in
        major)
            echo "v$((major + 1)).0.0"
            ;;
        minor)
            echo "v${major}.$((minor + 1)).0"
            ;;
        patch)
            echo "v${major}.${minor}.$((patch + 1))"
            ;;
        *)
            echo ""
            ;;
    esac
}

# Generate release notes from commits
generate_release_notes() {
    local last_tag="$1"
    local commits

    if [[ "$last_tag" == "v0.0.0" ]]; then
        commits=$(git log --format="%s" HEAD)
    else
        commits=$(git log --format="%s" "${last_tag}..HEAD")
    fi

    local features=""
    local fixes=""
    local other=""

    while IFS= read -r line; do
        if [[ "$line" =~ ^feat(\(.+\))?:\ (.+) ]]; then
            features+="- ${BASH_REMATCH[2]}"$'\n'
        elif [[ "$line" =~ ^fix(\(.+\))?:\ (.+) ]]; then
            fixes+="- ${BASH_REMATCH[2]}"$'\n'
        elif [[ "$line" =~ ^(refactor|docs|style|perf|test|chore)(\(.+\))?:\ (.+) ]]; then
            other+="- ${BASH_REMATCH[3]}"$'\n'
        fi
    done <<< "$commits"

    local notes=""
    [[ -n "$features" ]] && notes+="### Features"$'\n'"$features"$'\n'
    [[ -n "$fixes" ]] && notes+="### Fixes"$'\n'"$fixes"$'\n'
    [[ -n "$other" ]] && notes+="### Other"$'\n'"$other"$'\n'

    echo "$notes"
}

# Main
main() {
    local dry_run=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --dry-run|-n)
                dry_run=true
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [--dry-run|-n]"
                echo ""
                echo "Creates a new semantic version tag based on conventional commits."
                echo ""
                echo "Options:"
                echo "  --dry-run, -n  Show what would be done without creating tag"
                echo "  --help, -h     Show this help message"
                exit 0
                ;;
            *)
                echo -e "${RED}Unknown option: $1${NC}"
                exit 1
                ;;
        esac
    done

    # Ensure we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        echo -e "${RED}Error: Not a git repository${NC}"
        exit 1
    fi

    # Ensure working directory is clean
    if [[ -n $(git status --porcelain) ]]; then
        echo -e "${RED}Error: Working directory is not clean. Commit or stash changes first.${NC}"
        exit 1
    fi

    local last_tag
    last_tag=$(get_latest_tag)
    echo -e "Latest tag: ${YELLOW}${last_tag}${NC}"

    read -r major minor patch <<< "$(parse_version "$last_tag")"

    local bump
    bump=$(determine_bump "$last_tag")

    if [[ "$bump" == "none" ]]; then
        echo -e "${YELLOW}No conventional commits found since ${last_tag}. Nothing to release.${NC}"
        exit 0
    fi

    local new_version
    new_version=$(calculate_new_version "$major" "$minor" "$patch" "$bump")

    echo -e "Bump type: ${YELLOW}${bump}${NC}"
    echo -e "New version: ${GREEN}${new_version}${NC}"
    echo ""

    local notes
    notes=$(generate_release_notes "$last_tag")

    if [[ -n "$notes" ]]; then
        echo "Release notes:"
        echo "$notes"
    fi

    if [[ "$dry_run" == true ]]; then
        echo -e "${YELLOW}Dry run - no tag created${NC}"
        exit 0
    fi

    # Create annotated tag
    git tag -a "$new_version" -m "$new_version"$'\n\n'"$notes"

    echo -e "${GREEN}Created tag ${new_version}${NC}"
    echo ""
    echo "To push the tag, run:"
    echo "  git push origin ${new_version}"
}

main "$@"

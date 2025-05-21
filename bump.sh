#!/bin/bash
# Script to increment a git version tag and push it to the remote repository

# Default values
VERSION_PART="patch"
TAG_MESSAGE="new release"
WHAT_IF=false

# Help function
function show_help {
  echo "Usage: $0 [options]"
  echo "Increments a git version tag and pushes it to the remote repository."
  echo ""
  echo "Options:"
  echo "  -p, --part PART    Specify which part of the version to increment (major, minor, patch). Default: patch"
  echo "  -m, --message MSG  Specify the tag message. Default: 'new release'"
  echo "  -w, --whatif       Show what would happen without making changes"
  echo "  -h, --help         Show this help message"
  echo ""
  echo "Examples:"
  echo "  $0                 # Increments patch version"
  echo "  $0 -p minor        # Increments minor version"
  echo "  $0 -p major -m 'Major release with breaking changes'  # Increments major version with custom message"
  exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -p|--part)
      VERSION_PART="$2"
      if [[ ! "$VERSION_PART" =~ ^(major|minor|patch)$ ]]; then
        echo "Error: Version part must be 'major', 'minor', or 'patch'"
        exit 1
      fi
      shift 2
      ;;
    -m|--message)
      TAG_MESSAGE="$2"
      shift 2
      ;;
    -w|--whatif)
      WHAT_IF=true
      shift
      ;;
    -h|--help)
      show_help
      ;;
    *)
      echo "Unknown option: $1"
      show_help
      ;;
  esac
done

# Get the latest tag
if ! VERSION=$(git describe --tags --abbrev=0 2>/dev/null); then
  echo "Error: Failed to retrieve tags. Make sure you have at least one tag in your repository."
  exit 1
fi

echo "Current version: $VERSION"

# Validate version format
if ! [[ $VERSION =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
  echo "Error: Invalid version format. Expected format: x.y.z where x, y, and z are numbers."
  exit 1
fi

# Parse version components
MAJOR="${BASH_REMATCH[1]}"
MINOR="${BASH_REMATCH[2]}"
PATCH="${BASH_REMATCH[3]}"

# Increment version according to semantic versioning rules
case $VERSION_PART in
  "major")
    echo "Incrementing major version"
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  "minor")
    echo "Incrementing minor version"
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  "patch")
    echo "Incrementing patch version"
    PATCH=$((PATCH + 1))
    ;;
esac

NEW_VERSION="$MAJOR.$MINOR.$PATCH"
echo "New version: $NEW_VERSION"

# Exit if in whatif mode
if [ "$WHAT_IF" = true ]; then
  echo "WhatIf: Would create tag $NEW_VERSION with message '$TAG_MESSAGE'"
  echo "WhatIf: Would push tag $NEW_VERSION to origin"
  exit 0
fi

# Create new tag
echo "Creating new tag..."
if ! git tag -a "$NEW_VERSION" -m "$TAG_MESSAGE"; then
  echo "Error: Failed to create tag"
  exit 1
fi

# Push new tag
echo "Pushing new tag to origin..."
if ! git push origin "$NEW_VERSION"; then
  echo "Error: Failed to push tag"
  exit 1
fi

echo -e "\e[32mSuccessfully created and pushed tag $NEW_VERSION\e[0m"

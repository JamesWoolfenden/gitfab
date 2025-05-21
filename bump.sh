#!/bin/bash

# Get the latest tag
version=$(git describe --tags --abbrev=0)

# Split the version string by dots
IFS='.' read -ra splitter <<< "$version"

# Increment the build/patch number
build=$((${splitter[2]} + 1))

# Create the new version string
newVersion="${splitter[0]}.${splitter[1]}.$build"

# Display the new version
echo "$newVersion"

# Create a new tag
git tag -a "$newVersion" -m "new release"

# Push the tag to origin
git push origin "$newVersion"

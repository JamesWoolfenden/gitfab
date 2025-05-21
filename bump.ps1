<#
.SYNOPSIS
    Increments a git version tag and pushes it to the remote repository.
.DESCRIPTION
    This script retrieves the latest git tag, increments the specified version part (major, minor, or patch),
    creates a new tag, and pushes it to the remote repository.
.PARAMETER VersionPart
    Specifies which part of the version to increment. Valid values are "major", "minor", or "patch".
    Default is "patch".
.PARAMETER TagMessage
    Specifies the message to use for the new tag. Default is "new release".
.PARAMETER WhatIf
    Shows what would happen if the script runs without actually creating or pushing tags.
.EXAMPLE
    .\bump.ps1
    Increments the patch version and pushes the new tag.
.EXAMPLE
    .\bump.ps1 -VersionPart "minor"
    Increments the minor version, resets the patch version to 0, and pushes the new tag.
.EXAMPLE
    .\bump.ps1 -VersionPart "major" -TagMessage "Major release with breaking changes"
    Increments the major version, resets minor and patch to 0, and uses a custom tag message.
#>

param (
    [Parameter(HelpMessage="Which part of the version to increment")]
    [ValidateSet("major", "minor", "patch")]
    [string]$VersionPart = "patch",

    [Parameter(HelpMessage="Message to use for the new tag")]
    [string]$TagMessage = "new release",

    [Parameter(HelpMessage="Show what would happen without making changes")]
    [switch]$WhatIf
)

try {
    $version = $(git describe --tags --abbrev=0)
    Write-Host "Current version: $version"
} catch {
    Write-Error "Failed to retrieve tags. Make sure you have at least one tag in your repository."
    exit 1
}

# Validate version format
$splitter = $version.split(".")
if ($splitter.Length -ne 3) {
    Write-Error "Invalid version format. Expected format: x.y.z"
    exit 1
}

# Parse version components
try {
    $major = [int]($splitter[0])
    $minor = [int]($splitter[1])
    $patch = [int]($splitter[2])
} catch {
    Write-Error "Failed to parse version components as integers. Make sure your version follows the format x.y.z where x, y, and z are numbers."
    exit 1
}

# Increment version according to semantic versioning rules
switch ($VersionPart) {
    "major" {
        $major++
        $minor = 0
        $patch = 0
        Write-Host "Incrementing major version"
    }
    "minor" {
        $minor++
        $patch = 0
        Write-Host "Incrementing minor version"
    }
    "patch" {
        $patch++
        Write-Host "Incrementing patch version"
    }
}

$newVersion = "$major.$minor.$patch"
Write-Host "New version: $newVersion"

if ($WhatIf) {
    Write-Host "WhatIf: Would create tag $newVersion with message '$TagMessage'"
    Write-Host "WhatIf: Would push tag $newVersion to origin"
    exit 0
}

# Create new tag
Write-Host "Creating new tag..."
$tagResult = git tag -a $newVersion -m $TagMessage

if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to create tag: $tagResult"
    exit 1
}

# Push new tag
Write-Host "Pushing new tag to origin..."
$pushResult = git push origin $newVersion
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to push tag: $pushResult"
    exit 1
}

Write-Host "Successfully created and pushed tag $newVersion" -ForegroundColor Green

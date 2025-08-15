# Bump

Bump is a CLI tool designed to help you manage the versioning of your project using semantic versioning. It allows you to bump the version of your project and tag the repository accordingly. This tool operates directly within your project's Git repository.

## Features

- **Bump Patch Version**: Increment the patch version for small fixes.
- **Bump Minor Version**: Increment the minor version for backward-compatible feature additions.
- **Bump Major Version**: Increment the major version for breaking changes.
- **Add Suffix to Version**: Add a custom suffix to the version, allowing for pre-release and build metadata.
- **Dry Run Mode**: Preview version changes without making actual modifications.
- **File Updates**: Automatically update version constants in Go source files.
- **Per-Repository Push Preferences**: Configure default push behavior per repository.

## Installation

Before you can use Bump, you need to ensure that you have a Go environment set up and that your project is using Git for version control.

### Option 1: Install directly (Recommended)

```sh
go install github.com/klauern/bump/cmd/bump@latest
```

### Option 2: Build from source

1. Clone the repository where `bump` is hosted.
2. Navigate into the directory and build the tool using Go:

    ```sh
    go build ./cmd/bump
    ```

3. You can now use `./bump` from the command line within your project's directory.

## Usage

After installation, you can use `bump` by running the executable followed by the specific version bump you want to apply.

### Basic Commands

```sh
bump patch              # Bump the patch version (creates tag, does not push)
bump patch --push       # Bump the patch version and push the tag to remote
bump minor              # Bump the minor version (creates tag, does not push)
bump minor --push       # Bump the minor version and push the tag to remote
bump major --suffix rc1 # Bump the major version with a suffix (creates tag, does not push)
bump major --suffix rc1 --push # Bump the major version with a suffix and push the tag
bump push               # Push all tags to remote (can be run separately)
```

### Command Aliases

For convenience, you can use short aliases:
- `bump p` (alias for `patch`)
- `bump m` (alias for `minor`)  
- `bump M` (alias for `major`)

### Additional Options

```sh
# Preview changes without making them
bump patch --dry-run

# Update a Go source file with the next development version
bump minor --update-file version.go

# Combine options
bump major --suffix rc1 --push --dry-run
```

If you do not specify the `--push` flag, the tool will print the command you should run to push the tag manually (e.g., `git push --tags`).

## Configuration

### Per-Repository Default Push Preference

You can set a default for whether tags are pushed after bumping, on a per-repository basis. This preference is stored in your repo's `.git/config`.

To set the default to push after bumping:

```sh
bump config --default-push
```

To set the default to NOT push after bumping:

```sh
bump config --default-push=false
```

If you do not specify `--push` on the command line, the tool will use the repository default. If neither is set, it will not push by default.

### File Updates

The `--update-file` option allows you to automatically update version constants in Go source files after creating a tag. This is useful for keeping development versions in sync:

```sh
bump patch --update-file version.go
```

This will:
1. Create the version tag
2. Parse the specified Go file 
3. Find a `const Version = "..."` declaration
4. Update it with the next development version (e.g., `1.2.4-dev`)
5. Commit the change automatically

## How It Works

1. **Git Repository Detection**: The tool automatically detects and opens your project's Git repository by walking up the directory tree to find the `.git` folder.
2. **Fetching Tags**: It fetches all existing semantic version tags from the repository to determine the latest version.
3. **Version Calculation**: Based on the command (major, minor, patch) and optional suffix, it calculates the next semantic version following SemVer rules.
4. **Tag Creation**: Creates a new Git tag locally with the calculated version.
5. **Optional Operations**:
   - **Push**: If `--push` is specified or configured as default, pushes the tag to the remote repository
   - **File Update**: If `--update-file` is specified, updates Go source files with development versions
   - **Dry Run**: If `--dry-run` is specified, shows what would happen without making changes

The tool uses Git's native tagging system and follows semantic versioning (SemVer) conventions with tags formatted as `v1.2.3` or `v1.2.3-suffix`.

## Contributing

We welcome contributions to `bump`, whether it's improving documentation, adding new features, or reporting bugs, your contributions are greatly appreciated.

## License

This project is licensed under the [MIT License](LICENSE). Feel free to clone, modify, and use it in your own projects.

## Disclaimer

The `bump` is provided "as is", without warranty of any kind. Use at your own risk.

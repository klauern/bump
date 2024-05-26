# Bump (`bump`)

Bump is a CLI tool designed to help you manage the versioning of your project using semantic versioning. It allows you to bump the version of your project and tag the repository accordingly. This tool operates directly within your project's Git repository.

## Features

- **Bump Patch Version**: Increment the patch version for small fixes.
- **Bump Minor Version**: Increment the minor version for backward-compatible feature additions.
- **Bump Major Version**: Increment the major version for breaking changes.
- **Add Suffix to Version**: Add a custom suffix to the version, allowing for pre-release and build metadata.

## Installation

Before you can use Bump, you need to ensure that you have a Go environment set up and that your project is using Git for version control.

1. Clone the repository where `bump` is hosted.
2. Navigate into the directory and build the tool using Go:

    ```sh
    go build -o version-bumper
    ```

3. You can now use `version-bumper` from the command line within your project's directory.

## Usage

After you have successfully built the `version-bumper`, you can use it by running the executable followed by the specific version bump you want to apply.

```sh
./bump patch              # Bump the patch version
./bump minor              # Bump the minor version
./bump major --suffix rc1 # Bump the major version (all commands support optional --suffix flag)
```

## How It Works

1. **Opening the Git Repository**: The tool automatically detects and opens your project's Git repository.
2. **Fetching Tags**: It fetches all the tags from the repository to determine the latest version.
3. **Determining and Creating the Next Version**: Based on the command (major, minor, patch, or suffix), it calculates the next version, creates a new tag, and pushes this tag to the remote repository.

## Contributing

We welcome contributions to `bump`, whether it's improving documentation, adding new features, or reporting bugs, your contributions are greatly appreciated.

## License

This project is licensed under the [MIT License](LICENSE). Feel free to clone, modify, and use it in your own projects.

## Disclaimer

The `bump` is provided "as is", without warranty of any kind. Use at your own risk.

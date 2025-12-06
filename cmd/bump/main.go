package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/klauern/bump"
	"github.com/urfave/cli/v2"
)

func init() {
	if os.Getenv("DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	app := &cli.App{
		Name:  "bump",
		Usage: "Bump the version of your project",
		Commands: []*cli.Command{
			createCommand("patch", "p", "Bump the patch version"),
			createCommand("minor", "m", "Bump the minor version"),
			createCommand("major", "M", "Bump the major version"),
			{
				Name:  "push",
				Usage: "Push tags to remote",
				Action: func(c *cli.Context) error {
					if err := bump.PushTag(); err != nil {
						return fmt.Errorf("failed to push tags: %v", err)
					}
					fmt.Println("Successfully pushed tags to remote.")
					return nil
				},
			},
			{
				Name:  "config",
				Usage: "Configure bump settings for this repo",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "default-push",
						Usage: "Set default to push tags after bumping",
					},
				},
				Action: func(c *cli.Context) error {
					repoPath, err := findGitRoot(".")
					if err != nil {
						return fmt.Errorf("failed to find git root: %v", err)
					}
					if c.IsSet("default-push") {
						val := c.Bool("default-push")
						err := bump.SetDefaultPushPreference(repoPath, val)
						if err != nil {
							return fmt.Errorf("failed to set default push: %v", err)
						}
						fmt.Printf("Set default push to %v for this repo.\n", val)
						return nil
					}
					return cli.ShowSubcommandHelp(c)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func createCommand(name, alias, usage string) *cli.Command {
	return &cli.Command{
		Name:    name,
		Aliases: []string{alias},
		Usage:   usage,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "suffix",
				Usage: "Add a suffix to the version",
			},
			&cli.StringFlag{
				Name:  "update-file",
				Usage: "Update a file with the next dev version",
			},
			&cli.BoolFlag{
				Name:  "push",
				Usage: "Push the tag to remote after creating it",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Show what version would be created without making changes",
			},
		},
		Action: func(c *cli.Context) error {
			pushFlag := c.Bool("push")
			pushSet := c.IsSet("push")
			repoPath, err := findGitRoot(".")
			if err != nil {
				return fmt.Errorf("failed to find git root: %v", err)
			}
			var doPush bool
			if pushSet {
				doPush = pushFlag
			} else {
				// Not set on CLI, check repo default
				val, isSet, err := bump.GetDefaultPushPreference(repoPath)
				if err == nil && isSet {
					doPush = val // Use explicitly configured value
				} else {
					doPush = false // Use default (false) when not configured or error
				}
			}
			return bumpVersion(name, c.String("suffix"), c.String("update-file"), doPush, c.Bool("dry-run"))
		},
	}
}

// findGitRoot walks up the directory tree from the given startPath until it finds a .git directory.
// If no .git directory is found, it returns an error.
func findGitRoot(startPath string) (string, error) {
	log.Debug("Find Git Root", "startPath", startPath)
	currentPath := startPath
	for {
		if _, err := os.Stat(filepath.Join(currentPath, ".git")); err == nil {
			log.Debug(".git found", "path", currentPath)
			return currentPath, nil
		}

		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			log.Error("no .git directory found")
			return "", fmt.Errorf("no .git directory found")
		}

		currentPath = parentPath
	}
}

// bumpVersion bumps the version using the BumpService.
func bumpVersion(bumpType, suffix, updateFile string, doPush, dryRun bool) error {
	// Find git root
	repoPath, err := findGitRoot(".")
	if err != nil {
		return fmt.Errorf("failed to find git root: %v", err)
	}

	// Open repository
	repo, err := NewGoGitRepository(repoPath)
	if err != nil {
		return err
	}

	// Create service
	svc := NewBumpService(repo, nil, os.Stdout)

	// Build options
	opts := BumpOptions{
		BumpType:   bumpType,
		Suffix:     suffix,
		UpdateFile: updateFile,
		Push:       doPush,
		DryRun:     dryRun,
	}

	// Execute bump
	_, err = svc.Bump(opts)
	return err
}

// validateFilePath performs comprehensive validation to prevent path traversal attacks
func validateFilePath(filePath, repoPath string) error {
	// Check for empty or whitespace-only paths
	if strings.TrimSpace(filePath) == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Check for suspicious patterns that indicate path traversal attempts
	suspiciousPatterns := []string{
		"..",           // Directory traversal
		"\x00",         // Null byte injection
		"\r",           // Carriage return
		"\n",           // Newline injection
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(filePath, pattern) {
			return fmt.Errorf("path contains invalid characters")
		}
	}

	// Clean the path and resolve to absolute path
	cleanPath := filepath.Clean(filePath)
	
	// Prevent paths that would resolve outside the working directory
	if filepath.IsAbs(cleanPath) {
		return fmt.Errorf("absolute paths are not allowed")
	}

	// Resolve to absolute path for boundary checking
	absPath, err := filepath.Abs(filepath.Join(repoPath, cleanPath))
	if err != nil {
		return fmt.Errorf("unable to resolve path")
	}

	// Get canonical repository path
	repoAbsPath, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("unable to resolve repository path")
	}

	// Resolve symlinks to prevent symlink attacks
	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If symlink resolution fails, use the original path but ensure it exists within bounds
		resolvedPath = absPath
	}

	resolvedRepoPath, err := filepath.EvalSymlinks(repoAbsPath)
	if err != nil {
		resolvedRepoPath = repoAbsPath
	}

	// Ensure the resolved path is within repository boundaries
	if !strings.HasPrefix(resolvedPath, resolvedRepoPath+string(filepath.Separator)) &&
		resolvedPath != resolvedRepoPath {
		return fmt.Errorf("file path must be within repository")
	}

	// Additional check: ensure the path doesn't contain relative components after cleaning
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains invalid components")
	}

	return nil
}

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5"
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
			createCommand("patch", "Bump the patch version"),
			createCommand("minor", "Bump the minor version"),
			createCommand("major", "Bump the major version"),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func createCommand(name, usage string) *cli.Command {
	return &cli.Command{
		Name:  name,
		Usage: usage,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "suffix",
				Usage: "Add a suffix to the version",
			},
		},
		Action: func(c *cli.Context) error {
			return bumpVersion(name, c.String("suffix"))
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

// bumpVersion bumps the version of a project's .git directory to the next semantic version passed in as a string.
func bumpVersion(bumpType, suffix string) error {
	repoPath, err := findGitRoot(".")
	if err != nil {
		return fmt.Errorf("failed to find git root: %v", err)
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open git repo: %v", err)
	}

	tagRefs, err := repo.Tags()
	if err != nil {
		return fmt.Errorf("failed to fetch tags: %v", err)
	}

	latestTag, err := bump.GetLatestTag(tagRefs)
	if err != nil {
		return fmt.Errorf("failed to determine latest tag: %v", err)
	}

	var nextTag string
	if latestTag != "" {
		nextTag, err = bump.GetNextTag(latestTag, bumpType, suffix)
		if err != nil {
			return fmt.Errorf("failed to determine next tag: %v", err)
		}
	} else {
		fmt.Println("No tags found, starting at v0.1.0")
		nextTag = "v0.1.0"
	}

	err = bump.TagAndPush(repoPath, nextTag)
	if err != nil {
		return fmt.Errorf("failed to tag and push: %v", err)
	}

	fmt.Printf("Successfully created and pushed tag %s\n", nextTag)
	return nil
}

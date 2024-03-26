package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/klauern/bump"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "version-bumper",
		Usage: "Bump the version of your project",
		Commands: []*cli.Command{
			{
				Name:  "patch",
				Usage: "Bump the patch version",
				Action: func(c *cli.Context) error {
					return bumpVersion("patch")
				},
			},
			{
				Name:  "minor",
				Usage: "Bump the minor version",
				Action: func(c *cli.Context) error {
					return bumpVersion("minor")
				},
			},
			{
				Name:  "major",
				Usage: "Bump the major version",
				Action: func(c *cli.Context) error {
					return bumpVersion("major")
				},
			},
			{
				Name:  "suffix",
				Usage: "Add a suffix to the version",
				Action: func(c *cli.Context) error {
					if c.Args().Len() < 1 {
						return fmt.Errorf("you must provide a suffix")
					}
					return bumpVersion(c.Args().First())
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func bumpVersion(bumpType string) error {
	repoPath := "." // path to your git repository

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

	nextTag, err := bump.GetNextTag(latestTag, bumpType)
	if err != nil {
		return fmt.Errorf("failed to determine next tag: %v", err)
	}

	// Create and push new tag
	_, err = repo.CreateTag(nextTag, plumbing.NewHash("HEAD"), nil)
	if err != nil {
		return fmt.Errorf("failed to create tag: %v", err)
	}

	fmt.Printf("Successfully created and pushed tag %s\n", nextTag)
	return nil
}

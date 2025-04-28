package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
				val, err := bump.GetDefaultPushPreference(repoPath)
				if err == nil {
					doPush = val
				} else {
					doPush = false // fallback default
				}
			}
			return bumpVersion(name, c.String("suffix"), c.String("update-file"), doPush)
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
func bumpVersion(bumpType, suffix, updateFile string, doPush bool) error {
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

	err = bump.CreateTag(nextTag)
	if err != nil {
		return fmt.Errorf("failed to create tag: %v", err)
	}

	if doPush {
		err = bump.PushTag()
		if err != nil {
			return fmt.Errorf("failed to push tag: %v", err)
		}
		fmt.Printf("Successfully created and pushed tag %s\n", nextTag)
	} else {
		fmt.Printf("Successfully created tag %s. To push, run: git push --tags\n", nextTag)
	}

	if updateFile != "" {
		err = updateVersionFile(updateFile, nextTag)
		if err != nil {
			return fmt.Errorf("failed to update file: %v", err)
		}
	}
	return nil
}

func updateVersionFile(filePath, nextTag string) error {
	// Parse the next tag
	nextVersion, ok := bump.ParseTagVersion(nextTag)
	if !ok {
		return fmt.Errorf("failed to parse next tag: %s", nextTag)
	}

	// Create the development version
	devVersion := fmt.Sprintf("%d.%d.%d-dev", nextVersion.Major, nextVersion.Minor, nextVersion.Patch+1)

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %v", err)
	}

	// Find and update the Version constant
	updated := false
	ast.Inspect(node, func(n ast.Node) bool {
		if gen, ok := n.(*ast.GenDecl); ok && gen.Tok == token.CONST {
			for _, spec := range gen.Specs {
				if value, ok := spec.(*ast.ValueSpec); ok {
					for i, ident := range value.Names {
						if ident.Name == "Version" {
							value.Values[i] = &ast.BasicLit{
								Kind:  token.STRING,
								Value: fmt.Sprintf(`"%s"`, devVersion),
							}
							updated = true
							return false
						}
					}
				}
			}
		}
		return true
	})

	if !updated {
		return fmt.Errorf("version constant not found in file")
	}

	// Write the updated AST back to the file
	var buf strings.Builder
	err = format.Node(&buf, fset, node)
	if err != nil {
		return fmt.Errorf("failed to format updated AST: %v", err)
	}

	err = os.WriteFile(filePath, []byte(buf.String()), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	// Commit the change
	cmd := exec.Command("git", "add", filePath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to stage file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("Bump version to %s", devVersion))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to commit file: %v", err)
	}

	return nil
}

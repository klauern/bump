package bump

import (
	"fmt"
	"os"
	"path/filepath"
)

func ExampleGetNextTag() {
	next, err := GetNextTag("v1.2.3", "minor", "")
	if err != nil {
		panic(err)
	}
	fmt.Println(next)

	next, err = GetNextTag("v1.2.3", "patch", "rc.1")
	if err != nil {
		panic(err)
	}
	fmt.Println(next)

	// Output:
	// v1.3.0
	// v1.2.4-rc.1
}

func ExampleParseTagVersion() {
	v, ok := ParseTagVersion("v2.10.3-beta.2")
	fmt.Println(ok)
	fmt.Println(v.Major, v.Minor, v.Patch, v.Suffix)

	// Output:
	// true
	// 2 10 3 -beta.2
}

func ExampleGetDefaultPushPreference() {
	repoPath, err := os.MkdirTemp("", "bump-example-")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.RemoveAll(repoPath)
	}()

	if err := os.MkdirAll(filepath.Join(repoPath, ".git"), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(
		filepath.Join(repoPath, ".git", "config"),
		[]byte("[bump]\n defaultPush=true\n"),
		0o644,
	); err != nil {
		panic(err)
	}

	val, isSet, err := GetDefaultPushPreference(repoPath)
	if err != nil {
		panic(err)
	}
	fmt.Println(val, isSet)

	if err := SetDefaultPushPreference(repoPath, false); err != nil {
		panic(err)
	}
	val, isSet, err = GetDefaultPushPreference(repoPath)
	if err != nil {
		panic(err)
	}
	fmt.Println(val, isSet)

	// Output:
	// true true
	// false true
}

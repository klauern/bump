// Package bump implements the core versioning and tag-management logic used by the `bump` CLI.
//
// It provides helpers for working with semantic version git tags (SemVer 2.0). Tags are
// expected to look like:
//
//	vMAJOR.MINOR.PATCH
//	vMAJOR.MINOR.PATCH-PRERELEASE
//
// Use [GetNextTag] to compute the next tag string from a current tag and bump type. Use
// [NewGitInfo] and [GetLatestTag] to discover existing tags in a repository.
//
// For repository-local configuration, [GetDefaultPushPreference] and
// [SetDefaultPushPreference] read and write the `[bump] defaultPush` setting in `.git/config`.
//
// Operations that mutate the repository via external git commands (e.g. [CreateTag] and
// [PushTag]) are serialized with a repository-scoped lock file at `.git/bump.lock`.
package bump

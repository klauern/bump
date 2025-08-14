---
id: task-6
title: Fix resource leak in ReferenceIter usage
status: Done
assignee:
  - '@myself'
created_date: '2025-08-14 04:54'
updated_date: '2025-08-14 05:09'
labels:
  - bug
  - resource-management
dependencies: []
priority: high
---

## Description

ReferenceIter objects returned by repo.Tags() are not properly closed, causing resource leaks. This affects getTags and getTagVersions functions, potentially causing memory issues in long-running processes.

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 All ReferenceIter objects are closed using defer statements,No resource leaks detected in memory profiling tests,Functions properly handle early returns without leaking resources,Integration tests verify proper cleanup under error conditions
<!-- AC:END -->

## Implementation Notes

Fixed resource leaks by adding defer tagRefs.Close() statements in two critical locations: 1) NewGitInfo function in bump.go:147 - ensures proper cleanup when getting all git tags, 2) bumpVersion function in cmd/bump/main.go:165 - ensures proper cleanup when determining latest tags. All tests pass, no functional changes to behavior, but prevents memory leaks in long-running processes.

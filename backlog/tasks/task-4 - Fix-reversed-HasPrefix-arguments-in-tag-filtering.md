---
id: task-4
title: Fix reversed HasPrefix arguments in tag filtering
status: Done
assignee:
  - '@myself'
created_date: '2025-08-14 04:54'
updated_date: '2025-08-14 05:07'
labels:
  - bug
  - high
dependencies: []
priority: high
---

## Description

Line 71 in bump.go has reversed arguments to strings.HasPrefix, causing tag filtering logic to fail. This prevents proper semantic version tag detection and may cause incorrect version calculations.

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 HasPrefix call uses correct argument order (tagRef.Name().String(), prefix),All semantic version tags are properly detected and filtered,Tag filtering tests validate correct prefix matching behavior,No regression in existing tag parsing functionality
<!-- AC:END -->

## Implementation Notes

Fixed reversed HasPrefix arguments in getVersions function. Changed from HasPrefix("v", tagRef.Name().String()) to HasPrefix(tagRef.Name().Short(), "v") on line 173. This fixes the critical tag filtering bug that was preventing proper version tag detection. All tests pass.

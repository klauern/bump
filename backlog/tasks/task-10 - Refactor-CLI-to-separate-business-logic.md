---
id: task-10
title: Refactor CLI to separate business logic
status: To Do
assignee: []
created_date: '2025-08-14 04:55'
labels:
  - architecture
  - refactoring
dependencies: []
priority: medium
---

## Description

The main.go file mixes CLI command handling with business logic (updateVersionFile function), violating separation of concerns. This makes testing difficult and reduces code maintainability.

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 updateVersionFile function moved to main bump package,CLI layer only handles argument parsing and command routing,Business logic functions are independently testable,No functional changes to CLI behavior or API
<!-- AC:END -->

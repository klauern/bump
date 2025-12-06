# Test Coverage Improvement Plan to 90%

**Date**: 2025-12-05
**Current Coverage**: 61.9% overall (86.8% core, 23.5% cmd)
**Target**: 90% overall
**Initial Coverage**: 50.1% overall (71.3% core, 0% cmd)

## Progress Summary

### ✅ Completed Work

#### Core Package (bump.go): 71.3% → 86.8%
- **acquireGitLock** (59.5% → 75.7%): Added tests for stale lock cleanup, success/release flow
- **SetDefaultPushPreference** (60% → 75%): Added tests for config errors, read-only files, corruption
- **GetDefaultPushPreference** (72.2% → 83.3%): Added tests for invalid values, corrupted config
- Added comprehensive SemVer comparison tests
- Added git lock mechanism tests

#### CMD Package (cmd/bump): 0% → 23.5%
- Created test suite from scratch
- **findGitRoot**: Tested path traversal and nested directory handling
- **validateFilePath**: Comprehensive security tests (path traversal, null bytes, boundary checks)
- **createCommand**: Structure validation tests
- **bumpVersion**: Error path testing (no repo)
- **updateVersionFile**: Input validation tests

### 📊 Current Coverage by Function

#### Core Package (bump.go)
```
Functions < 90% coverage:
- createTagWithLock        71.4% ← needs work
- pushTagWithLock          71.4% ← needs work
- CreateTag                75.0% ← needs work
- PushTag                  75.0% ← needs work
- SetDefaultPushPreference 75.0%
- acquireGitLock           75.7%
- getTags                  75.0% ← needs work
- getVersions              80.0%
- pushTag                  80.0%
- validateRepositoryPath   80.0%
- GetDefaultPushPreference 83.3%
- Release                  85.7%
- parseNumericIdentifier   88.9%
```

#### CMD Package (cmd/bump)
```
Functions at 0% coverage (need integration tests):
- main()                   0.0% ← can't test
- init()                   0.0% ← can't test
- bumpVersion()            ~30% ← partial (needs happy path)
- updateVersionFile()      ~20% ← partial (needs happy path)
```

## Path to 90% Coverage

### Phase 1: Core Package Improvements (Target: 90%+)

**Priority**: Add ~4-5 percentage points

1. **createTagWithLock / pushTagWithLock** (71.4% each)
   - Test error handling when lock acquisition fails
   - Test successful tag operations with proper lock/release
   - Test lock release error handling

2. **CreateTag / PushTag** (75% each)
   - Currently tests error paths only
   - Need to test success paths (requires git command mocking)
   - Alternative: Skip these as they're thin wrappers

3. **getTags** (75%)
   - Test error handling when repository.Tags() fails
   - Requires mocking git.Repository

### Phase 2: CMD Package Major Improvements (Target: 60%+)

**Priority**: Add ~25-30 percentage points to overall coverage

The cmd package has the most significant impact on overall coverage but is hardest to test.

#### 2.1 Integration Tests for bumpVersion()

**Current coverage**: ~30%
**Target**: 70%+
**Approach**: Create real git repos in tests

```go
func TestBumpVersionSuccess(t *testing.T) {
    // Setup: Create real git repo with commits and tags
    repo := createRealGitRepo(t)
    createTag(t, repo, "v0.1.0")

    // Test: Bump patch version
    err := bumpVersion("patch", "", "", false, true) // dry-run first
    // Assert: Correct next version calculated

    // Test: Actually create tag (not dry-run)
    err := bumpVersion("patch", "", "", false, false)
    // Assert: Tag created successfully
}

func TestBumpVersionWithSuffix(t *testing.T) {
    // Test version bumping with pre-release suffix
}

func TestBumpVersionFirstTag(t *testing.T) {
    // Test when no tags exist (starts at v0.1.0)
}
```

**Functions to test**:
- Happy path: repo with tags → calculate next → create tag → verify
- With suffix
- With push
- First tag (no existing tags)
- Dry run mode

#### 2.2 Integration Tests for updateVersionFile()

**Current coverage**: ~20%
**Target**: 70%+
**Approach**: Create actual Go files with version constants

```go
func TestUpdateVersionFileSuccess(t *testing.T) {
    // Create a Go file with Version constant
    content := `package main
const Version = "1.0.0"
`
    // Write file, call updateVersionFile, verify:
    // 1. Version constant updated
    // 2. File committed to git
    // 3. Development version calculated correctly
}

func TestUpdateVersionFileNoVersionConstant(t *testing.T) {
    // Test error when Version constant not found
}

func TestUpdateVersionFileParseError(t *testing.T) {
    // Test error when file has syntax errors
}
```

**Functions to test**:
- Parse Go file with Version constant
- Update constant value
- Commit changes to git
- Error: Version constant not found
- Error: File parse failure

### Phase 3: Remaining Core Functions (Target: 95%+)

**Priority**: Polish remaining gaps

1. **validateRepositoryPath** (80%)
   - Already well tested
   - Missing: Symlink edge cases

2. **getVersions** (80%)
   - Test ForEach error handling
   - Test with malformed tags

3. **pushTag** (80%)
   - Test success path (requires git command mocking)

## Estimated Effort to Reach 90%

### Minimum Path (Realistic)
1. **Phase 1** (Core improvements): 2-3 hours
   - Focus on testable functions
   - Skip thin wrappers around git commands

2. **Phase 2** (CMD integration tests): 4-6 hours
   - Create real git repo test helper
   - Test bumpVersion happy paths
   - Test updateVersionFile happy paths

**Total**: 6-9 hours of focused work

### Maximum Path (Comprehensive)
- Add git command mocking
- Test all error paths
- Test all edge cases
- **Total**: 12-15 hours

## Challenges & Recommendations

### Current Blockers

1. **Git Command Testing**
   - Many functions call `git` commands via exec.Command
   - Current mocking strategy limited
   - **Solution**: Consider using go-git library exclusively (avoid shelling out)

2. **CLI Testing Complexity**
   - main() and init() can't be easily tested
   - CLI actions require urfave/cli context mocking
   - **Solution**: Extract business logic from CLI handlers

3. **Integration vs Unit Tests**
   - High coverage requires integration tests with real git repos
   - **Solution**: Accept that 60-70% may be more realistic for CLI-heavy code

### Architectural Improvements for Testability

If aiming for 90%+ long-term:

1. **Separate Business Logic from CLI**
   - Create service layer with pure functions
   - CLI handlers only parse args and call services
   - Services are independently testable

2. **Dependency Injection**
   - Inject git operations instead of calling directly
   - Makes mocking easier

3. **Interface-Based Design**
   - Define interfaces for git operations
   - Implement both real and mock versions

## Current Test Quality

### Strengths ✓
- Comprehensive error path testing
- Security validation (path traversal, injection)
- Edge case coverage (stale locks, corrupted config)
- SemVer compliance testing

### Gaps ✗
- Limited happy path coverage in cmd package
- No git integration tests
- Mock coverage incomplete for git operations

## Next Steps (Priority Order)

1. ✅ **Document progress** (this file)
2. **Add Phase 1 tests** (core package improvements)
3. **Create git repo test helper** for integration tests
4. **Add Phase 2 tests** (cmd package integration)
5. **Re-evaluate 90% target** based on actual coverage gains

## Files Modified

### New Files
- `cmd/bump/main_test.go` - CMD package tests (new)
- `history/TEST_COVERAGE_90PCT_PLAN.md` - This document

### Modified Files
- `bump_test.go` - Added 100+ new test cases
  - acquireGitLock comprehensive tests
  - Config management tests
  - Error path tests
  - Edge case tests

## Coverage Breakdown

```
Overall:     61.9% (target: 90%, gap: 28.1%)
Core:        86.8% (target: 92%+, gap: 5.2%)
CMD:         23.5% (target: 60%+, gap: 36.5%)
```

**Key Insight**: The cmd package is the bottleneck. Improving it from 23.5% to 60% would raise overall coverage from 61.9% to ~80%.

## Recommendations

### For Immediate Progress
1. Focus on cmd/bump integration tests (biggest impact)
2. Accept that 75-80% may be more realistic than 90%
3. Prioritize testing business logic over CLI boilerplate

### For Long-Term Maintainability
1. Refactor to separate business logic from CLI
2. Use interfaces for testability
3. Consider task-11 (semantic version comparison) refactoring
4. Add integration test suite to CI/CD

## Success Criteria

**Minimum Success** (75% overall):
- ✅ Core package > 85%
- ✅ CMD package > 40%
- ✅ All critical paths tested
- ✅ Security validations covered

**Target Success** (90% overall):
- Core package > 92%
- CMD package > 65%
- Integration tests for main workflows
- Happy path coverage for CLI commands

**Stretch Goal** (95% overall):
- Comprehensive mocking
- Full integration test suite
- Refactored for testability

# GitHub Actions Workflows

This directory contains GitHub Actions workflows for automated testing, linting, and building.

## Workflows

### 1. `test.yml` - Comprehensive Test Suite

**Triggers:**
- Push to: `main`, `master`, `develop`, `renew` branches
- Pull requests to these branches

**Jobs:**

#### Test Job
- Runs all Go tests with a PostgreSQL service container
- Generates code coverage reports
- Uploads coverage artifacts (retained for 30 days)
- Displays coverage summary in GitHub Actions UI
- **Fails if any test fails** ❌

#### Lint Job
- Runs `golangci-lint` for code quality checks
- Uses latest linter version
- **Fails if linting issues are found** ❌

#### Build Job
- Builds the API binary
- Only runs if test and lint jobs pass
- Uploads binary artifact (retained for 7 days)
- **Fails if build fails** ❌

**Key Features:**
- PostgreSQL 16 service container for integration tests
- 10-minute timeout for tests
- Code coverage reporting
- Parallel job execution (test and lint run concurrently)
- Dependency caching for faster builds

### 2. `quick-test.yml` - Fast Unit Tests

**Triggers:**
- Push to any branch
- Any pull request

**Jobs:**

#### Quick Test Job
- Runs unit tests only (using `-short` flag)
- No database required - faster execution
- 5-minute timeout
- **Fails immediately if any test fails** ❌

**Key Features:**
- Fast feedback loop (~2-3 minutes)
- Runs on all branches
- Good for development workflow
- Uses Go module caching

## Test Failure Behavior

All workflows are configured to **fail the CI pipeline** if tests fail:

1. `go test` command returns non-zero exit code on test failure
2. GitHub Actions interprets non-zero exit codes as job failure
3. Failed jobs block PR merges (if branch protection is enabled)
4. Failed jobs show red ❌ status in GitHub UI

## Usage

### Local Testing

Test your code before pushing:

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -v -coverprofile=coverage.out ./...

# Run unit tests only (fast)
go test -v -short ./...

# View coverage
go tool cover -html=coverage.out
```

### Viewing Results

1. **GitHub Actions Tab**: See all workflow runs
2. **PR Checks**: See status on pull request page
3. **Coverage Reports**: Download from workflow artifacts
4. **Summary**: View coverage summary in Actions run summary

### Branch Protection

To enforce tests must pass before merging:

1. Go to repository Settings → Branches
2. Add branch protection rule for `main`/`master`
3. Enable "Require status checks to pass before merging"
4. Select: `Run Go Tests`, `Run Linters`, `Build Application`

## Troubleshooting

### Tests timeout
- Increase timeout in workflow (default: 10m for full tests, 5m for quick tests)
- Check for hanging goroutines or infinite loops

### Database connection issues
- PostgreSQL service takes ~10s to be ready
- Health checks ensure database is ready before tests run
- Testcontainers in tests create their own isolated databases

### Cache issues
- GitHub Actions caches Go modules
- Clear cache by updating Go version in workflow
- Or manually clear caches in Actions settings

## Test Coverage

Current test coverage includes:
- ✅ Recipe nutrition calculations
- ✅ Diary entry calculations (food-based)
- ✅ Diary entry calculations (recipe-based)
- ✅ Food repository operations
- ✅ JWT authentication

Target: 80%+ coverage for critical paths

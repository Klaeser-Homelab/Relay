# GitHub Actions CI/CD Pipeline

This directory contains GitHub Actions workflows for automated testing, building, and deployment of Relay Server.

## Workflows

### `ci.yml` - Main CI Pipeline

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` branch

**Jobs:**

#### 1. **Test Job** (`test`)
- **Matrix Strategy**: Tests against Go 1.21.x and 1.22.x
- **Platform**: Ubuntu latest
- **Steps**:
  - Checkout code
  - Set up Go environment with caching
  - Install dependencies and verify modules
  - Run linting (`go fmt`, `go vet`)
  - Build Relay binary
  - Set up mock test environment
  - Install mock Claude CLI for testing
  - Run comprehensive test suite
  - Upload test artifacts on failure

#### 2. **Build and Release Job** (`build-and-release`)
- **Depends on**: Test job must pass
- **Trigger**: Only on pushes to `main` branch
- **Matrix Strategy**: Linux, macOS, Windows
- **Outputs**: Cross-platform binaries
- **Artifacts**: Retained for 30 days

#### 3. **Security Job** (`security`)
- **Platform**: Ubuntu latest
- **Security Scanning**: 
  - SARIF security reports
  - Go vulnerability checking (`govulncheck`)
- **Continues on error**: Won't fail the build

#### 4. **Performance Job** (`performance`)
- **Depends on**: Test job must pass
- **Trigger**: Only on pushes to `main` branch
- **Measures**: Command execution timing
- **Artifacts**: Performance reports retained for 30 days

## Mock Environment for CI

Since CI environments don't have access to the real Claude CLI, we create a mock version:

### Mock Claude CLI Features:
- **Version command**: Returns expected version string
- **Print command**: Provides mock responses
- **JSON output**: Returns properly formatted JSON responses
- **Error handling**: Simulates Claude CLI error conditions

### Mock Setup:
```bash
# Created during CI workflow
cat > /tmp/claude << 'EOF'
#!/bin/bash
case "$1" in
  "--version") echo "1.0.17 (Claude Code)" ;;
  "--print") echo '{"type":"result","result":"Mock response"}' ;;
  *) echo "Mock Claude CLI"; exit 1 ;;
esac
EOF
chmod +x /tmp/claude
sudo mv /tmp/claude /usr/local/bin/claude
```

## CI-Specific Test Adaptations

### Environment Detection:
- Tests detect CI mode via `CI_MODE=true` or `CI=true` environment variables
- Uses different configurations for CI vs local development

### CI Test Modifications:
- **RelayTest Path**: Uses `/tmp/RelayTest` instead of local path
- **Database**: Uses temporary database files
- **Git Setup**: Creates minimal git repository for testing
- **Claude Integration**: Uses mock Claude CLI responses
- **Expectations**: Some tests have relaxed expectations for CI environment

### CI Configuration (`ci_config.go`):
```go
func getCITestConfig() *TestConfig {
    if isCI {
        return &TestConfig{
            RelayTestPath:   "/tmp/RelayTest",
            TestDBPath:      "/tmp/relay_ci_test.db", 
            RelayBinaryPath: "../relay",
        }
    }
    return getTestConfig() // Local config
}
```

## Test Results and Artifacts

### Success Indicators:
- ✅ All test suites pass (27/27 tests)
- ✅ Binary builds successfully
- ✅ Linting passes
- ✅ Security scans complete

### Failure Handling:
- **Test Logs**: Uploaded as artifacts for debugging
- **Database Files**: Preserved for analysis
- **Performance Data**: Baseline comparisons
- **Build Artifacts**: Cross-platform binaries

### Artifacts Retention:
- **Test logs**: 7 days
- **Release binaries**: 30 days  
- **Performance reports**: 30 days

## Local Testing with CI Configuration

To test CI behavior locally:

```bash
# Set CI mode
export CI_MODE=true
export RELAY_TEST_PATH=/tmp/RelayTest

# Set up mock environment
mkdir -p /tmp/RelayTest
cd /tmp/RelayTest && git init

# Run tests
cd relay/server/test
go run *.go
```

## Status Badges

The CI status is displayed in the main README:

```markdown
[![CI Status](https://github.com/username/relay/workflows/Relay%20Server%20CI/badge.svg)](https://github.com/username/relay/actions)
```

## Adding New Tests

When adding tests that might behave differently in CI:

1. **Check CI mode** in test logic:
   ```go
   isCI := os.Getenv("CI_MODE") == "true"
   ```

2. **Adapt expectations** for CI environment:
   ```go
   if adaptTestForCI(testName) {
       return runCIAdaptedTest(testName, originalTest)
   }
   ```

3. **Handle mock responses** for Claude integration tests

4. **Use temporary paths** for file operations

## Troubleshooting CI Issues

### Common CI Failures:

**"No module named claude"**
- Mock Claude CLI installation failed
- Check `/usr/local/bin/claude` permissions

**"RelayTest directory not found"**
- Git init failed in CI setup
- Check `/tmp/RelayTest` creation

**"Database permission denied"**
- Temp directory permissions issue
- Use `/tmp/` paths in CI mode

**"Test timeout"**
- CI environment slower than expected
- Increase timeout values for CI

### Debugging Steps:

1. **Check workflow logs** in GitHub Actions tab
2. **Download test artifacts** for detailed logs
3. **Run locally with CI mode** to reproduce
4. **Compare local vs CI test results**

## Security Considerations

- **No secrets required**: Tests use mock APIs
- **Isolated environment**: Each run uses fresh containers
- **Limited permissions**: Only read/write to designated paths
- **Artifact cleanup**: Automated retention policies

The CI pipeline ensures code quality while maintaining security and providing fast feedback for development.
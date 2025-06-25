# Relay Test Suite

This directory contains a comprehensive testing system for Relay Server to ensure all functionality works correctly after changes.

## Test Structure

### Test Files
- `test_runner.go` - Main test orchestrator and framework
- `database_test.go` - Database operation tests
- `project_test.go` - Project management tests  
- `claude_test.go` - Claude CLI integration tests
- `git_test.go` - Git operation tests
- `integration_test.go` - End-to-end workflow tests
- `cleanup.sh` - Test cleanup script
- `README.md` - This documentation

### Test Categories

#### Unit Tests
- Database operations (add/remove/list projects)
- Project management (open/switch context)
- Claude CLI integration (command execution, response parsing)
- Git operations validation

#### Integration Tests
- Full project lifecycle (add → open → commit → push)
- Multi-project switching
- Error handling scenarios
- Database persistence across sessions

#### End-to-End Tests
- Complete workflows using RelayTest repository
- Smart commit functionality
- Real git operations
- Claude integration in project context

## Running Tests

### Quick Test
```bash
# From server directory
make test
```

### Manual Test Steps
```bash
# Build and run tests
make build
cd test
go run *.go
```

### Cleanup After Tests
```bash
# Clean up test artifacts
./test/cleanup.sh
```

### Setup Test Environment
```bash
# Create RelayTest repository if needed
make setup-test-project
```

## Test Configuration

The test suite uses these paths:
- **RelayTest Path**: `/Users/reed/Code/Personal/RelayTest`
- **Test Database**: `~/.relay/test_relay.db`
- **Relay Binary**: `../relay`
- **Backup Database**: `~/.relay/relay.db.backup`

## Expected Test Results

### Passing Tests ✅
All tests should pass in a properly configured environment with:
- Go installed and working
- Claude CLI available and configured
- Git configured with user name/email
- RelayTest repository exists and is a valid git repo

### Expected Failures ⚠️
Some tests may fail in restricted environments:
- **Claude permissions**: Smart commit tests may fail if Claude CLI lacks git permissions
- **Git authentication**: Push operations may fail without proper git credentials
- **Network access**: Tests requiring Claude API access may fail offline

## Test Output

### Successful Run
```
🚀 Starting Relay Test Suite
============================
🔧 Setting up test environment...
✅ Test environment ready

📊 Test Suite: Database Operations
----------------------------------------
✅ PASS Database Creation (15ms)
✅ PASS Add Project (8ms)
✅ PASS Get Project (5ms)
... (more tests)

🎉 All tests passed! Relay is working correctly.
```

### Failed Run
```
❌ FAIL Smart Commit (2.3s)
   Error: Claude CLI permissions restricted

⚠️ 1 test(s) failed
❌ 23/24 tests passed. Please review the failures above.
```

## Adding New Tests

### Creating a New Test
1. Add test function to appropriate test file
2. Register test in the corresponding `run*Tests()` function
3. Follow the pattern: `testFunctionName(config *TestConfig) error`
4. Return `nil` for success, `error` for failure

### Test Best Practices
- Clean up after each test (remove created files/projects)
- Use the provided `TestConfig` for consistent paths
- Handle expected failures gracefully
- Provide descriptive error messages
- Test both success and failure cases

## Troubleshooting

### Common Issues

**"Claude CLI not available"**
- Install Claude CLI: Follow Claude Code installation instructions
- Verify with: `claude --version`

**"RelayTest directory not found"**
- Run: `make setup-test-project`
- Or manually create git repository at the expected path

**"Database permission errors"**
- Ensure `~/.relay` directory is writable
- Run cleanup script: `./test/cleanup.sh`

**"Git user not configured"**
- Set git user: `git config --global user.name "Your Name"`
- Set git email: `git config --global user.email "your@email.com"`

### Test Environment Reset
```bash
# Complete reset
./test/cleanup.sh
make clean
make setup-test-project
make test
```

## Continuous Integration

The test suite is designed to work in CI/CD environments:
- Exit code 0 for all tests passing
- Exit code 1 for any test failures
- Detailed output for debugging
- Cleanup of test artifacts
- No external dependencies beyond Go and git

Add to your CI pipeline:
```yaml
- name: Run Relay Tests
  run: |
    cd server
    make test
```
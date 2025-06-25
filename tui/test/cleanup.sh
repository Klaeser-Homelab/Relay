#!/bin/bash

# Relay Test Cleanup Script
# This script cleans up test artifacts and restores the system to a clean state

echo "🧹 Relay Test Cleanup"
echo "===================="

# Get home directory
HOME_DIR=$(eval echo ~$USER)
RELAY_DIR="$HOME_DIR/.relay"

# Function to safely remove file if it exists
safe_remove() {
    if [ -f "$1" ]; then
        rm -f "$1"
        echo "✅ Removed: $1"
    fi
}

# Function to safely remove directory if it exists
safe_remove_dir() {
    if [ -d "$1" ]; then
        rm -rf "$1"
        echo "✅ Removed directory: $1"
    fi
}

# Clean up test databases
echo "🗄️ Cleaning up test databases..."
safe_remove "$RELAY_DIR/test_relay.db"
safe_remove "$RELAY_DIR/relay.db.backup"

# Clean up test files in RelayTest repository
echo "📁 Cleaning up test files in RelayTest..."
RELAY_TEST_DIR="/Users/reed/Code/Personal/RelayTest"

if [ -d "$RELAY_TEST_DIR" ]; then
    cd "$RELAY_TEST_DIR"
    
    # Remove test files created during testing
    safe_remove "test_commit_file.txt"
    safe_remove "smart_commit_test.txt"
    safe_remove "commit_test.md"
    
    # Reset git status if there are any uncommitted test files
    if command -v git &> /dev/null; then
        echo "🔄 Resetting git status in RelayTest..."
        git status --porcelain | while read status file; do
            if [[ "$file" == *"test"* ]] || [[ "$file" == *"_test.txt" ]] || [[ "$file" == *"_test.md" ]]; then
                git clean -f "$file" 2>/dev/null || true
                echo "   Cleaned: $file"
            fi
        done
    fi
else
    echo "ℹ️  RelayTest directory not found (this is okay)"
fi

# Clean up temporary test projects
echo "🗂️ Cleaning up temporary test directories..."
# Remove any temp directories that might have been created
find /tmp -maxdepth 1 -name "relay-test-*" -type d 2>/dev/null | while read dir; do
    safe_remove_dir "$dir"
done

# Clean up build artifacts in test directory
echo "🔨 Cleaning up test build artifacts..."
TEST_DIR="$(dirname "$0")"
safe_remove "$TEST_DIR/test_runner"
safe_remove "$TEST_DIR/test_runner.exe"

# Clean up any lock files or temporary files
echo "🔒 Cleaning up lock and temporary files..."
safe_remove "$RELAY_DIR/.lock"
safe_remove "$RELAY_DIR/temp_*"

# Verify cleanup
echo ""
echo "🔍 Verification:"

if [ -d "$RELAY_DIR" ]; then
    DB_FILES=$(find "$RELAY_DIR" -name "*.db" -o -name "*.backup" 2>/dev/null | wc -l)
    if [ "$DB_FILES" -eq 0 ]; then
        echo "✅ No test database files remaining"
    else
        echo "⚠️  Some database files still exist:"
        find "$RELAY_DIR" -name "*.db" -o -name "*.backup" 2>/dev/null
    fi
else
    echo "✅ No relay directory found"
fi

# Check for any remaining test files
if [ -d "$RELAY_TEST_DIR" ]; then
    TEST_FILES=$(find "$RELAY_TEST_DIR" -name "*test*" -type f 2>/dev/null | wc -l)
    if [ "$TEST_FILES" -eq 0 ]; then
        echo "✅ No test files remaining in RelayTest"
    else
        echo "⚠️  Some test files still exist in RelayTest:"
        find "$RELAY_TEST_DIR" -name "*test*" -type f 2>/dev/null
    fi
fi

echo ""
echo "🎉 Cleanup completed!"
echo ""
echo "To run a fresh test suite:"
echo "  cd $(dirname $(dirname "$0"))"
echo "  make test"
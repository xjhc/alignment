import pytest
import os
import sys

def main():
    """Runs all end-to-end tests using pytest."""
    
    # Add the current directory to the Python path to find test modules
    sys.path.insert(0, os.path.abspath(os.path.dirname(__file__)))
    
    # Run pytest on all test files in the 'e2e' directory
    # -v for verbose output
    # --durations=5 to show the 5 slowest tests
    # --ignore to skip non-test files
    # The 'e2e/' argument tells pytest to look for tests in that directory
    result = pytest.main([
        "-v", 
        "--durations=5",
        "--ignore=e2e/test_example.py",
        "e2e/"
    ])
    
    # Exit with the same code as pytest
    sys.exit(result)

if __name__ == "__main__":
    main()
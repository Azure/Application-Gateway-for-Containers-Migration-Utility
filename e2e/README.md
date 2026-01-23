# AGIC -> AGC Migration End to End Tests

We have two types of E2E tests for the migration tooling:

1. CLI tool input/output validation tests.
2. AGC integration tests - these tests are not available on GitHub.

## Standardized Structure
Each test lives in its own sub directory, and uses standardized directory structure to make them easy to read and write:

- `input/`: YAML files describing AGIC resources to be fed into the CLI migration tool.
- `output/`: Expected output (AGC resources) of the CLI migration tool.
- `report.yaml`: Expected migration report; this report must match the generated report, and resources in it are cross checked
  with the actual output.
- `XXX_test.go`: test code.

## CLI Tool Input/Output Validation Tests
These tests execute the CLI migration tool over a set of input YAML files, and check the tool output matches the expected output.

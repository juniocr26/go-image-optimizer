# Test Data

`storage/testdata/images` contains versioned real image fixtures for backend integration tests.

These files are test inputs only. They are not application upload storage, compressed outputs, processing history, or user data. The application still processes uploads synchronously and returns compressed bytes directly to the caller without persisting them under `storage`.

Compressed test results must stay in memory or in Go test temporary paths such as `t.TempDir()`. Do not write generated outputs back into `storage/testdata/images`.

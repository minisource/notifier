package repository

import "errors"

// ErrVersionConflict is returned when an optimistic-concurrency update fails
// because the stored version changed between read and write.
var ErrVersionConflict = errors.New("version conflict: state changed concurrently")

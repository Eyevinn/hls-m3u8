### Added

- `ListType` constant `UNKNOWN`, returned by `Decode`, `DecodeFrom` and `DecodeWith` when the playlist type
  is not known (PR #99)

### Changed

- `DecodeFrom` and `DecodeWith` read an `io.Reader` line by line instead of copying all of it first,
  which uses less memory, and in strict mode they stop reading at the first error (PR #99)
- `ErrExtM3UAbsent` is now returned with the detected playlist type instead of 0 (PR #99)

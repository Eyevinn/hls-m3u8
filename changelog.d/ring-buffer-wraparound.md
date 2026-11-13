### Changed

- Encoded output of a live media playlist with more segments than `winsize` changes,
  since `EXT-X-MEDIA-SEQUENCE` now matches the first segment written (PR #91)

### Fixed

- Panic when encoding a media playlist whose segment ring buffer has wrapped around,
  e.g. after `capacity` calls to `Slide` (PR #91)
- `EXT-X-MEDIA-SEQUENCE` now reports the sequence number of the first segment in the
  encoded playlist, instead of that of the head segment (PR #91)

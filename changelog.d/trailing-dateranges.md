### Added

- `MediaPlaylist.TrailingDateRanges` provides `EXT-X-DATERANGE` tags (SCTE-35) found after the last segment
- `MediaPlaylist.AppendTrailingDateRange` to add such a tag when generating a playlist

### Changed

- Decoding no longer fails on SCTE-35 `EXT-X-DATERANGE` tags after the last segment

### Deprecated

- `ErrDanglingSCTE35DateRange` is never returned anymore

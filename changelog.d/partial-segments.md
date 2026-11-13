### Fixed

- `Encode` no longer consumes the partial segments of a low-latency media playlist,
  so that repeated calls keep writing all `EXT-X-PART` tags (PR #91)
- Expired partial segments are now also removed when a full segment is appended, so
  they no longer linger, and are no longer reported by `IsSegmentReady` (PR #91)

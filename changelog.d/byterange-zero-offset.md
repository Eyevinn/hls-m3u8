### Fixed

- `EXT-X-BYTERANGE` no longer writes `@0` for a sub-range that continues the previous
  segment's range, since an omitted offset and `@0` are not equivalent (PR #86)

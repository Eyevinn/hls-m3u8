### Changed

- `EXT-X-VERSION` in encoded output may be higher than before, since it is now raised to the
  minimum version required by the playlist content. `SetVersion` therefore sets a floor, and a
  version lower than the required minimum is no longer output (issue #95)

### Fixed

- `Encode` now raises `EXT-X-VERSION` to the minimum version required by the playlist content,
  so e.g. a media playlist with an `EXT-X-MAP` tag but no `EXT-X-I-FRAMES-ONLY` tag signals
  version 6 instead of 5 (issue #95)

### Changed

- `EXT-X-BYTERANGE` omits its offset only for a sub-range that starts exactly where the
  previous sub-range of the same resource ended. A segment whose `Offset` is zero but does
  not continue the previous range is now written as `@0` (PR #93)
- `EXT-X-PART` `BYTERANGE` is always written with an explicit offset, which is valid and
  unambiguous, so an omitted part offset is not reproduced verbatim on encode (PR #93)
- `EXT-X-MAP` with a `BYTERANGE` that lacks the required offset now reports that, rather
  than a generic parse error (PR #93)

### Fixed

- `MediaSegment.Offset` and `PartialSegment.Offset` are now resolved to absolute offsets
  when the playlist omits them, instead of being left as zero (PR #93)
- Decoding no longer fails on an `EXT-X-PART` whose `BYTERANGE` omits the optional offset (PR #93)
- An `EXT-X-BYTERANGE` or `EXT-X-PART` `BYTERANGE` that omits its offset without a preceding
  sub-range of the same resource is now reported in strict mode, instead of silently
  decoding to offset zero (PR #93)

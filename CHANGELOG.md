# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- Nothing yet

## [v0.6.1] 2025-10-28

### Added

- Reduced memory allocations using sync.Pool for playlists
- More documentation of examples

### Fixed

- Wrap-around bug in segment order returned from GetAllSegments (Issue #65)

## [v0.6.0] 2025-06-18
### ⚠️ Breaking changes ⚠️
- `Alternative.Channels` changed from `string` to `*Channels`.
### Added
- Support for `INSTREAM-ID` for non-CC tracks.
- Support for new `CHANNELS` parameters (spatial audio identifiers and channel usage indicators).

### Chore
- Verify support for preload date ranges.

## [v0.5.4] 2025-06-03
### Added
Support for setting skippedsegments on a MediaPlaylist

## [v0.5.3] 2025-06-02

### Fixed
- Calculate skips based on amount of segments rather than target duration (#46)

## [v0.5.2] 2025-05-08

### Fixed

- SCTE decoding state is now decoupled from daterange decoding, fixing an issue where segments with CUE and DATERANGE tags lost the data from one whike retaining data from the other.

## [v0.5.1] 2025-04-24

### Fixed

- EXT-X-DISCONTINUITY is now written before other segment-level tags to ensure that no information is lost to players

## [v0.5.0] 2025-04-23

### ⚠️ Breaking changes ⚠️

- Now requires go 1.21 or newer.
- MediaPlaylist.Key changed to MediaPlaylist.Keys, is now []Key instead of *Key.
- MediaSegment.Key changed to MediaPlaylist.Keys, is now []Key instead of *Key.

### Added

- Support for multiple Keys on media playlist level
- Support for multiple EXT-X-KEY tags per segment

### Chore

- Bumped minimum go version to 1.21


## [v0.4.0] 2025-03-11

### Added

- EncodeWithSkip() to media playlist
- EXT-X-PART support
- EXT-X-PART-INF support
- EXT-X-SERVER-CONTROL support
- EXT-X-PRELOAD-HINT support
- EXT-X-SKIP support
- EXT-X-GAP support
- EXT-X-SESSION-KEY support
- EXT-X-CONTENT-STEERING support
- EXT-X-START to master playlist
- SetWritePrecision() and WritePrecision() methods to set the number of decimal places for floating point numbers

### Fixed

- EXT-X-START was not written for negative values
- EXT-X-MAP is written when changed

### Chore

- Refactored EXT-X-DEFINE and EXT-X-MAP parsing and writing

## [v0.3.0] 2025-01-14

### Added

- Support for multiple EXT-X-DATERANGE tags in a media playlist
- SCTE-35 EXT-X-DATERANGE tags are attached to current segment
- MediaPlaylist.SCTE35Syntax() method
- SCTE35Syntax has String() method
- EXT-X-DEFINE support in both master and media playlists
- EXT-X-SESSION-DATA support

### Changed

- EXT-X-DATERANGE for SCTE-35 are stored as slice in Segment
- SCTE35Syntax type has a new default SCTE35_NONE

## [v0.2.0] 2025-01-07

### Changed

- FORCED and AUTOSELECT types changed from string to bool
- Removed SUBTITLES from EXT-X-MEDIA since not in [rfc8216bis-16][rfc8216-bis]
- Changed tests to use matryer.is for conciseness
- Improved documentation
- TargetDuration is now an uint
- PROGRAM-ID parameter is obsolete from version 6. Changed from uin32 to *int.
- Only remove quotes on Quoted-String parameters, and not in general.

### Added

- Complete list of EXT-X-MEDIA attributes: ASSOC-LANGUAGE, STABLE-RENDITION-ID, INSTREAM-ID, BIT-DEPTH, SAMPLE-RATE
- GetAllAlternatives() method to MasterPlaylist
- Improved playlist type detection
- Support for SCTE-35 signaling using EXT-X-DATERANGE (following [rfc8216-bis][rfc8216-bis])
- Support for full EXT-X-DATERANGE parsing and writing
- TARGETDURATION calculation depends on HLS version
- New function CalculateTargetDuration
- New method MediaPlaylist.SetTargetDuration that sets and locks the value
- Full parsing and rendering of EXT-X-STREAM-INF parameters
- FUll parsing and writing of EXT-X-I-FRAME-STREAM-INF parameters
- EXT-X-ALLOW-CACHE support in MediaPlaylist (obsolete since version 7)

### Fixed

- Renditions were not properly distributed to Variants
- EXT-X-KEY was written twice before first segment
- FORCED attribute had quotes

### Removed

- Removed HLSv2 support (integer EXTINF durations)

## v0.1.0 - cleaned grafov/m3u8 code

### Changed

The following changes are wrt to initial copy of [grafov/m3u8][grafov] files:

- code changes to pass linting including Example names
- made errors more consistent and more verbose
- removed all Widevine-specific HLS extensions (obsolete)

### Added

- initial version of the repo

[Unreleased]: https://github.com/Eyevinn/hls-m3u8/compare/v0.6.1...HEAD
[v0.6.1]: https://github.com/Eyevinn/hls-m3u8/compare/v0.6.0...v0.6.1
[v0.6.0]: https://github.com/Eyevinn/hls-m3u8/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/Eyevinn/hls-m3u8/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/Eyevinn/hls-m3u8/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/Eyevinn/hls-m3u8/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/Eyevinn/hls-m3u8/compare/v0.1.0...v0.2.0
[grafov]: https://github.com/grafov/m3u8
[rfc8216bis-16]: https://datatracker.ietf.org/doc/html/draft-pantos-hls-rfc8216bis-16
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Support for multiple EXT-X-DATERANGE tags in a media playlist
- SCTE-35 EXT-X-DATERANGE tags are attached to current segment
- MediaPlaylist.SCTE35Syntax() method
- SCTE35Syntax has String() method
- EXT-X-DEFINE support in both master and media playlists

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

[Unreleased]: https://github.com/Eyevinn/mp4ff/compare/v0.2.0...HEAD
[v0.2.0]: https://github.com/Eyevinn/mp4ff/compare/v0.1.0...v0.2.0
[grafov]: https://github.com/grafov/m3u8
[rfc8216bis-16]: https://datatracker.ietf.org/doc/html/draft-pantos-hls-rfc8216bis-16
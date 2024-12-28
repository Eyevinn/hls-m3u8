# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- CHANNELS attribute to EXT-X-MEDIA
- GetAllAlternatives() method to MasterPlaylist
- Improved playlist type detection

### Fixed

- Renditions were not properly distributed to Variants
- EXT-X-KEY was written twice before first segment

## v0.1.0 - cleaned grafov/m3u8 code

### Changed

The following changes are wrt to initial copy of [grafov/m3u8][grafov] files:

- code changes to pass linting including Example names
- made errors more consistent and more verbose
- removed all Widevine-specific HLS extensions (obsolete)

### Added

- initial version of the repo

[Unreleased]: https://github.com/Eyevinn/mp4ff/compare/v0.1.0...HEAD
[grafov]: https://github.com/grafov/m3u8
# Changelog entries

Each change that users will notice adds a file to this directory instead of editing
[CHANGELOG.md](../CHANGELOG.md). Separate files never conflict, so pull requests can be merged in any
order. When a version is released, the entries move into CHANGELOG.md and the files are removed.

## Writing an entry

Name the file after the change, for example `byterange-zero-offset.md`, and use the sections of
[Keep a Changelog](https://keepachangelog.com/en/1.0.0/):

```markdown
### Fixed

- `EXT-X-BYTERANGE` no longer writes `@0` for a sub-range that continues the previous
  segment's range (PR #86)
```

- The sections are `### Added`, `### Changed`, `### Deprecated`, `### Removed`, `### Fixed` and
  `### Security`. A file can have several.
- Keep each entry to one or two lines that say what changed for users and name the affected API or
  tag. Indent continuation lines by two spaces. End the entry with the issue or pull request number,
  such as `(issue #95)` or `(PR #91)`, when there is one. End the entry of a breaking change with
  "This is a breaking change".
- A pull request with several changes can add a file per change, or one file with all its entries.
- Changes that users don't notice, such as tests, CI and refactoring, need no entry.

[`scripts/check-changelog.sh`](../scripts/check-changelog.sh) checks the files. It runs in CI, and as
a pre-commit hook when CHANGELOG.md or this directory changes.

## At a release

The pull request or commit that prepares release vX.Y.Z

1. adds `## [vX.Y.Z] YYYY-MM-DD` below `## [Unreleased]` in CHANGELOG.md, with the entries of all
   files here under one heading per section, in the order Added, Changed, Deprecated, Removed, Fixed,
   Security. Entries can be edited on the way, for example to merge related ones;
2. removes the entry files, so that only this README remains;
3. updates the compare links at the end of CHANGELOG.md.

While the latest version in CHANGELOG.md has no tag, the check fails if this directory holds entry
files. Tag `vX.Y.Z` only on a commit where `scripts/check-changelog.sh vX.Y.Z` passes.

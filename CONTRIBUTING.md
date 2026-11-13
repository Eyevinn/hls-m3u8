# Contributing to hls-m3u8

Contributions are welcome, as issues and as pull requests against the `main` branch.
Coding agents follow these conventions too, with the additions in [AGENTS.md](AGENTS.md).

## Submitting Issues

We use GitHub issues to track public bugs. If you are submitting a bug, please provide
as much info as possible to make it easier to reproduce the issue and update unit tests.

## Setting up

The module needs Go 1.21 or later. Lint, test and check the licenses of the dependencies with

    make lint
    make test
    make check-licenses

`make` runs them all and writes a coverage report. `make lint` needs
[golangci-lint](https://golangci-lint.run) and `make check-licenses` needs
[wwhrd](https://github.com/frapposelli/wwhrd). Run a single test with `go test -run TestName ./m3u8`.

Pre-commit hooks run `go fmt`, `golangci-lint`, the unit tests and the changelog check. Install
[pre-commit](https://pre-commit.com) and run `pre-commit install` once per clone.

## Contributing Code

We follow the [GitHub Flow](https://guides.github.com/introduction/flow/index.html) so all contributions happen through pull requests. We actively welcome your pull requests:

1. Fork the repo and create your branch from main.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. For a change that users will notice, add a changelog entry file to [changelog.d](changelog.d/)
   instead of editing [CHANGELOG.md](CHANGELOG.md), as described in
   [changelog.d/README.md](changelog.d/README.md).
6. Add yourself to the AUTHORS list, if not already there
7. Issue that pull request!

When submitting code changes your submissions are understood to be under the same MIT License that covers the project. Feel free to contact Eyevinn Technology if that's a concern.

### Working with coding agents

Changes written with the help of a coding agent are welcome. Credit the agent as a co-author with a
`Co-Authored-By:` trailer in each commit it helped write, naming the agent and its model, for example
`Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. The person who opens the pull request is
its author, reviews what the agent wrote and answers questions about it. AUTHORS lists people, not
agents.

## Commit messages

Commit messages follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/#summary)
specification. Examples:

- `feat: support a CUE-IN and a CUE-OUT on the same segment`
- `fix: resolve omitted byte range offsets to absolute offsets when decoding`
- `perf: decode from io.Reader line by line instead of buffering the whole input`
- `ci: use latest checkout and setup-node majors (v7)`

The subject says what the commit does. The body says why: what was broken, how it was found and what
the change affects.

## Code conventions

### Specification

The library follows the latest [rfc8216bis](https://datatracker.ietf.org/doc/draft-pantos-hls-rfc8216bis/)
draft that [README.md](README.md) names. Code that implements a rule of the specification says in a
comment which section it follows.

### Decoding and encoding

Every tag that is decoded is also encoded. A playlist tag is decoded in `decodeLineOfMasterPlaylist`
or `decodeLineOfMediaPlaylist` in [m3u8/reader.go](m3u8/reader.go) and encoded in
[m3u8/writer.go](m3u8/writer.go). For a new tag, add a sample playlist to
[m3u8/sample-playlists](m3u8/sample-playlists/) and its name to `TestReadWritePlaylists` in
[m3u8/read_write_test.go](m3u8/read_write_test.go), which checks that decoding and encoding give the
file back.

Decoding takes a `strict` flag. In strict mode, a playlist that breaks a rule of the specification gives
an error. Otherwise, decoding accepts what it can still make sense of.

A tag or attribute that needs a minimum protocol version is added to `CalcMinVersion` in
[m3u8/calcversion.go](m3u8/calcversion.go), since `Encode` raises `EXT-X-VERSION` to that version.

### The encoding cache

`Encode` keeps its output in a buffer that serves as a cache, and must not change the playlist, so
that it can be called repeatedly. Every method that changes a playlist calls `ResetCache`, so that the
next `Encode` writes the change.

### Tests

Tests use the [is](https://github.com/matryer/is) package: `is := is.New(t)`, then `is.NoErr(err)`,
`is.Equal(got, want)` and so on. Sample playlists for tests are in
[m3u8/sample-playlists](m3u8/sample-playlists/).

### Comments and changelog

Code comments describe the current behaviour. What was broken before and why it changed belongs
in the commit message and the pull request description.

For each change that users will notice, add a file to [changelog.d](changelog.d/) instead of editing
[CHANGELOG.md](CHANGELOG.md), as described in [changelog.d/README.md](changelog.d/README.md). Keep
each entry to one or two lines that say what changed and name the affected API or tag. The entries
move into CHANGELOG.md when a version is released.

### Documentation

The guide to the library is the package documentation in [m3u8/doc.go](m3u8/doc.go), shown on
[pkg.go.dev](https://pkg.go.dev/github.com/Eyevinn/hls-m3u8/m3u8), and the structure and design
section of [README.md](README.md). Update them when you change how the library is used.

## Releases

Maintainers prepare release vX.Y.Z by moving the changelog entries into CHANGELOG.md, as described in
[changelog.d/README.md](changelog.d/README.md), and tag that commit `vX.Y.Z`. A tag is permanent for
Go modules, since the module proxy caches it, so never move or reuse one.

## Code of Conduct

## Our Pledge

In the interest of fostering an open and welcoming environment, we as contributors and maintainers pledge to making participation in our project and our community a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

## Our Standards

Examples of behavior that contributes to creating a positive environment include:

- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

Examples of unacceptable behavior by participants include:

- The use of sexualized language or imagery and unwelcome sexual attention or advances
- Trolling, insulting/derogatory comments, and personal or political attacks
- Public or private harassment
- Publishing others' private information, such as a physical or electronic address, without explicit permission
- Other conduct which could reasonably be considered inappropriate in a professional setting

## Our Responsibilities

Project maintainers are responsible for clarifying the standards of acceptable behavior and are expected to take appropriate and fair corrective action in response to any instances of unacceptable behavior.

Project maintainers have the right and responsibility to remove, edit, or reject comments, commits, code, wiki edits, issues, and other contributions that are not aligned to this Code of Conduct, or to ban temporarily or permanently any contributor for other behaviors that they deem inappropriate, threatening, offensive, or harmful.

## Scope

This Code of Conduct applies both within project spaces and in public spaces when an individual is representing the project or its community. Examples of representing a project or community include using an official project e-mail address, posting via an official social media account, or acting as an appointed representative at an online or offline event. Representation of a project may be further defined and clarified by project maintainers.

## Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be reported by contacting the project team. All complaints will be reviewed and investigated and will result in a response that is deemed necessary and appropriate to the circumstances. The project team is obligated to maintain confidentiality with regard to the reporter of an incident. Further details of specific enforcement policies may be posted separately.

Project maintainers who do not follow or enforce the Code of Conduct in good faith may face temporary or permanent repercussions as determined by other members of the project's leadership.

## Attribution

This Code of Conduct is adapted from the Contributor Covenant, version 1.4, available at http://contributor-covenant.org/version/1/4

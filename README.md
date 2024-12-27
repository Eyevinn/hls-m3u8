# hls-m3u8 - HLS m3u8 playlist library

![Test](https://github.com/Eyevinn/hls-m3u8/workflows/Go/badge.svg)
![golangci-lint](https://github.com/Eyevinn/mp4ff/workflows/golangci-lint/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Eyevinn/hls-m3u8)](https://goreportcard.com/report/github.com/Eyevinn/hls-m3u8)
[![license](https://img.shields.io/github/license/Eyevinn/hls-m3u8.svg)](https://github.com/Eyevinn/hls-m3u8/blob/master/LICENSE)

hls-m3u8 implements parsing and generation of HLS m3u8 playlists.
HLS (HTTP Live Streaming) is an evolving protocol with multiple versions.
Versions 1-7 are described in [IETF RFC8216][rfc8216], but the protocol has continued
to evolve with new features and versions in a
series of [Internet Drafts rfc8216bis][rfc8216bis].
The current version (Dec. 27 2024) is [rfc8216bis-16][rfc8216bis].

One of the major libraries in Go for parsing and generating HLS playlists,
aka m3u8 files, has been the Github project [grafov/m3u8][grafov].
However, the majority of that code was written up to version 5, and
the library has hardly been updated in a long time. It was finally archived in Dec. 2024.

The goal of this library, hls-m3u8,  is to provide an up-to-date replacement and improvement
of  the [m3u8][grafov] library. The aim is to follow the HLS specification
as it evolves and add all new elements and do other updates in order that
all m3u8 documents (from version 3 and forward) can be parsed and generated.

## Installation / Usage

This is a library that should be downloaded like other Go code.

To enable it in your Go project, run

```sh
go get github.com/Eyevinn/hls-m3u8/m3u8
```

## Development

There are tests and sample-playlists available.
There is also a `Makefile` that runs test, checks coverage, and the license
of dependencies.

### pre-commit checks

To run checks before any commit is accepted, install [pre-commit][pre-commit] and then run

```sh
> pre-commit install --hook-type commit-msg
```

to set up the automatic tests.

## Compatibility with grafov/m3u8

This project tries to align to the archived library [grafov/m3u8][grafov] to make the transition from that library relatively simple.

As a first release (v0.1.0), the code should be a cleaned version of [grafov/m3u8][grafov],
but later versions will probably have some non-compatible changes.

Some notable changes in the HLS specification are:

* the introduction of Low-Latency HLS and
partial segments in [rfc8216bis-07][rfc8216bis-07]
* change of name of the top-level multi-variant playlist from `Master Playlist` to `Multivariant Playlist`[rfc8216bis-10][rfc8216bis-10]

There area also plenty of new tags and use cases.

The aim is to provide upgrade instructions, when the library API changes.

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md)

## License

This project is licensed under the BSD Clause 3 License, see [LICENSE](LICENSE)
to align with the included code from the [grafov/m3u8][grafov] project.

## Issues and Discussion

Use the [issues][issues] to file an issue. Questions and discussions belong to the
[discussions][discussions] part of the repo. For support questions, see below.

## Support

Join our [community on Slack](https://slack.streamingtech.se) where you can post any questions regarding any of Eyevinn's open source projects. Eyevinn's consulting business can also offer you:

- Further development of this component
- Customization and integration of this component into your platform
- Support and maintenance agreement

Contact [sales@eyevinn.se](mailto:sales@eyevinn.se) if you are interested.

# About Eyevinn Technology

[Eyevinn Technology](https://www.eyevinntechnology.se) is an independent consultant firm specialized in video and streaming. Independent in a way that we are not commercially tied to any platform or technology vendor. As our way to innovate and push the industry forward we develop proof-of-concepts and tools. The things we learn and the code we write we share with the industry in [blogs](https://dev.to/video) and by open sourcing the code we have written.

Want to know more about Eyevinn and how it is to work here. Contact us at work@eyevinn.se!

[rfc8216]: https://datatracker.ietf.org/doc/html/rfc8216
[rfc8216bis]: https://datatracker.ietf.org/doc/draft-pantos-hls-rfc8216bis/
[rfc8216bis-07]: https://datatracker.ietf.org/doc/html/draft-pantos-hls-rfc8216bis-07
[rfc8216bis-10]: https://datatracker.ietf.org/doc/html/draft-pantos-hls-rfc8216bis-10
[rfc8216bis-16]: https://datatracker.ietf.org/doc/html/draft-pantos-hls-rfc8216bis-16
[grafov]: https://github.com/grafov/m3u8
[issues]: https://github.com/Eyevinn/hls-m3u8/issues
[discussions]: https://github.com/Eyevinn/hls-m3u8/discussions
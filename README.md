# hls-m3u8 - HLS m3u8 playlist library

![Test](https://github.com/Eyevinn/hls-m3u8/workflows/Go/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/Eyevinn/hls-m3u8/badge.svg?branch=main)](https://coveralls.io/github/Eyevinn/hls-m3u8?branch=main)
[![GoDoc](https://godoc.org/github.com/Eyevinn/hls-m3u8?status.svg)](http://godoc.org/github.com/Eyevinn/hls-m3u8)
[![Go Report Card](https://goreportcard.com/badge/github.com/Eyevinn/hls-m3u8)](https://goreportcard.com/report/github.com/Eyevinn/hls-m3u8)
[![license](https://img.shields.io/github/license/Eyevinn/hls-m3u8.svg)](https://github.com/Eyevinn/hls-m3u8/blob/main/LICENSE)

hls-m3u8 implements parsing and generation of HLS m3u8 playlists.
HLS (HTTP Live Streaming) is an evolving protocol with multiple versions.
Versions 1-7 are described in [IETF RFC8216][rfc8216], but the protocol has continued
to evolve with new features and versions in a
series of [Internet Drafts rfc8216bis][rfc8216bis].

One of the major libraries in Go for parsing and generating HLS playlists,
aka m3u8 files, has been the Github project [grafov/m3u8][grafov].
However, the majority of that code was written up to version 5,
It was finally archived in Dec. 2024.

The goal of this library, `hls-m3u8`,  is to provide an up-to-date replacement and improvement of the [grafov/m3u8][grafov] library. The aim is to follow the HLS specification
as it evolves and add all new elements and do other updates in order that
all m3u8 documents (from version 3 and forward) can be parsed and generated.
There is typically a new draft every 6 months.
The current version (Sep 11 2025) is [rfc8216bis-18][rfc8216bis].
Its specification should be supported by this repo, but all parameter
values are not validated.

The HLS protocol has different versions, and there are rules for what minimal
version to signal depending on features being used. That mechanism is implemented
in the `CalcMinVersion()` method of the `Playlist` interface.

## Structure and design of the code

There are two types of m3u8 playlists: `Master` or `Multivariant` playlists, and `Media` playlists.
These are represented as two different structs, but they have a common interface `Playlist`.

There is a function `Decode`, that decodes and autodetects the type of playlist, by decoding
both in parallel, and stopping one, once the type is known.

For generating playlists, one starts by calling either `NewMasterPlaylist` or `NewMediaPlaylist`.
One can then `Set` or `Append` extra data. For example, one can `Append` a full media segment or
`AppendPartial` a partial segment to Low-Latency HLS media playlists.

For live media playlists with a fixed sliding window, one
can set a `winsize` and it will be used to always output
the latest segments.

For VOD or EVENT media playlists, the `winsize` should be 0.

For writing, there are `Encode` methods that return a `*bytes.Buffer`. This buffer serves as a cache.
It is also possible to call `EncodeWithSkip` to signal skipping of the first `n` segments.
The `String` method makes it easy to use the standard `fmt.Print` functions.

## Installation / Usage

This is a library that should be downloaded like other Go code.

To enable it in your Go project, run

```sh
go get github.com/Eyevinn/hls-m3u8/m3u8
```

To use the code add

```go
import github.com/Eyevinn/hls-m3u8/m3u8
```

to your source files.

### Examples

Testable examples can be found in m3u8/*_test.go files. For nicer browser formatting, see
[pkg.go.dev][m3u8-examples].

Sample playlists are provided in `m3u8/sample-playlists`.

In addition, there are some simple examples below (without error handling)

#### Create simple media playlist

This uses a sliding window of 3 segments and
maximum of 50 segments.

	p, _ := NewMediaPlaylist(3, 50)
	for i := 0; i < 5; i++ {
	  _ = p.Append(fmt.Sprintf("test%d.ts", i), 5.0, "")
	}
	fmt.Println(p)

We add 5 testX.ts segments to the playlist before encoding it to M3U8 format and converting to string.
Due to the sliding window, only the last three
segments will be included.

#### Parsing of a master playlist:

	f, _ := os.Open("sample-playlists/master.m3u8")
	p := NewMasterPlaylist()
	_ = p.DecodeFrom(bufio.NewReader(f), true /*strict*/)
	fmt.Printf("Playlist object: %+v\n", p)

Here the `strict` parameter is set to true. If false, the parser
allows more errors.

## Development

There are tests and sample-playlists available.
There is also a `Makefile` that runs test, checks coverage, and the license
of dependencies.

For tests, the [is][is] package is used. It outputs failing tests with
line numbers and colors, but colors do not work properly in VisualStudio Code.
To turn them off in VSC, add the following configuration:

```json
    "go.testEnvVars": {
        "NO_COLOR" : "YES"
    },
```

### pre-commit checks

To run checks before any commit is accepted, install [pre-commit][pre-commit] and then run

```sh
> pre-commit install
```

to set up the automatic tests.

## Compatibility with grafov/m3u8

This project tries to align to the archived library [grafov/m3u8][grafov] to make the transition from that library relatively simple.

The first release (v0.1.0), is essentially a cleaned
and slightly bug-fixed  version of [grafov/m3u8][grafov]

Replace `import github.com/grafov/m3u8` with
`import github.com/Eyevinn/hls-m3u8/m3u8` and you should
hopefully be fine to go.

Later versions have more changes and additions, so it is likely
that more changes are needed to legacy code.

See [CHANGELOG](CHANGELOG) for a list of changes.

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
[is]: https://github.com/matryer/is
[m3u8-examples]: https://pkg.go.dev/github.com/Eyevinn/hls-m3u8/m3u8#pkg-examples
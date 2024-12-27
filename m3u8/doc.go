package m3u8

/* Package m3u8 is parser & generator library for Apple HLS.

M3U8 is line-based text format and parsing library for it must be simple
too. It did not offer ways to play HLS or handle playlists over
HTTP. Library features are:

  - Support HLS specs up to version 5 of the protocol.
  - Parsing and generation of master-playlists and media-playlists.
  - Autodetect input streams as master or media playlists.
  - Offer structures for keeping playlists metadata.
  - Support for sliding window for live media playlists.
  - Encryption keys support for usage with DRM systems.

Library coded accordingly with IETF draft
http://tools.ietf.org/html/draft-pantos-http-live-streaming

Examples of usage may be found in *_test.go files of a package. Also
see below some simple examples.

Create simple media playlist with sliding window of 3 segments and
maximum of 50 segments.

	p, e := NewMediaPlaylist(3, 50)
	if e != nil {
	  panic(fmt.Sprintf("Create media playlist failed: %s", e))
	}
	for i := 0; i < 5; i++ {
	  e = p.Add(fmt.Sprintf("test%d.ts", i), 5.0)
	  if e != nil {
		panic(fmt.Sprintf("Add segment #%d to a media playlist failed: %s", i, e))
	  }
	}
	fmt.Println(p.Encode(true).String())

We add 5 testX.ts segments to playlist then encode it to M3U8 format
and convert to string.

Next example shows parsing of master playlist:

	f, err := os.Open("sample-playlists/master.m3u8")
	if err != nil {
	  fmt.Println(err)
	}
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	if err != nil {
	  fmt.Println(err)
	}

	fmt.Printf("Playlist object: %+v\n", p)

We are open playlist from the file and parse it as master playlist.
*/

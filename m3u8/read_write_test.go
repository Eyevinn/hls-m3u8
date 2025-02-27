package m3u8

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestReadWriteAlternative(t *testing.T) {
	is := is.New(t)
	cases := []struct {
		desc   string
		line   string
		strict bool
		error  string
	}{
		{
			desc:   "empty line",
			line:   "",
			strict: false,
			error:  `invalid line: ""`,
		},
		{
			desc: "audio line",
			line: `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="English",LANGUAGE="en",DEFAULT=YES,AUTOSELECT=YES,` +
				`CHANNELS="6",URI="english.m3u8"`,
			strict: true,
			error:  "",
		},
		{
			desc: "closed captions",
			line: `#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="cc",NAME="English",LANGUAGE="en",DEFAULT=NO,INSTREAM-ID="CC1"`,
		},
		{
			desc: "all attributes",
			line: `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="English",LANGUAGE="en",ASSOC-LANGUAGE="dk",` +
				`STABLE-RENDITION-ID="a+0",DEFAULT=YES,AUTOSELECT=YES,FORCED=YES,INSTREAM-ID="CC1",BIT-DEPTH=16,` +
				`SAMPLE-RATE=48000,CHARACTERISTICS="public.accessibility.describes-video",CHANNELS="6/-/BINAURAL",` +
				`URI="english.m3u8"`,
		},
		{
			desc: "bad DEFAULT strict",
			line: `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="English",LANGUAGE="en",` +
				`DEFAULT=YE,URI="english.m3u8"`,
			strict: true,
			error:  `DEFAULT:YE value must be YES or NO`,
		},
		{
			desc: "bad AUTOSELECT strict",
			line: `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="English",LANGUAGE="en",` +
				`AUTOSELECT=yes,URI="english.m3u8"`,
			strict: true,
			error:  `AUTOSELECT:yes value must be YES or NO`,
		},
		{
			desc: "bad FORCED strict",
			line: `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="English",LANGUAGE="en",` +
				`FORCED=no,URI="english.m3u8"`,
			strict: true,
			error:  `FORCED:no value must be YES or NO`,
		},
		{
			desc: "bad BIT-DEPTH",
			line: `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="English",LANGUAGE="en",` +
				`BIT-DEPTH=ten,URI="english.m3u8"`,
			strict: true,
			error:  `invalid BIT-DEPTH: strconv.Atoi: parsing "ten": invalid syntax`,
		},
		{
			desc: "bad SAMPLE-RATE",
			line: `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="English",LANGUAGE="en",` +
				`SAMPLE-RATE="hi",URI="english.m3u8"`,
			strict: true,
			error:  `invalid SAMPLE-RATE: strconv.Atoi: parsing "hi": invalid syntax`,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			alt, err := parseExtXMedia(c.line, c.strict)
			if c.error != "" {
				is.Equal(err != nil, true)     // must return an error
				is.Equal(c.error, err.Error()) // error message must match
				return
			}
			is.NoErr(err)
			out := bytes.Buffer{}
			writeExtXMedia(&out, &alt)
			is.Equal(c.line, trimLineEnd(out.String())) // EXT-X-MEDIA line must match
		})
	}
}

func TestReadWriteDateRange(t *testing.T) {
	is := is.New(t)
	cases := []struct {
		desc        string
		line        string
		strict      bool
		expectedErr bool
	}{
		{
			desc: "Minimal",
			line: `#EXT-X-DATERANGE:ID="min",START-DATE="2014-03-05T11:15:00Z"`,
		},
		{
			desc: "Minimal with duration",
			line: `#EXT-X-DATERANGE:ID="min",START-DATE="2014-03-05T11:15:00Z"` +
				`,DURATION=60.500`,
		},
		{
			desc: "All fields",
			line: `#EXT-X-DATERANGE:ID="min",CLASS="a",START-DATE="2014-03-05T11:15:00Z"` +
				`,CUE=PRE,END-DATE="2014-03-05T11:16:00Z",DURATION=60.000` +
				`,PLANNED-DURATION=60.000,SCTE35-CMD=0xFC00,SCTE35-OUT=0xFC0A` +
				`,SCTE35-IN=0xFC00,END-ON-NEXT=YES`,
		},
		{
			desc: "Interstitials from rfc8216bis-16",
			line: `#EXT-X-DATERANGE:ID="ad1",CLASS="com.apple.hls.interstitial",` +
				`START-DATE="2020-01-02T21:55:44.12Z",DURATION=15.000,` +
				`X-ASSET-URI="http://example.com/ad1.m3u8",X-RESUME-OFFSET=0,` +
				`X-RESTRICT="SKIP,JUMP",X-COM-EXAMPLE-BEACON=123`,
		},
		{
			desc:        "Bad start date",
			line:        `#EXT-X-DATERANGE:ID="min",START-DATE="2014/03/05T11:15:00Z"`,
			expectedErr: true,
		},
		{
			desc: "Bad end date",
			line: `#EXT-X-DATERANGE:ID="min",START-DATE="2014-03-05T11:15:00Z"` +
				`,END-DATE="2014/03/05T11:15:00Z"`,
			expectedErr: true,
		},
		{
			desc: "Bad duration",
			line: `#EXT-X-DATERANGE:ID="min",START-DATE="2014-03-05T11:15:00Z"` +
				`,DURATION=60.500.500`,
			expectedErr: true,
		},
		{
			desc: "Bad planned duration",
			line: `#EXT-X-DATERANGE:ID="min",START-DATE="2014-03-05T11:15:00Z"` +
				`,PLANNED-DURATION=60.500.500`,
			expectedErr: true,
		},
		{
			desc: "Bad SCTE35-CMD",
			line: `#EXT-X-DATERANGE:ID="min",START-DATE="2014-03-05T11:15:00Z"` +
				`,SCTE35-CMD=FC00`,
			expectedErr: true,
		},
		{
			desc: "Bad SCTE35-OUT",
			line: `#EXT-X-DATERANGE:ID="min",START-DATE="2014-03-05T11:15:00Z"` +
				`,SCTE35-OUT="0xFC00"`,
			expectedErr: true,
		},
		{
			desc: "Bad SCTE35-IN",
			line: `#EXT-X-DATERANGE:ID="min",START-DATE="2014-03-05T11:15:00Z"` +
				`,SCTE35-IN="0xFC"`,
			expectedErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			dr, err := parseDateRange(c.line)
			if c.expectedErr {
				is.Equal(err != nil, true) // must return an error
				return
			}
			is.NoErr(err) // parseDateRange did not succeed
			out := bytes.Buffer{}
			writeDateRange(&out, dr)
			is.Equal(c.line, trimLineEnd(out.String())) // EXT-X-DATERANGE line must match
		})
	}
}

func TestReadWriteInterstital(t *testing.T) {
	is := is.New(t)
	asset := "sample-playlists/media-playlist-with-interstitial.m3u8"
	f, err := os.Open(asset)
	is.NoErr(err) // open file should succeed
	p, _, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err) // decode playlist should succeed
	mp := p.(*MediaPlaylist)
	f.Close()
	out := trimLineEnd(mp.String())
	inData, err := os.ReadFile(asset)
	is.NoErr(err) // read file should succeed
	inStr := string(inData)
	inStr = trimLineEnd(strings.Replace(inStr, "\r\n", "\n", -1))
	is.Equal(inStr, out) // output must match input
}

func TestReadWriteSCTE35DateRange(t *testing.T) {
	is := is.New(t)
	asset := "sample-playlists/media-playlist-with-scte35-daterange.m3u8"
	f, err := os.Open(asset)
	is.NoErr(err) // open file should succeed
	p, _, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err) // decode playlist should succeed
	mp := p.(*MediaPlaylist)
	f.Close()
	out := trimLineEnd(mp.String())
	inData, err := os.ReadFile(asset)
	is.NoErr(err) // read file should succeed
	inStr := string(inData)
	inStr = trimLineEnd(strings.Replace(inStr, "\r\n", "\n", -1))
	is.Equal(inStr, out) // output must match input
}

func TestReadWriteExtXStreamInf(t *testing.T) {
	is := is.New(t)
	cases := []struct {
		desc   string
		line   string
		strict bool
		error  bool
	}{
		{
			desc:  "minimal",
			line:  `#EXT-X-STREAM-INF:BANDWIDTH=128000`,
			error: false,
		},
		{
			desc: "max",
			line: `#EXT-X-STREAM-INF:BANDWIDTH=128000,AVERAGE-BANDWIDTH=128000,SCORE=5.000,CODECS="avc1.4d400d,mp4a.40.2",` +
				`SUPPLEMENTAL-CODECS="mp4a.40.5",RESOLUTION=320x240,FRAME-RATE=25.000,HDCP-LEVEL=NONE,` +
				`ALLOWED-CPC="com.example.drm1:SMART-TV/PC",VIDEO-RANGE=SDR,REQ-VIDEO-LAYOUT="CH-MONO",` +
				`STABLE-VARIANT-ID="a_0",AUDIO="audio",VIDEO="video",SUBTITLES="subs",CLOSED-CAPTIONS="cc",PATHWAY-ID="X",` +
				`PROGRAM-ID=1,NAME="prop"`,
			error: false,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			vnt, err := parseExtXStreamInf(c.line, c.strict)
			if c.error {
				is.Equal(err != nil, true) // must return an error
				return
			}
			is.NoErr(err)
			out := bytes.Buffer{}
			writeExtXStreamInf(&out, vnt)
			outStr := trimLineEnd(out.String())
			is.Equal(c.line, outStr) // EXT-X-STREAM-INF line must match
		})
	}
}

func TestReadWriteExtXIFrameStreamInf(t *testing.T) {
	is := is.New(t)
	cases := []struct {
		desc   string
		line   string
		strict bool
		error  bool
	}{
		{
			desc:  "minimal",
			line:  `#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=128000,URI="iframe.m3u8"`,
			error: false,
		},
		{
			desc: "max",
			line: `#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=128000,AVERAGE-BANDWIDTH=128000,SCORE=5.000,CODECS="avc1.4d400d,mp4a.40.2",` +
				`SUPPLEMENTAL-CODECS="mp4a.40.5",RESOLUTION=320x240,HDCP-LEVEL=NONE,` +
				`ALLOWED-CPC="com.example.drm1:SMART-TV/PC",VIDEO-RANGE=SDR,REQ-VIDEO-LAYOUT="CH-MONO",` +
				`STABLE-VARIANT-ID="a_0",VIDEO="video",PATHWAY-ID="X",` +
				`PROGRAM-ID=1,NAME="prop",URI="iframe.m3u8"`,
			error: false,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			vnt, err := parseExtXStreamInf(c.line, c.strict)
			if c.error {
				is.Equal(err != nil, true) // must return an error
				return
			}
			is.NoErr(err)
			out := bytes.Buffer{}
			writeExtXIFrameStreamInf(&out, vnt)
			outStr := trimLineEnd(out.String())
			is.Equal(c.line, outStr) // EXT-X-STREAM-INF line must match
		})
	}
}

func TestReadWriteSessionData(t *testing.T) {
	is := is.New(t)
	cases := []struct {
		desc  string
		line  string
		error bool
	}{
		{
			desc:  "minimal",
			line:  `#EXT-X-SESSION-DATA:DATA-ID="com.example.lyrics",VALUE="example",LANGUAGE="en"`,
			error: false,
		},
		{
			desc:  "bad tag",
			line:  `#EXT-X-SESSION-DAT:DATA-ID="com.example.lyrics",VALUE="example",LANGUAGE="en"`,
			error: true,
		},
		{
			desc:  "raw uri",
			line:  `#EXT-X-SESSION-DATA:DATA-ID="co.l",URI="dataURI",FORMAT=RAW,LANGUAGE="en"`,
			error: false,
		},
		{
			desc:  "bad format",
			line:  `#EXT-X-SESSION-DATA:DATA-ID="co.l",URI="dataURI",FORMAT=raw,LANGUAGE="en"`,
			error: true,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			sd, err := parseSessionData(c.line)
			if c.error {
				is.Equal(err != nil, true) // must return an error
				return
			}
			is.NoErr(err)
			out := bytes.Buffer{}
			writeSessionData(&out, sd)
			outStr := trimLineEnd(out.String())
			is.Equal(c.line, outStr) // EXT-X-SESSION-DATA line must match
		})
	}
}

func TestReadWriteExtXStart(t *testing.T) {
	is := is.New(t)
	cases := []struct {
		desc    string
		line    string
		start   float64
		precise bool
		error   bool
	}{
		{
			desc:    "positive precise",
			line:    `#EXT-X-START:TIME-OFFSET=10.000,PRECISE=YES`,
			start:   10,
			precise: true,
		},
		{
			desc:    "negative",
			line:    `#EXT-X-START:TIME-OFFSET=-5.720`,
			start:   -5.72,
			precise: false,
		},
		{
			desc:  "offset not a float",
			line:  `#EXT-X-START:TIME-OFFSET="5.72"`,
			error: true,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			start, precise, err := parseExtXStartParams(c.line[len("#EXT-X-START:"):])
			if c.error {
				is.Equal(err != nil, true) // must return an error
				return
			}
			is.NoErr(err)
			out := bytes.Buffer{}
			is.Equal(c.start, start)     // start time must match
			is.Equal(c.precise, precise) // precise must match
			writeExtXStart(&out, start, precise)
			outStr := trimLineEnd(out.String())
			is.Equal(c.line, outStr) // EXT-X-START line must match
		})
	}

}

// TestReadWriteMediaPlaylist tests reading and writing media playlists from sample-playlists
// Looks at verbatim match, so the order of tags and attributes must match.
func TestReadWritePlaylists(t *testing.T) {
	is := is.New(t)
	files := []string{
		"master-with-sessiondata.m3u8",
		"master-with-closed-captions.m3u8",
		"media-playlist-with-program-date-time.m3u8",
		"master-groups-and-iframe.m3u8",
		"media-playlist-with-multiple-dateranges.m3u8",
		"media-playlist-with-start-time.m3u8",
		"master-with-independent-segments.m3u8",
		"media-playlist-with-gap.m3u8",
	}

	for _, fileName := range files {
		t.Run(fileName, func(t *testing.T) {
			f, err := os.Open("sample-playlists/" + fileName)
			is.NoErr(err) // open file should succeed
			p, _, err := DecodeFrom(bufio.NewReader(f), true)
			is.NoErr(err) // decode playlist should succeed
			f.Close()
			got := trimLineEnd(p.String())
			// os.WriteFile("out.m3u8", []byte(out), 0644)
			inData, err := os.ReadFile("sample-playlists/" + fileName)
			is.NoErr(err) // read file should succeed
			want := trimLineEnd(strings.Replace(string(inData), "\r\n", "\n", -1))
			if got != want {
				t.Errorf("got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

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

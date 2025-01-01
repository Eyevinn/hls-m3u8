package m3u8

import (
	"bytes"
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

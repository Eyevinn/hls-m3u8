package m3u8

/*
Playlist parsing tests.
*/

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestDecodeMasterPlaylist(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master.m3u8")
	is.NoErr(err) // must open file
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err) // must decode playlist
	// check parsed values
	is.Equal(p.ver, uint8(3))    // version must be 3
	is.Equal(len(p.Variants), 5) // must be 5 variants
	// TODO check other values
	// fmt.Println(p.Encode().String())
}

func TestDecodeMasterPlaylistWithMultipleCodecs(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-with-multiple-codecs.m3u8")
	is.NoErr(err) // must open file
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err) // must decode playlist
	// check parsed values
	is.Equal(p.ver, uint8(3))    // version must be 3
	is.Equal(len(p.Variants), 5) // must be 5 variants
	for _, v := range p.Variants {
		is.Equal(v.Codecs, "avc1.42c015,mp4a.40.2") // codecs must be combined
	}
	// TODO check other values
	// fmt.Println(p.Encode().String())
}

func TestDecodeMasterPlaylistWithAlternatives(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-with-alternatives.m3u8")
	is.NoErr(err) // must open file
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err) // must decode playlist
	// check parsed values
	is.Equal(p.ver, uint8(3))    // version must be 3
	is.Equal(len(p.Variants), 4) // must be 4 variants
	// TODO check other values
	for i, v := range p.Variants {
		switch i {
		case 0, 1, 2:
			is.Equal(len(v.Alternatives), 3) // not all alternatives from #EXT-X-MEDIA parsed
		case 3:
			is.Equal(len(v.Alternatives), 0) // should not be alternatives for this variant
		default:
			t.Errorf("unexpected variant index: %d", i)
		}
	}
	// fmt.Println(p.Encode().String())
}

func TestDecodeMasterPlaylistWithAlternativesB(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-with-alternatives-b.m3u8")
	is.NoErr(err) // must open file
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err) // must decode playlist
	// check parsed values
	is.Equal(p.ver, uint8(3))    // version must be 3
	is.Equal(len(p.Variants), 4) // must be 4 variants
	// TODO check other values
	for i, v := range p.Variants {
		switch i {
		case 0, 1, 2:
			is.Equal(len(v.Alternatives), 3) // not all alternatives from #EXT-X-MEDIA parsed
		case 3:
			is.Equal(len(v.Alternatives), 0) // should not be alternatives for this variant
		default:
			t.Errorf("unexpected variant index: %d", i)
		}
	}
	// fmt.Println(p.Encode().String())
}

func TestDecodeMasterPlaylistWithClosedCaptionEqNone(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-with-closed-captions-eq-none.m3u8")
	is.NoErr(err) // must open file
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err) // must decode playlist

	// check parsed values
	is.Equal(p.ver, uint8(4)) // version must be 4
	for _, v := range p.Variants {
		is.Equal(v.Captions, "NONE") // all variants must have CLOSED-CAPTIONS=NONE
	}
}

// Decode a master playlist with Name tag in EXT-X-STREAM-INF
func TestDecodeMasterPlaylistWithStreamInfName(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-with-stream-inf-name.m3u8")
	is.NoErr(err) // must open file
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err) // must decode playlist
	for _, variant := range p.Variants {
		is.True(variant.Name != "") // name tag must not be empty
	}
}

func TestDecodeMediaPlaylistByteRange(t *testing.T) {
	f, _ := os.Open("sample-playlists/media-playlist-with-byterange.m3u8")
	p, _ := NewMediaPlaylist(3, 3)
	_ = p.DecodeFrom(bufio.NewReader(f), true)
	expected := []*MediaSegment{
		{URI: "video.ts", Duration: 10, Limit: 75232, SeqId: 0},
		{URI: "video.ts", Duration: 10, Limit: 82112, Offset: 752321, SeqId: 1},
		{URI: "video.ts", Duration: 10, Limit: 69864, SeqId: 2},
	}
	for i, seg := range p.Segments {
		if !reflect.DeepEqual(*seg, *expected[i]) {
			t.Errorf("exp: %+v\ngot: %+v", expected[i], seg)
		}
	}
}

// Decode a master playlist with i-frame-stream-inf
func TestDecodeMasterPlaylistWithIFrameStreamInf(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-with-i-frame-stream-inf.m3u8")
	is.NoErr(err) // must open file
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err) // must decode playlist
	expected := map[int]*Variant{
		86000: {URI: "low/iframe.m3u8", VariantParams: VariantParams{Bandwidth: 86000, Codecs: "c1",
			Resolution: "1x1", Video: "1", Iframe: true}},
		150000: {URI: "mid/iframe.m3u8", VariantParams: VariantParams{Bandwidth: 150000, Codecs: "c2",
			Resolution: "2x2", Video: "2", Iframe: true}},
		550000: {URI: "hi/iframe.m3u8", VariantParams: VariantParams{Bandwidth: 550000, Codecs: "c2",
			Resolution: "2x2", Video: "2", Iframe: true}},
	}
	for _, variant := range p.Variants {
		for k, expect := range expected {
			if reflect.DeepEqual(*variant, *expect) {
				delete(expected, k)
			}
		}
	}
	for _, expect := range expected {
		t.Errorf("not found:%+v", expect)
	}
}

func TestDecodeMasterPlaylistWithStreamInfAverageBandwidth(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-with-stream-inf-1.m3u8")
	is.NoErr(err) // must open file
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err) // must decode playlist
	for _, variant := range p.Variants {
		is.True(variant.AverageBandwidth > 0) // average bandwidth must be greater than 0
	}
}

func TestDecodeMasterPlaylistWithStreamInfFrameRate(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-with-stream-inf-1.m3u8")
	is.NoErr(err) // must open file
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err) // must decode playlist
	for _, variant := range p.Variants {
		is.True(variant.VariantParams.FrameRate > 0) // frame rate must be greater than 0
	}
}

func TestDecodeMasterPlaylistWithIndependentSegments(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-with-independent-segments.m3u8")
	is.NoErr(err) // must open file
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err)                    // must decode playlist
	is.True(p.IndependentSegments()) // independent segments must be true
}

func TestDecodeMasterWithHLSV7(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-with-hlsv7.m3u8")
	is.NoErr(err) // must open file
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err) // must decode playlist
	var unexpected []*Variant
	expected := map[string]VariantParams{
		"sdr_720/prog_index.m3u8": {Bandwidth: 3971374, AverageBandwidth: 2778321, Codecs: "hvc1.2.4.L123.B0",
			Resolution: "1280x720", Captions: "NONE", VideoRange: "SDR", HDCPLevel: "NONE", FrameRate: 23.976},
		"sdr_1080/prog_index.m3u8": {Bandwidth: 10022043, AverageBandwidth: 6759875, Codecs: "hvc1.2.4.L123.B0",
			Resolution: "1920x1080", Captions: "NONE", VideoRange: "SDR", HDCPLevel: "TYPE-0", FrameRate: 23.976},
		"sdr_2160/prog_index.m3u8": {Bandwidth: 28058971, AverageBandwidth: 20985770, Codecs: "hvc1.2.4.L150.B0",
			Resolution: "3840x2160", Captions: "NONE", VideoRange: "SDR", HDCPLevel: "TYPE-1", FrameRate: 23.976},
		"dolby_720/prog_index.m3u8": {Bandwidth: 5327059, AverageBandwidth: 3385450, Codecs: "dvh1.05.01",
			Resolution: "1280x720", Captions: "NONE", VideoRange: "PQ", HDCPLevel: "NONE", FrameRate: 23.976},
		"dolby_1080/prog_index.m3u8": {Bandwidth: 12876596, AverageBandwidth: 7999361, Codecs: "dvh1.05.03",
			Resolution: "1920x1080", Captions: "NONE", VideoRange: "PQ", HDCPLevel: "TYPE-0", FrameRate: 23.976},
		"dolby_2160/prog_index.m3u8": {Bandwidth: 30041698, AverageBandwidth: 24975091, Codecs: "dvh1.05.06",
			Resolution: "3840x2160", Captions: "NONE", VideoRange: "PQ", HDCPLevel: "TYPE-1", FrameRate: 23.976},
		"hdr10_720/prog_index.m3u8": {Bandwidth: 5280654, AverageBandwidth: 3320040, Codecs: "hvc1.2.4.L123.B0",
			Resolution: "1280x720", Captions: "NONE", VideoRange: "PQ", HDCPLevel: "NONE", FrameRate: 23.976},
		"hdr10_1080/prog_index.m3u8": {Bandwidth: 12886714, AverageBandwidth: 7964551, Codecs: "hvc1.2.4.L123.B0",
			Resolution: "1920x1080", Captions: "NONE", VideoRange: "PQ", HDCPLevel: "TYPE-0", FrameRate: 23.976},
		"hdr10_2160/prog_index.m3u8": {Bandwidth: 29983769, AverageBandwidth: 24833402, Codecs: "hvc1.2.4.L150.B0",
			Resolution: "3840x2160", Captions: "NONE", VideoRange: "PQ", HDCPLevel: "TYPE-1", FrameRate: 23.976},
		"sdr_720/iframe_index.m3u8": {Bandwidth: 593626, AverageBandwidth: 248586, Codecs: "hvc1.2.4.L123.B0",
			Resolution: "1280x720", Iframe: true, VideoRange: "SDR", HDCPLevel: "NONE"},
		"sdr_1080/iframe_index.m3u8": {Bandwidth: 956552, AverageBandwidth: 399790, Codecs: "hvc1.2.4.L123.B0",
			Resolution: "1920x1080", Iframe: true, VideoRange: "SDR", HDCPLevel: "TYPE-0"},
		"sdr_2160/iframe_index.m3u8": {Bandwidth: 1941397, AverageBandwidth: 826971, Codecs: "hvc1.2.4.L150.B0",
			Resolution: "3840x2160", Iframe: true, VideoRange: "SDR", HDCPLevel: "TYPE-1"},
		"dolby_720/iframe_index.m3u8": {Bandwidth: 573073, AverageBandwidth: 232253, Codecs: "dvh1.05.01",
			Resolution: "1280x720", Iframe: true, VideoRange: "PQ", HDCPLevel: "NONE"},
		"dolby_1080/iframe_index.m3u8": {Bandwidth: 905037, AverageBandwidth: 365337, Codecs: "dvh1.05.03",
			Resolution: "1920x1080", Iframe: true, VideoRange: "PQ", HDCPLevel: "TYPE-0"},
		"dolby_2160/iframe_index.m3u8": {Bandwidth: 1893236, AverageBandwidth: 739114, Codecs: "dvh1.05.06",
			Resolution: "3840x2160", Iframe: true, VideoRange: "PQ", HDCPLevel: "TYPE-1"},
		"hdr10_720/iframe_index.m3u8": {Bandwidth: 572673, AverageBandwidth: 232511, Codecs: "hvc1.2.4.L123.B0",
			Resolution: "1280x720", Iframe: true, VideoRange: "PQ", HDCPLevel: "NONE"},
		"hdr10_1080/iframe_index.m3u8": {Bandwidth: 905053, AverageBandwidth: 364552, Codecs: "hvc1.2.4.L123.B0",
			Resolution: "1920x1080", Iframe: true, VideoRange: "PQ", HDCPLevel: "TYPE-0"},
		"hdr10_2160/iframe_index.m3u8": {Bandwidth: 1895477, AverageBandwidth: 739757, Codecs: "hvc1.2.4.L150.B0",
			Resolution: "3840x2160", Iframe: true, VideoRange: "PQ", HDCPLevel: "TYPE-1"},
	}
	for _, variant := range p.Variants {
		var found bool
		for uri, vp := range expected {
			if variant == nil || variant.URI != uri {
				continue
			}
			if reflect.DeepEqual(variant.VariantParams, vp) {
				delete(expected, uri)
				found = true
			}
		}
		if !found {
			unexpected = append(unexpected, variant)
		}
	}
	for uri, expect := range expected {
		t.Errorf("not found: uri=%q %+v", uri, expect)
	}
	for _, unexpect := range unexpected {
		t.Errorf("found but not expecting:%+v", unexpect)
	}
}

func TestReadBadMasterPlaylists(t *testing.T) {
	is := is.New(t)
	cases := []struct {
		desc     string
		playlist string
	}{
		{
			desc: "bad session data",
			playlist: `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-SESSION-DATA:DATA-ID="com.example.title",VALUE="This is an example title",FORMAT=bad
#EXT-X-INF:BANDWIDTH=1280000
video.m3u8
		`,
		},
		{
			desc: "bad define",
			playlist: `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-DEFINE:NAME="example.com"
#EXT-X-INF:BANDWIDTH=1280000
video.m3u8
		`,
		},
		{
			desc: "bad start",
			playlist: `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-START:TIME-OFFSET=bad
#EXT-X-INF:BANDWIDTH=1280000
video.m3u8
		`,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			p := NewMasterPlaylist()
			err := p.DecodeFrom(bufio.NewReader(bytes.NewBufferString(c.playlist)), true)
			is.True(err != nil) // must return an error
		})
	}
}

/****************************
 * Begin Test MediaPlaylist *
 ****************************/

func TestDecodeMediaPlaylist(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/wowza-vod-chunklist.m3u8")
	is.NoErr(err) // must open file
	p, err := NewMediaPlaylist(5, 798)
	is.NoErr(err) // must create playlist
	err = p.DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err) // must decode playlist

	// check parsed values
	is.Equal(p.ver, uint8(3))            // version must be 3
	is.Equal(p.TargetDuration, uint(12)) // target duration must be 12
	is.True(p.Closed)                    // closed (VOD) playlist but Close field = false")
	titles := []string{"Title 1", "Title 2", ""}
	for i, s := range p.Segments {
		if i > len(titles)-1 {
			break
		}
		is.Equal(s.Title, titles[i]) // title must be "Title 1", "Title 2", ""
	}
	is.Equal(p.Count(), uint(522)) // segment count must be 522
	var seqId, idx uint
	for seqId, idx = 1, 0; idx < p.Count(); seqId, idx = seqId+1, idx+1 {
		is.Equal(p.Segments[idx].SeqId, uint64(seqId)) // SeqId must match
	}
	// TODO check other values…
	//fmt.Println(p.Encode().String()), stream.Name}
}

func TestDecodeMediaPlaylistExtInfNonStrict2(t *testing.T) {
	is := is.New(t)
	header := `#EXTM3U
#EXT-X-TARGETDURATION:10
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
%s
1.ts
`

	tests := []struct {
		strict      bool
		extInf      string
		wantError   bool
		wantSegment *MediaSegment
	}{
		// strict mode on
		{true, "#EXTINF:10.000,", false, &MediaSegment{Duration: 10.0, Title: "", URI: "1.ts"}},
		{true, "#EXTINF:10.000,Title", false, &MediaSegment{Duration: 10.0, Title: "Title", URI: "1.ts"}},
		{true, "#EXTINF:10.000,Title,Track", false, &MediaSegment{Duration: 10.0, Title: "Title,Track", URI: "1.ts"}},
		{true, "#EXTINF:invalid,", true, nil},
		{true, "#EXTINF:10.000", true, nil},

		// strict mode off
		{false, "#EXTINF:10.000,", false, &MediaSegment{Duration: 10.0, Title: "", URI: "1.ts"}},
		{false, "#EXTINF:10.000,Title", false, &MediaSegment{Duration: 10.0, Title: "Title", URI: "1.ts"}},
		{false, "#EXTINF:10.000,Title,Track", false, &MediaSegment{Duration: 10.0, Title: "Title,Track", URI: "1.ts"}},
		{false, "#EXTINF:invalid,", false, &MediaSegment{Duration: 0.0, Title: "", URI: "1.ts"}},
		{false, "#EXTINF:10.000", false, &MediaSegment{Duration: 10.0, Title: "", URI: "1.ts"}},
	}

	for nr, test := range tests {
		p, err := NewMediaPlaylist(1, 1)
		is.NoErr(err) // create playlist

		reader := bytes.NewBufferString(fmt.Sprintf(header, test.extInf))
		err = p.DecodeFrom(reader, test.strict)
		if test.wantError {
			is.True(err != nil) // must return an error
			continue
		}
		is.NoErr(err) // must decode playlist
		if !reflect.DeepEqual(p.Segments[0], test.wantSegment) {
			t.Errorf("\nnr %d: have: %+v\nwant: %+v", nr, p.Segments[0], test.wantSegment)
		}
	}
}

func TestDecodeMasterPlaylistWithAutodetection(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master.m3u8")
	is.NoErr(err) // must open file
	m, listType, err := DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err)              // must decode playlist
	is.Equal(listType, MASTER) // must be master playlist
	mp := m.(*MasterPlaylist)
	// fmt.Printf(">%+v\n", mp)
	// for _, v := range mp.Variants {
	//	fmt.Printf(">>%+v +v\n", v)
	// }
	//fmt.Println("Type below must be MasterPlaylist:")
	CheckType(t, mp)
}

func TestDecodeMediaPlaylistWithAutodetection(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/wowza-vod-chunklist.m3u8")
	is.NoErr(err) // must open file
	p, listType, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err) // must decode playlist
	pp := p.(*MediaPlaylist)
	CheckType(t, pp)
	is.Equal(listType, MEDIA) // must be media playlist
	// check parsed values
	is.Equal(pp.TargetDuration, uint(12)) // target duration must be 12
	is.True(pp.Closed)                    // closed (VOD) playlist but Close field = false")
	is.Equal(pp.winsize, uint(0))         // window size must be 0
	// TODO check other values…
	// fmt.Println(pp.Encode().String())
}

// TestDecodeMediaPlaylistAutoDetectExtend tests a very large playlist auto
// extends to the appropriate size.
func TestDecodeMediaPlaylistAutoDetectExtend(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/media-playlist-large.m3u8")
	is.NoErr(err) // must open file
	p, listType, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err) // must decode playlist
	pp := p.(*MediaPlaylist)
	CheckType(t, pp)
	is.Equal(listType, MEDIA) // must be media playlist
	var exp uint = 40001
	is.Equal(pp.Count(), exp) // segment count must be 40001
}

// Test for FullTimeParse of EXT-X-PROGRAM-DATE-TIME
// We testing ISO/IEC 8601:2004 where we can get time in UTC, UTC with Nanoseconds
// timeZone in formats '±00:00', '±0000', '±00'
// m3u8.FullTimeParse()
func TestFullTimeParse(t *testing.T) {
	var timestamps = []struct {
		name  string
		value string
	}{
		{"time_in_utc", "2006-01-02T15:04:05Z"},
		{"time_in_utc_nano", "2006-01-02T15:04:05.123456789Z"},
		{"time_with_positive_zone_and_colon", "2006-01-02T15:04:05+01:00"},
		{"time_with_positive_zone_no_colon", "2006-01-02T15:04:05+0100"},
		{"time_with_positive_zone_2digits", "2006-01-02T15:04:05+01"},
		{"time_with_negative_zone_and_colon", "2006-01-02T15:04:05-01:00"},
		{"time_with_negative_zone_no_colon", "2006-01-02T15:04:05-0100"},
		{"time_with_negative_zone_2digits", "2006-01-02T15:04:05-01"},
	}

	var err error
	for _, tstamp := range timestamps {
		_, err = FullTimeParse(tstamp.value)
		if err != nil {
			t.Errorf("FullTimeParse Error at %s [%s]: %s", tstamp.name, tstamp.value, err)
		}
	}
}

// Test for StrictTimeParse of EXT-X-PROGRAM-DATE-TIME
// We testing Strict format of RFC3339 where we can get time in UTC, UTC with Nanoseconds
// timeZone in formats '±00:00', '±0000', '±00'
// m3u8.StrictTimeParse()
func TestStrictTimeParse(t *testing.T) {
	var timestamps = []struct {
		name  string
		value string
	}{
		{"time_in_utc", "2006-01-02T15:04:05Z"},
		{"time_in_utc_nano", "2006-01-02T15:04:05.123456789Z"},
		{"time_with_positive_zone_and_colon", "2006-01-02T15:04:05+01:00"},
		{"time_with_negative_zone_and_colon", "2006-01-02T15:04:05-01:00"},
	}

	var err error
	for _, tstamp := range timestamps {
		_, err = StrictTimeParse(tstamp.value)
		if err != nil {
			t.Errorf("StrictTimeParse Error at %s [%s]: %s", tstamp.name, tstamp.value, err)
		}
	}
}

func TestMediaPlaylistWithOATCLSSCTE35Tag(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/media-playlist-with-oatcls-scte35.m3u8")
	is.NoErr(err) // must open file
	p, _, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err) // must decode playlist
	pp := p.(*MediaPlaylist)

	expect := map[int]*SCTE{
		0: {Syntax: SCTE35_OATCLS, CueType: SCTE35Cue_Start,
			Cue: "/DAlAAAAAAAAAP/wFAUAAAABf+/+ANgNkv4AFJlwAAEBAQAA5xULLA==", Time: 15},
		1: {Syntax: SCTE35_OATCLS, CueType: SCTE35Cue_Mid,
			Cue: "/DAlAAAAAAAAAP/wFAUAAAABf+/+ANgNkv4AFJlwAAEBAQAA5xULLA==", Time: 15, Elapsed: 8.844},
		2: {Syntax: SCTE35_OATCLS, CueType: SCTE35Cue_End},
	}
	for i := 0; i < int(pp.Count()); i++ {
		if !reflect.DeepEqual(pp.Segments[i].SCTE, expect[i]) {
			t.Errorf("OATCLS SCTE35 segment %v (uri: %v)\ngot: %#v\nexp: %#v",
				i, pp.Segments[i].URI, pp.Segments[i].SCTE, expect[i],
			)
		}
	}
}

func TestMediaPlaylistWithDateRangeSCTE35Tag(t *testing.T) {
	f, err := os.Open("sample-playlists/media-playlist-with-scte35-daterange.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	p, _, err := DecodeFrom(bufio.NewReader(f), true)
	if err != nil {
		t.Fatal(err)
	}
	pp := p.(*MediaPlaylist)
	startDateOut, _ := time.Parse(DATETIME, "2014-03-05T11:15:00Z")
	startDateIn, _ := time.Parse(DATETIME, "2014-03-05T11:16:00Z")

	ptr := func(f float64) *float64 {
		return &f
	}

	expect := map[int]*DateRange{
		0: {
			ID:              "SPLICE-6FFFFFF0",
			Class:           "",
			StartDate:       startDateOut,
			EndDate:         nil,
			Duration:        ptr(60.0),
			PlannedDuration: ptr(60.0),
			SCTE35Out:       "0xFC002F0000000000FF00",
		},
		1: {
			ID:              "SPLICE-6FFFFFF0",
			Class:           "",
			StartDate:       startDateIn,
			EndDate:         nil,
			Duration:        ptr(60.0),
			PlannedDuration: ptr(60.0),
			SCTE35In:        "0xFC002F0000000000FF10",
		},
	}

	actual := make([]*DateRange, 0, 2)
	for i := 0; i < int(pp.Count()); i++ {
		actual = append(actual, pp.Segments[i].SCTE35DateRanges...)
	}

	for i := 0; i < len(expect); i++ {
		if !reflect.DeepEqual(actual[i], expect[i]) {
			t.Errorf("DATERANGE SCTE35 segment %v \ngot: %#v\nexp: %#v",
				i, actual[i], expect[i],
			)
		}
	}
}

func TestDecodeMediaPlaylistWithDiscontinuitySeq(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/media-playlist-with-discontinuity-seq.m3u8")
	is.NoErr(err) // must open file
	p, listType, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err) // must decode playlist
	pp := p.(*MediaPlaylist)
	CheckType(t, pp)
	is.Equal(listType, MEDIA)         // must be media playlist
	is.True(pp.DiscontinuitySeq != 0) // discontinuity sequence must be non-zeo
	is.Equal(pp.Count(), uint(4))     // segment count must be 4
	is.Equal(pp.SeqNo, uint64(0))     // sequence number must be 0
	var seqId, idx uint
	for seqId, idx = 0, 0; idx < pp.Count(); seqId, idx = seqId+1, idx+1 {
		if pp.Segments[idx].SeqId != uint64(seqId) {
			t.Errorf("Excepted SeqId for %vth segment: %v, got: %v", idx+1, seqId, pp.Segments[idx].SeqId)
		}
	}
}

func TestDecodeMasterPlaylistWithCustomTags(t *testing.T) {
	is := is.New(t)
	cases := []struct {
		src                  string
		customDecoders       []CustomDecoder
		expectedError        string
		expectedPlaylistTags []string
	}{
		{
			src:                  "sample-playlists/master-playlist-with-custom-tags.m3u8",
			customDecoders:       nil,
			expectedError:        "",
			expectedPlaylistTags: nil,
		},
		{
			src: "sample-playlists/master-playlist-with-custom-tags.m3u8",
			customDecoders: []CustomDecoder{
				&MockCustomTag{
					name:          "#CUSTOM-PLAYLIST-TAG:",
					err:           fmt.Errorf("Error decoding tag"),
					segment:       false,
					encodedString: "#CUSTOM-PLAYLIST-TAG:42",
				},
			},
			expectedError:        "Error decoding tag",
			expectedPlaylistTags: nil,
		},
		{
			src: "sample-playlists/master-playlist-with-custom-tags.m3u8",
			customDecoders: []CustomDecoder{
				&MockCustomTag{
					name:          "#CUSTOM-PLAYLIST-TAG:",
					err:           nil,
					segment:       false,
					encodedString: "#CUSTOM-PLAYLIST-TAG:42",
				},
			},
			expectedError: "",
			expectedPlaylistTags: []string{
				"#CUSTOM-PLAYLIST-TAG:",
			},
		},
	}

	for _, testCase := range cases {
		f, err := os.Open(testCase.src)
		is.NoErr(err) // must open file

		p, listType, err := DecodeWith(bufio.NewReader(f), true, testCase.customDecoders)

		if testCase.expectedError != "" {
			is.True(err != nil) // must return an error
			is.Equal(err.Error(), testCase.expectedError)
			continue
		}

		pp := p.(*MasterPlaylist)

		CheckType(t, pp)
		is.Equal(listType, MASTER) // must be master playlist

		if len(pp.Custom) != len(testCase.expectedPlaylistTags) {
			t.Errorf("Did not parse expected number of custom tags. Got: %d Expected: %d", len(pp.Custom),
				len(testCase.expectedPlaylistTags))
		} else {
			// we have the same count, lets confirm its the right tags
			for _, expectedTag := range testCase.expectedPlaylistTags {
				if _, ok := pp.Custom[expectedTag]; !ok {
					t.Errorf("Did not parse custom tag %s", expectedTag)
				}
			}
		}
	}
}

func TestDecodeMediaPlaylistWithCustomTags(t *testing.T) {
	is := is.New(t)
	cases := []struct {
		src                  string
		customDecoders       []CustomDecoder
		expectedError        string
		expectedPlaylistTags []string
		expectedSegmentTags  []*struct {
			index int
			names []string
		}
	}{
		{
			src:                  "sample-playlists/media-playlist-with-custom-tags.m3u8",
			customDecoders:       nil,
			expectedError:        "",
			expectedPlaylistTags: nil,
			expectedSegmentTags:  nil,
		},
		{
			src: "sample-playlists/media-playlist-with-custom-tags.m3u8",
			customDecoders: []CustomDecoder{
				&MockCustomTag{
					name:          "#CUSTOM-PLAYLIST-TAG:",
					err:           errors.New("Error decoding tag"),
					segment:       false,
					encodedString: "#CUSTOM-PLAYLIST-TAG:42",
				},
			},
			expectedError:        "Error decoding tag",
			expectedPlaylistTags: nil,
			expectedSegmentTags:  nil,
		},
		{
			src: "sample-playlists/media-playlist-with-custom-tags.m3u8",
			customDecoders: []CustomDecoder{
				&MockCustomTag{
					name:          "#CUSTOM-PLAYLIST-TAG:",
					err:           nil,
					segment:       false,
					encodedString: "#CUSTOM-PLAYLIST-TAG:42",
				},
				&MockCustomTag{
					name:          "#CUSTOM-SEGMENT-TAG:",
					err:           nil,
					segment:       true,
					encodedString: "#CUSTOM-SEGMENT-TAG:NAME=\"Yoda\",JEDI=YES",
				},
				&MockCustomTag{
					name:          "#CUSTOM-SEGMENT-TAG-B",
					err:           nil,
					segment:       true,
					encodedString: "#CUSTOM-SEGMENT-TAG-B",
				},
			},
			expectedError: "",
			expectedPlaylistTags: []string{
				"#CUSTOM-PLAYLIST-TAG:",
			},
			expectedSegmentTags: []*struct {
				index int
				names []string
			}{
				{1, []string{"#CUSTOM-SEGMENT-TAG:"}},
				{2, []string{"#CUSTOM-SEGMENT-TAG:", "#CUSTOM-SEGMENT-TAG-B"}},
			},
		},
	}

	for _, testCase := range cases {
		f, err := os.Open(testCase.src)
		is.NoErr(err) // must open file
		p, listType, err := DecodeWith(bufio.NewReader(f), true, testCase.customDecoders)

		if testCase.expectedError != "" {
			is.True(err != nil) // must return an error
			is.Equal(err.Error(), testCase.expectedError)
			continue
		}

		pp := p.(*MediaPlaylist)

		CheckType(t, pp)

		is.Equal(listType, MEDIA) // must be media playlist

		if len(pp.Custom) != len(testCase.expectedPlaylistTags) {
			t.Errorf("Did not parse expected number of custom tags. Got: %d Expected: %d", len(pp.Custom),
				len(testCase.expectedPlaylistTags))
		} else {
			// we have the same count, lets confirm its the right tags
			for _, expectedTag := range testCase.expectedPlaylistTags {
				if _, ok := pp.Custom[expectedTag]; !ok {
					t.Errorf("Did not parse custom tag %s", expectedTag)
				}
			}
		}

		var expectedSegmentTag *struct {
			index int
			names []string
		}

		expectedIndex := 0

		for i := 0; i < int(pp.Count()); i++ {
			seg := pp.Segments[i]
			if expectedIndex != len(testCase.expectedSegmentTags) {
				expectedSegmentTag = testCase.expectedSegmentTags[expectedIndex]
			} else {
				// we are at the end of the expectedSegmentTags list, the rest of the segments
				// should have no custom tags
				expectedSegmentTag = nil
			}

			if expectedSegmentTag == nil || expectedSegmentTag.index != i {
				if len(seg.Custom) != 0 {
					t.Errorf("Did not parse expected number of custom tags on Segment %d. Got: %d Expected: %d",
						i, len(seg.Custom), 0)
				}
				continue
			}

			// We are now checking the segment corresponding to expectedSegmentTag
			// increase our expectedIndex for next iteration
			expectedIndex++

			if len(expectedSegmentTag.names) != len(seg.Custom) {
				t.Errorf("Did not parse expected number of custom tags on Segment %d. Got: %d Expected: %d", i,
					len(seg.Custom), len(expectedSegmentTag.names))
			} else {
				// we have the same count, lets confirm its the right tags
				for _, expectedTag := range expectedSegmentTag.names {
					if _, ok := seg.Custom[expectedTag]; !ok {
						t.Errorf("Did not parse customTag %s on Segment %d", expectedTag, i)
					}
				}
			}
		}

		if expectedIndex != len(testCase.expectedSegmentTags) {
			t.Errorf("Did not parse custom tags on all expected segments. Parsed Segments: %d Expected: %d",
				expectedIndex, len(testCase.expectedSegmentTags))
		}
	}
}

func TestMediaPlaylistWithSCTE35Tag(t *testing.T) {
	cases := []struct {
		playlistLocation  string
		expectedSCTEIndex int
		expectedSCTECue   string
		expectedSCTEID    string
		expectedSCTETime  float64
	}{
		{
			"sample-playlists/media-playlist-with-scte35.m3u8",
			2,
			"/DAIAAAAAAAAAAAQAAZ/I0VniQAQAgBDVUVJQAAAAH+cAAAAAA==",
			"123",
			123.12,
		},
		{
			"sample-playlists/media-playlist-with-scte35-1.m3u8",
			1,
			"/DAIAAAAAAAAAAAQAAZ/I0VniQAQAgBDVUVJQAA",
			"",
			0,
		},
	}
	for _, c := range cases {
		f, _ := os.Open(c.playlistLocation)
		playlist, _, _ := DecodeFrom(bufio.NewReader(f), true)
		mediaPlaylist := playlist.(*MediaPlaylist)
		for index, item := range mediaPlaylist.Segments {
			if item == nil {
				break
			}
			if index != c.expectedSCTEIndex && item.SCTE != nil {
				t.Error("Not expecting SCTE information on this segment")
			} else if index == c.expectedSCTEIndex && item.SCTE == nil {
				t.Error("Expecting SCTE information on this segment")
			} else if index == c.expectedSCTEIndex && item.SCTE != nil {
				if (*item.SCTE).Cue != c.expectedSCTECue {
					t.Error("Expected ", c.expectedSCTECue, " got ", (*item.SCTE).Cue)
				} else if (*item.SCTE).ID != c.expectedSCTEID {
					t.Error("Expected ", c.expectedSCTEID, " got ", (*item.SCTE).ID)
				} else if (*item.SCTE).Time != c.expectedSCTETime {
					t.Error("Expected ", c.expectedSCTETime, " got ", (*item.SCTE).Time)
				}
			}
		}
	}
}

func TestDecodeLowLatencyMediaPlaylistWithNonZeroInitialSequenceNumber(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/media-playlist-low-latency-initial-sequence-num-5.m3u8")
	is.NoErr(err) // must open file
	p, listType, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err) // must decode playlist
	pp := p.(*MediaPlaylist)
	CheckType(t, pp)
	is.Equal(listType, MEDIA)              // must be media playlist
	is.Equal(pp.TargetDuration, uint(4))   // target duration must be 4
	is.True(!pp.Closed)                    // live playlist
	is.Equal(pp.SeqNo, uint64(11))         // sequence number must be 11
	is.Equal(pp.Count(), uint(16))         // segment count must be 16
	is.Equal(pp.PartTargetDuration, 1.002) // partial target duration must be 1.002
	startIndex := uint(5)
	pp.Map.URI = fmt.Sprintf("fileSequence%d.m4s", startIndex) // init segment should be fileSequence5.m4s

	for i := uint(0); i < pp.Count(); i++ {
		s := pp.Segments[i]
		is.True(pp.IsSegmentReady(s.URI)) // all segments should be ready
		// segment names should be in the following format fileSequence%d.m4s
		expectedSeqNo := startIndex + uint(pp.SeqNo) + i + 1
		expected := fmt.Sprintf("fileSequence%d.m4s", expectedSeqNo)
		if s.URI != expected {
			t.Errorf("Segment name mismatch: %s != %s", s.URI, expected)
		}
	}

	// Existing segments
	is.Equal(pp.LastSegIndex(), uint64(27))          // Last segment (fileSequence32.m4s) has index 27
	is.True(pp.IsSegmentReady("fileSequence32.m4s")) // it should be ready

	// Existing partial segments
	is.Equal(pp.LastPartSegIndex(), uint64(1))     // Last partial segment (filePart33.2.m4s) has index 1
	is.True(pp.IsSegmentReady("filePart33.2.m4s")) // it should be ready

	seq, part := pp.GetNextSequenceAndPart()
	is.Equal(seq, uint64(27))                       // Has seqId 27
	is.Equal(part, uint64(2))                       // Has index 2
	is.True(!pp.IsSegmentReady("filePart33.3.m4s")) // it should not be ready
}

func TestDecodeLowLatencyMediaPlaylist(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/media-playlist-low-latency.m3u8")
	is.NoErr(err) // must open file
	p, listType, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err) // must decode playlist
	pp := p.(*MediaPlaylist)
	CheckType(t, pp)
	is.Equal(listType, MEDIA)                  // must be media playlist
	is.Equal(pp.TargetDuration, uint(4))       // target duration must be 4
	is.True(!pp.Closed)                        // live playlist
	is.Equal(pp.Count(), uint(8))              // segment count must be 8
	is.Equal(pp.PartTargetDuration, 1.002)     // partial target duration must be 1.002
	is.Equal(len(pp.PartialSegments), int(10)) // partial segment count must be 10

	// The ProgramDateTime of the 4th segment should be: 2025-02-10T14:42:30.134Z
	st, _ := time.Parse(time.RFC3339, "2025-02-10T14:43:10.134+00:00")
	if !pp.Segments[3].ProgramDateTime.Equal(st) {
		t.Errorf("The program date time of the 1st segment should be: %v, actual value: %v",
			st, pp.Segments[1].ProgramDateTime)
	}

	for i, ps := range pp.PartialSegments {
		is.Equal(ps.Duration, float64(1.0)) // partial segment duration must be 1.0
		if i < 8 {
			is.True(ps.Independent) // the first 8 partial segment must be independent
		}
		is.True(strings.HasPrefix(ps.URI, "filePart")) // partial segment names should be in the following format filePart%d.%d.m4s
	}

	// The ProgramDateTime of the 9th partial segment should be: 2025-02-10T14:43:30.134Z
	st, _ = time.Parse(time.RFC3339, "2025-02-10T14:43:30.134+00:00")
	if !pp.PartialSegments[8].ProgramDateTime.Equal(st) {
		t.Errorf("The program date time of the 8st partial segment should be: %v, actual value: %v",
			st, pp.PartialSegments[8].ProgramDateTime)
	}

	// Preload Hints
	is.Equal(pp.PreloadHints.Type, "PART")             // preload hints type must be PART
	is.Equal(pp.PreloadHints.URI, "filePart251.3.m4s") // preload hints uri must be filePart251.3.m4s

	// Existing segments
	is.Equal(pp.LastSegIndex(), uint64(250)) // Last segment index (fileSequence250.m4s has index 250)
	// Existing partial segments
	is.Equal(pp.LastPartSegIndex(), uint64(1)) // Last partial segment index (filePart251.2.m4s has index 1)

	seq, part := pp.GetNextSequenceAndPart()
	is.Equal(seq, uint64(250)) // Has seqId 250
	is.Equal(part, uint64(2))  // Has index 2

	{
		// Test for appending partial segments
		_ = pp.AppendPartial("filePart251.3.m4s", 1.0, false)
		is.Equal(pp.LastPartSegIndex(), uint64(2)) // Last partial segment index (filePart251.3.m4s has index 2)
		is.True(pp.IsSegmentReady("filePart251.3.m4s"))

		_ = pp.AppendPartial("filePart251.4.m4s", 1.0, false)
		_ = pp.Append("fileSequence251.m4s", 1.0, "")
		is.Equal(pp.LastSegIndex(), uint64(250))   // Last segment index (filePart250.m4s has index 250)
		is.Equal(pp.LastPartSegIndex(), uint64(3)) // Last partial segment index (filePart251.4.m4s has index 3)
		is.True(pp.IsSegmentReady("filePart251.4.m4s"))
		is.True(pp.IsSegmentReady("fileSequence251.m4s"))

		// Rolled over
		seq, part = pp.GetNextSequenceAndPart()
		is.Equal(seq, uint64(251)) // Has seqId 251
		is.Equal(part, uint64(0))  // Has index 0

		// Add more
		_ = pp.AppendPartial("filePart252.1.m4s", 1.0, false)
		is.Equal(pp.LastSegIndex(), uint64(251))   // Last segment index (fileSequence251.m4s has index 251)
		is.Equal(pp.LastPartSegIndex(), uint64(0)) // Last partial segment index (filePart252.1.m4s has index 0)

		seq, part = pp.GetNextSequenceAndPart()
		is.Equal(seq, uint64(251)) // Has seqId 251
		is.Equal(part, uint64(1))  // Has index 1
	}
}

func TestDecodeMediaPlaylistWithSkip(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/media-playlist-with-skip.m3u8")
	is.NoErr(err) // must open file
	p, listType, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err) // must decode playlist
	pp := p.(*MediaPlaylist)
	CheckType(t, pp)
	is.Equal(listType, MEDIA)                  // must be media playlist
	is.Equal(pp.TargetDuration, uint(4))       // target duration must be 4
	is.True(!pp.Closed)                        // live playlist
	is.Equal(pp.Count(), uint(6))              // segment count must be 6
	is.Equal(pp.PartTargetDuration, 0.334)     // partial target duration must be 0.334
	is.Equal(len(pp.PartialSegments), int(28)) // partial segment count must be 24
	is.Equal(pp.SeqNo, uint64(264))            // seqNo is not 264
	is.Equal(pp.SkippedSegments(), uint64(3))  // skipped segments must be 3
	is.Equal(pp.ServerControl.CanSkipUntil,
		float64(pp.TargetDuration*6)) // Can skip 6 segments

	seq, part := pp.GetNextSequenceAndPart() // filePart273.4.mp4
	is.Equal(seq, uint64(273))               // Has seqId 273
	is.Equal(part, uint64(4))                // Has index 4

	// Try decoding the playlist with skip
	_, err = pp.EncodeWithSkip(1)
	is.True(err != nil) // must return an error
}

func TestDecodeMediaPlaylistWithProgramDateTime(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/media-playlist-with-program-date-time.m3u8")
	is.NoErr(err) // must open file
	p, listType, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err) // must decode playlist
	pp := p.(*MediaPlaylist)
	CheckType(t, pp)
	is.Equal(listType, MEDIA) // must be media playlist
	// check parsed values
	is.Equal(pp.TargetDuration, uint(15)) // target duration must be 15
	is.True(pp.Closed)                    // closed (VOD) playlist but Close field = false")
	is.Equal(pp.SeqNo, uint64(0))         // sequence number must be 0

	segNames := []string{"20181231/0555e0c371ea801726b92512c331399d_00000000.ts",
		"20181231/0555e0c371ea801726b92512c331399d_00000001.ts",
		"20181231/0555e0c371ea801726b92512c331399d_00000002.ts",
		"20181231/0555e0c371ea801726b92512c331399d_00000003.ts"}
	if pp.Count() != uint(len(segNames)) {
		t.Errorf("Segments in playlist %d != %d", pp.Count(), len(segNames))
	}

	for idx, name := range segNames {
		if pp.Segments[idx].URI != name {
			t.Errorf("Segment name mismatch (%d/%d): %s != %s", idx, pp.Count(), pp.Segments[idx].Title, name)
		}
	}

	// The ProgramDateTime of the 1st segment should be: 2018-12-31T09:47:22+08:00
	st, _ := time.Parse(time.RFC3339, "2018-12-31T09:47:22+08:00")
	if !pp.Segments[0].ProgramDateTime.Equal(st) {
		t.Errorf("The program date time of the 1st segment should be: %v, actual value: %v",
			st, pp.Segments[0].ProgramDateTime)
	}
}

func TestDecodeMediaPlaylistStartTime(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/media-playlist-with-start-time.m3u8")
	is.NoErr(err) // must open file
	p, listType, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err) // must decode playlist
	pp := p.(*MediaPlaylist)
	CheckType(t, pp)
	is.Equal(listType, MEDIA)            // must be media playlist
	is.Equal(pp.StartTime, float64(8.0)) // start time must be 8.0
}

func TestDecodeMediaPlaylistWithCueOutCueIn(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/media-playlist-with-cue-out-in-without-oatcls.m3u8")
	is.NoErr(err) // must open file
	p, listType, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err) // must decode playlist
	pp := p.(*MediaPlaylist)
	CheckType(t, pp)
	is.Equal(listType, MEDIA) // must be media playlist)

	is.Equal(pp.Segments[5].SCTE.CueType, SCTE35Cue_Start)  // EXT-CUE-OUT must result in SCTE35Cue_Start
	is.Equal(pp.Segments[5].SCTE.Time, float64(0))          // EXT-CUE-OUT without duration must not have Time set
	is.Equal(pp.Segments[9].SCTE.CueType, SCTE35Cue_End)    // EXT-CUE-IN must result in SCTE35Cue_End
	is.Equal(pp.Segments[30].SCTE.CueType, SCTE35Cue_Start) // EXT-CUE-OUT must result in SCTE35Cue_Start
	is.Equal(pp.Segments[30].SCTE.Time, float64(180))       // EXT-CUE-OUT:180.0 must have time set to 180
	is.Equal(pp.Segments[60].SCTE.CueType, SCTE35Cue_End)   // EXT-CUE-IN must result in SCTE35Cue_End
}

func TestDecodeMasterChannels(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-with-channels.m3u8")
	is.NoErr(err) // must open file
	p, listType, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err)              // must decode playlist
	is.Equal(listType, MASTER) // must be master playlist
	pp := p.(*MasterPlaylist)

	alt0 := pp.Variants[0].Alternatives[0]
	is.Equal(alt0.Type, "AUDIO") // Expected AUDIO track in test input Alternatives[0]
	is.Equal(alt0.Channels, "2") // Expected 2 channels track in test input Alternatives[0]

	alt1 := pp.Variants[1].Alternatives[0]
	is.Equal(alt1.Type, "AUDIO") // Expected AUDIO track in test input Alternatives[1]
	is.Equal(alt1.Channels, "6") // Expected 6 channels track in test input Alternatives[1]
}

func TestDecodeRenditionsAndIframes(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-groups-and-iframe.m3u8")
	is.NoErr(err) // must open file
	p, listType, err := DecodeFrom(bufio.NewReader(f), true)
	is.NoErr(err)              // must decode playlist
	is.Equal(listType, MASTER) // must be master playlist
	pp := p.(*MasterPlaylist)

	for _, v := range pp.Variants {
		switch v.Iframe {
		case true:
			is.Equal(len(v.Alternatives), 0) // Expected no alternatives in I-frame varian
		case false:
			is.True(len(v.Alternatives) > 0) // Expected at least one alternative in each video variant
		}
	}
	allRenditions := pp.GetAllAlternatives()
	is.Equal(len(allRenditions), 2) // Expected 2 renditions
}

// Check that dangling SCTE35 DateRange tags is reported as error since not supported.
func TestDanglingScte35DateRange(t *testing.T) {
	is := is.New(t)
	f1, err := os.Open("testdata/bad-playlists/media-playlist-dangling-scte35-daterange.m3u8")
	is.NoErr(err) // must open file
	defer f1.Close()
	_, listType, err := DecodeFrom(bufio.NewReader(f1), true)
	is.Equal(listType, MEDIA)                 // must be media playlist
	is.Equal(ErrDanglingSCTE35DateRange, err) // must return ErrDanglingSCTE35DateRange

	f2, err := os.Open("testdata/bad-playlists/media-playlist-dangling-scte35-daterange.m3u8")
	is.NoErr(err) // must open file
	defer f2.Close()
	pl, err := NewMediaPlaylist(0, 5)
	is.NoErr(err) // must create playlist
	err = pl.DecodeFrom(bufio.NewReader(f2), true)
	is.Equal(ErrDanglingSCTE35DateRange, err) // must return ErrDanglingSCTE35DateRange
}

func TestDecodeMasterPlaylistWithDefines(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-with-defines.m3u8")
	is.NoErr(err) // must open file
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err) // must decode playlist
	// check parsed values
	is.Equal(len(p.Defines), 2) // must be 2 defines

	is.Equal(p.Defines[0].Name, "Define1")
	is.Equal(p.Defines[0].Type, VALUE)
	is.Equal(p.Defines[0].Value, "Value1")
	is.Equal(p.Defines[1].Name, "Define2")
	is.Equal(p.Defines[1].Type, QUERYPARAM)
	is.Equal(p.Defines[1].Value, "")
}

func TestDecodeMasterPlaylistWithInvalidDefines(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/master-with-defines-invalid.m3u8")
	is.NoErr(err) // must open file
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err) // must decode playlist
	// check parsed values
	is.Equal(len(p.Defines), 0) // must be 0 defines
}

func TestDecodeMediaPlaylistWithDefines(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-playlists/media-playlist-with-queryparam.m3u8")
	is.NoErr(err) // must open file
	p, err := NewMediaPlaylist(1, 1)
	is.NoErr(err) // must create playlist
	err = p.DecodeFrom(bufio.NewReader(f), false)
	is.NoErr(err) // must decode playlist
	// check parsed values
	is.Equal(len(p.Defines), 3) // must be 3 defines

	is.Equal(p.Defines[0].Name, "Define1")
	is.Equal(p.Defines[0].Type, VALUE)
	is.Equal(p.Defines[0].Value, "Value1")
	is.Equal(p.Defines[1].Name, "Define2")
	is.Equal(p.Defines[1].Type, QUERYPARAM)
	is.Equal(p.Defines[1].Value, "")
	is.Equal(p.Defines[2].Name, "Define3")
	is.Equal(p.Defines[2].Type, IMPORT)
	is.Equal(p.Defines[2].Value, "")
}

func TestDecodeMediaPlaylistWithGaps(t *testing.T) {
	data := []struct {
		playlist string
		wantErr  bool
	}{
		{"sample-playlists/media-playlist-with-gap.m3u8", false},
		{"sample-playlists/bad-media-playlist-with-gap.m3u8", true},
	}
	for _, d := range data {
		t.Run(d.playlist, func(t *testing.T) {
			is := is.New(t)
			f, err := os.Open(d.playlist)
			is.NoErr(err) // must open file
			p, err := NewMediaPlaylist(0, 1)
			is.NoErr(err) // must create playlist
			err = p.DecodeFrom(bufio.NewReader(f), true)
			if d.wantErr {
				is.True(err != nil) // must return an error
			} else {
				is.NoErr(err)              // must decode playlist
				is.True(p.Segments[0].Gap) // First segment should signal gap
				is.True(p.Segments[4].Gap) // Fifth segment should signal gap
			}
		})
	}
}

func TestDeQuote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"quoted"`, "quoted"},
		{`"quoted with spaces"`, "quoted with spaces"},
		{`"quoted with \"escaped\" quotes"`, `quoted with \"escaped\" quotes`},
		{`"quoted with 'single' quotes"`, `quoted with 'single' quotes`},
		{`"quoted with \n new line"`, `quoted with \n new line`},
		{"not quoted", "not quoted"},
		{"", ""},
		{`"`, `"`},
		{`"a`, `"a`},
		{`a"`, `a"`},
	}

	for _, test := range tests {
		result := deQuote(test.input)
		if result != test.expected {
			t.Errorf("DeQuote(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestParsePartialSegment(t *testing.T) {
	tests := []struct {
		name       string
		parameters string
		want       *PartialSegment
		wantErr    bool
	}{
		{
			name:       "Valid parameters",
			parameters: `URI="segment.ts",DURATION=10.0,INDEPENDENT=YES,BYTERANGE=1000@2000`,
			want: &PartialSegment{
				URI:         "segment.ts",
				Duration:    10.0,
				Independent: true,
				Limit:       1000,
				Offset:      2000,
			},
			wantErr: false,
		},
		{
			name:       "Invalid duration",
			parameters: `URI="segment.ts",DURATION=invalid,INDEPENDENT=YES,BYTERANGE=1000@2000`,
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "Invalid byterange",
			parameters: `URI="segment.ts",DURATION=10.0,INDEPENDENT=YES,BYTERANGE=invalid`,
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "Missing URI",
			parameters: `DURATION=10.0,INDEPENDENT=YES,BYTERANGE=1000@2000`,
			want: &PartialSegment{
				URI:         "",
				Duration:    10.0,
				Independent: true,
				Limit:       1000,
				Offset:      2000,
			},
			wantErr: false,
		},
		{
			name:       "Missing duration",
			parameters: `URI="segment.ts",INDEPENDENT=YES,BYTERANGE=1000@2000`,
			want: &PartialSegment{
				URI:         "segment.ts",
				Duration:    0,
				Independent: true,
				Limit:       1000,
				Offset:      2000,
			},
			wantErr: false,
		},
		{
			name:       "Missing byterange",
			parameters: `URI="segment.ts",DURATION=10.0,INDEPENDENT=YES`,
			want: &PartialSegment{
				URI:         "segment.ts",
				Duration:    10.0,
				Independent: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePartialSegment(tt.parameters)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePartialSegment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePartialSegment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePreloadHint(t *testing.T) {
	tests := []struct {
		name       string
		parameters string
		want       *PreloadHint
		wantErr    bool
	}{
		{
			name:       "Valid parameters",
			parameters: `TYPE=PART,URI="filePart273.4.m4s",BYTERANGE-START=1234,BYTERANGE-LENGTH=5678`,
			want: &PreloadHint{
				Type:   "PART",
				URI:    "filePart273.4.m4s",
				Offset: 1234,
				Limit:  5678,
			},
			wantErr: false,
		},
		{
			name:       "Invalid BYTERANGE-START",
			parameters: `TYPE=PART,URI="filePart273.4.m4s",BYTERANGE-START=invalid,BYTERANGE-LENGTH=5678`,
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "Invalid BYTERANGE-LENGTH",
			parameters: `TYPE=PART,URI="filePart273.4.m4s",BYTERANGE-START=1234,BYTERANGE-LENGTH=invalid`,
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "Missing URI",
			parameters: `TYPE=PART,BYTERANGE-START=1234,BYTERANGE-LENGTH=5678`,
			want: &PreloadHint{
				Type:   "PART",
				URI:    "",
				Offset: 1234,
				Limit:  5678,
			},
			wantErr: false,
		},
		{
			name:       "Missing BYTERANGE-START",
			parameters: `TYPE=PART,URI="filePart273.4.m4s",BYTERANGE-LENGTH=5678`,
			want: &PreloadHint{
				Type:   "PART",
				URI:    "filePart273.4.m4s",
				Offset: 0,
				Limit:  5678,
			},
			wantErr: false,
		},
		{
			name:       "Missing BYTERANGE-LENGTH",
			parameters: `TYPE=PART,URI="filePart273.4.m4s",BYTERANGE-START=1234`,
			want: &PreloadHint{
				Type:   "PART",
				URI:    "filePart273.4.m4s",
				Offset: 1234,
				Limit:  0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePreloadHint(tt.parameters)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePreloadHint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePreloadHint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseServerControl(t *testing.T) {
	tests := []struct {
		name       string
		parameters string
		want       *ServerControl
		wantErr    bool
	}{
		{
			name:       "Valid parameters",
			parameters: `CAN-SKIP-UNTIL=12.5,CAN-SKIP-DATERANGES=YES,HOLD-BACK=10.0,PART-HOLD-BACK=5.0,CAN-BLOCK-RELOAD=YES`,
			want: &ServerControl{
				CanSkipUntil:      12.5,
				CanSkipDateRanges: true,
				HoldBack:          10.0,
				PartHoldBack:      5.0,
				CanBlockReload:    true,
			},
			wantErr: false,
		},
		{
			name:       "Invalid CAN-SKIP-UNTIL",
			parameters: `CAN-SKIP-UNTIL=invalid,CAN-SKIP-DATERANGES=YES,HOLD-BACK=10.0,PART-HOLD-BACK=5.0,CAN-BLOCK-RELOAD=YES`,
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "Invalid HOLD-BACK",
			parameters: `CAN-SKIP-UNTIL=12.5,CAN-SKIP-DATERANGES=YES,HOLD-BACK=invalid,PART-HOLD-BACK=5.0,CAN-BLOCK-RELOAD=YES`,
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "Invalid PART-HOLD-BACK",
			parameters: `CAN-SKIP-UNTIL=12.5,CAN-SKIP-DATERANGES=YES,HOLD-BACK=10.0,PART-HOLD-BACK=invalid,CAN-BLOCK-RELOAD=YES`,
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "Missing CAN-SKIP-UNTIL",
			parameters: `CAN-SKIP-DATERANGES=YES,HOLD-BACK=10.0,PART-HOLD-BACK=5.0,CAN-BLOCK-RELOAD=YES`,
			want: &ServerControl{
				CanSkipUntil:      0,
				CanSkipDateRanges: true,
				HoldBack:          10.0,
				PartHoldBack:      5.0,
				CanBlockReload:    true,
			},
			wantErr: false,
		},
		{
			name:       "Missing CAN-SKIP-DATERANGES",
			parameters: `CAN-SKIP-UNTIL=12.5,HOLD-BACK=10.0,PART-HOLD-BACK=5.0,CAN-BLOCK-RELOAD=YES`,
			want: &ServerControl{
				CanSkipUntil:      12.5,
				CanSkipDateRanges: false,
				HoldBack:          10.0,
				PartHoldBack:      5.0,
				CanBlockReload:    true,
			},
			wantErr: false,
		},
		{
			name:       "Missing HOLD-BACK",
			parameters: `CAN-SKIP-UNTIL=12.5,CAN-SKIP-DATERANGES=YES,PART-HOLD-BACK=5.0,CAN-BLOCK-RELOAD=YES`,
			want: &ServerControl{
				CanSkipUntil:      12.5,
				CanSkipDateRanges: true,
				HoldBack:          0,
				PartHoldBack:      5.0,
				CanBlockReload:    true,
			},
			wantErr: false,
		},
		{
			name:       "Missing PART-HOLD-BACK",
			parameters: `CAN-SKIP-UNTIL=12.5,CAN-SKIP-DATERANGES=YES,HOLD-BACK=10.0,CAN-BLOCK-RELOAD=YES`,
			want: &ServerControl{
				CanSkipUntil:      12.5,
				CanSkipDateRanges: true,
				HoldBack:          10.0,
				PartHoldBack:      0,
				CanBlockReload:    true,
			},
			wantErr: false,
		},
		{
			name:       "Missing CAN-BLOCK-RELOAD",
			parameters: `CAN-SKIP-UNTIL=12.5,CAN-SKIP-DATERANGES=YES,HOLD-BACK=10.0,PART-HOLD-BACK=5.0`,
			want: &ServerControl{
				CanSkipUntil:      12.5,
				CanSkipDateRanges: true,
				HoldBack:          10.0,
				PartHoldBack:      5.0,
				CanBlockReload:    false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseServerControl(tt.parameters)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseServerControl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseServerControl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSkipTag(t *testing.T) {
	tests := []struct {
		name       string
		parameters string
		want       uint64
		wantErr    bool
	}{
		{
			name:       "Valid SKIPPED-SEGMENTS",
			parameters: `SKIPPED-SEGMENTS=5`,
			want:       5,
			wantErr:    false,
		},
		{
			name:       "Invalid SKIPPED-SEGMENTS",
			parameters: `SKIPPED-SEGMENTS=invalid`,
			want:       0,
			wantErr:    true,
		},
		{
			name:       "Missing SKIPPED-SEGMENTS",
			parameters: ``,
			want:       0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSkipTag(tt.parameters)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSkipTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseSkipTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

/***************************
 *  Code parsing examples  *
 ***************************/

func ExampleDecodeFrom_withDiscontinuityAndOutput() {
	f, _ := os.Open("sample-playlists/media-playlist-with-discontinuity.m3u8")
	p, _, _ := DecodeFrom(bufio.NewReader(f), true)
	pp := p.(*MediaPlaylist)
	fmt.Printf("%s", pp)
	// Output:
	// #EXTM3U
	// #EXT-X-VERSION:3
	// #EXT-X-MEDIA-SEQUENCE:0
	// #EXT-X-TARGETDURATION:10
	// #EXTINF:10.000,
	// ad0.ts
	// #EXTINF:8.000,
	// ad1.ts
	// #EXT-X-DISCONTINUITY
	// #EXTINF:10.000,
	// movieA.ts
	// #EXTINF:10.000,
	// movieB.ts
}

/****************
 *  Benchmarks  *
 ****************/

func BenchmarkDecodeMasterPlaylist(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f, err := os.Open("sample-playlists/master.m3u8")
		if err != nil {
			b.Fatal(err)
		}
		p := NewMasterPlaylist()
		if err := p.DecodeFrom(bufio.NewReader(f), false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeMediaPlaylist(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f, err := os.Open("sample-playlists/media-playlist-large.m3u8")
		if err != nil {
			b.Fatal(err)
		}
		p, err := NewMediaPlaylist(50000, 50000)
		if err != nil {
			b.Fatalf("Create media playlist failed: %s", err)
		}
		if err = p.DecodeFrom(bufio.NewReader(f), true); err != nil {
			b.Fatal(err)
		}
	}
}

func readTestMasterPlaylist(t *testing.T, fileName string) (*MasterPlaylist, error) {
	t.Helper()
	f, err := os.Open(fileName)
	if err != nil {
		t.Fail()
	}
	defer f.Close()
	p := NewMasterPlaylist()
	err = p.DecodeFrom(bufio.NewReader(f), false)
	if err != nil {
		t.Fail()
	}
	return p, nil
}

func readTestMediaPlaylist(t *testing.T, fileName string) (*MediaPlaylist, error) {
	t.Helper()
	f, err := os.Open(fileName)
	if err != nil {
		t.Fail()
	}
	defer f.Close()
	p, err := NewMediaPlaylist(50000, 50000)
	if err != nil {
		t.Fatalf("Create media playlist failed: %s", err)
	}
	err = p.DecodeFrom(bufio.NewReader(f), true)
	if err != nil {
		t.Fail()
	}
	return p, nil
}

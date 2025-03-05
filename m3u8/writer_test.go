package m3u8

/*
Playlist generation tests.
*/

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/matryer/is"
)

// Check how master and media playlists implement common Playlist interface
func TestInterfaceImplemented(t *testing.T) {
	is := is.New(t)
	m := NewMasterPlaylist()
	CheckType(t, m)
	p, e := NewMediaPlaylist(1, 2)
	is.NoErr(e) // create media playlist must be successful
	CheckType(t, p)
}

// Create new media playlist with wrong size (must be failed)
func TestCreateMediaPlaylistWithWrongSize(t *testing.T) {
	is := is.New(t)
	_, e := NewMediaPlaylist(2, 1) // wrong winsize
	is.True(e != nil)              // create media playlist with wrong winsize  must fail
}

// Tests the last method on media playlist
func TestLastSegmentMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, _ := NewMediaPlaylist(5, 5)
	is.Equal(p.last(), uint(4)) // last segment of empty playlist must be 4
	for i := uint(0); i < 5; i++ {
		_ = p.Append("uri.ts", 4, "")
		is.Equal(p.last(), i) // last segment must be equal to i
	}
}

// Create new media playlist
// Add two segments to media playlist
func TestAddSegmentToMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, e := NewMediaPlaylist(1, 2)
	is.NoErr(e) // Create media playlist should be successful
	e = p.Append("test01.ts", 10.0, "title")
	is.NoErr(e)                              // Add 1st segment to a media playlist should be successful
	is.Equal(p.Segments[0].URI, "test01.ts") // Check URI of the 1st segment
	is.Equal(p.Segments[0].Duration, 10.0)   // Check duration of the 1st segment
	is.Equal(p.Segments[0].Title, "title")   // Check title of the 1st segment
	is.Equal(p.Segments[0].SeqId, uint64(0)) // Check SeqId of the 1st segment
}

func TestAppendSegmentToMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, _ := NewMediaPlaylist(2, 2)
	e := p.AppendSegment(&MediaSegment{Duration: 10})
	is.NoErr(e)                          // Add 1st segment to a media playlist should be successful
	is.Equal(p.TargetDuration, uint(10)) // target duration should be set to 10
	e = p.AppendSegment(&MediaSegment{Duration: 10})
	is.NoErr(e) // Add 2nd segment to a media playlist should be successful
	e = p.AppendSegment(&MediaSegment{Duration: 10})
	is.True(e != nil)            // Add 3rd segment to a media playlist should fail
	is.Equal(p.Count(), uint(2)) // Count of segments should be 2, the capacity of the playlist
	if p.SeqNo != 0 || p.Segments[0].SeqId != 0 || p.Segments[1].SeqId != 1 {
		t.Errorf("Excepted SeqNo and SeqId: 0/0/1, got: %v/%v/%v", p.SeqNo, p.Segments[0].SeqId, p.Segments[1].SeqId)
	}
}

// Create new media playlist
// Add three segments to media playlist
// Set discontinuity tag for the 2nd segment.
func TestDiscontinuityForMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, e := NewMediaPlaylist(3, 4)
	is.NoErr(e) // Create media playlist should be successful
	p.Close()
	e = p.Append("test01.ts", 5.0, "")
	is.NoErr(e) // Add 1st segment to a media playlist should be successful
	e = p.Append("test02.ts", 6.0, "")
	is.NoErr(e) // Add 2nd segment to a media playlist should be successful
	e = p.SetDiscontinuity()
	is.NoErr(e) // Set discontinuity tag should be successful
	e = p.Append("test03.ts", 6.0, "")
	is.NoErr(e) // Add 3nd segment to a media playlist should be successful
	// fmt.Println(p.Encode().String())
}

// Create new media playlist
// Add three segments to media playlist
// Set program date and time for 2nd segment.
// Set discontinuity tag for the 2nd segment.
func TestProgramDateTimeForMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, e := NewMediaPlaylist(3, 4)
	is.NoErr(e) // Create media playlist should be successful
	p.Close()
	e = p.Append("test01.ts", 5.0, "")
	is.NoErr(e) // Add 1st segment to a media playlist should be successful
	e = p.Append("test02.ts", 6.0, "")
	is.NoErr(e) // Add 2nd segment to a media playlist should be successful
	loc, _ := time.LoadLocation("Europe/Moscow")
	e = p.SetProgramDateTime(time.Date(2010, time.November, 30, 16, 25, 0, 125*1e6, loc))
	is.NoErr(e) // setProgramDateTime should be successful
	e = p.SetDiscontinuity()
	is.NoErr(e) // Set discontinuity tag should be successful
	e = p.Append("test03.ts", 6.0, "")
	is.NoErr(e) // Add 3nd segment to a media playlist should be successful
	// fmt.Println(p.Encode().String())
}

// Create new media playlist with capacity 10 elements
// Try to add 11 segments to media playlist (oversize error)
func TestOverAddSegmentsToMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, e := NewMediaPlaylist(1, 10)
	is.NoErr(e) // Create media playlist with capacity 10 should be successful
	for i := 0; i < 10; i++ {
		e = p.Append(fmt.Sprintf("test%d.ts", i), 5.0, "")
		is.NoErr(e) // Add first 10 segments should be successful
	}
	e = p.Append(fmt.Sprintf("test%d.ts", 10), 5.0, "")
	is.True(e != nil) // Add 11th segment should fail
}

func TestSetSCTE35(t *testing.T) {
	p, _ := NewMediaPlaylist(1, 2)
	scte := &SCTE{Cue: "some cue"}
	if err := p.SetSCTE35(scte); err == nil {
		t.Error("SetSCTE35 expected empty playlist error")
	}
	_ = p.Append("test01.ts", 10.0, "title")
	if err := p.SetSCTE35(scte); err != nil {
		t.Errorf("SetSCTE35 did not expect error: %v", err)
	}
	if !reflect.DeepEqual(p.Segments[0].SCTE, scte) {
		t.Errorf("SetSCTE35\nexp: %#v\ngot: %#v", scte, p.Segments[0].SCTE)
	}
}

// Create new media playlist
// Don't add segments
// Expect error when trying to set EXT-X-GAP
func TestGap(t *testing.T) {
	p, _ := NewMediaPlaylist(1, 2)
	if err := p.SetGap(); err == nil {
		t.Error("SetGap expected empty playlist error")
	}
	_ = p.Append("test01.ts", 10.0, "title")
	if err := p.SetGap(); err != nil {
		t.Errorf("SetGap did not expect error: %v", err)
	}
	if !p.Segments[0].Gap {
		t.Error("SetGap did not set gap")
	}
}

// Create new media playlist
// Add segment to media playlist
// Set SCTE
func TestSetSCTEForMediaPlaylist(t *testing.T) {
	is := is.New(t)
	tests := []struct {
		Cue      string
		ID       string
		Time     float64
		Expected string
	}{
		{"CueData1", "", 0, `#EXT-SCTE35:CUE="CueData1"` + "\n"},
		{"CueData2", "ID2", 0, `#EXT-SCTE35:CUE="CueData2",ID="ID2"` + "\n"},
		{"CueData3", "ID3", 3.141, `#EXT-SCTE35:CUE="CueData3",ID="ID3",TIME=3.141` + "\n"},
		{"CueData4", "", 3.1, `#EXT-SCTE35:CUE="CueData4",TIME=3.1` + "\n"},
		{"CueData5", "", 3.0, `#EXT-SCTE35:CUE="CueData5",TIME=3` + "\n"},
	}

	for _, test := range tests {
		p, e := NewMediaPlaylist(1, 1)
		is.NoErr(e) // Create media playlist should be successful
		e = p.Append("test01.ts", 5.0, "")
		is.NoErr(e) //  Add 1st segment to a media playlist should be successful
		e = p.SetSCTE(test.Cue, test.ID, test.Time)
		is.NoErr(e)                                          // Set SCTE to a media playlist should be successful
		is.True(strings.Contains(p.String(), test.Expected)) // Check SCTE in a media playlist
	}
}

// Create new media playlist
// Add segment to media playlist
// Set encryption key
func TestSetKeyForMediaPlaylist(t *testing.T) {
	is := is.New(t)
	tests := []struct {
		KeyFormat         string
		KeyFormatVersions string
		ExpectVersion     uint8
	}{
		{"", "", 3},
		{"Format", "", 5},
		{"", "Version", 5},
		{"Format", "Version", 5},
	}

	for _, test := range tests {
		p, e := NewMediaPlaylist(3, 5)
		is.NoErr(e) // Create media playlist should be successful
		e = p.Append("test01.ts", 5.0, "")
		is.NoErr(e) // Add 1st segment to a media playlist should be successful
		e = p.SetKey("AES-128", "https://example.com", "iv", test.KeyFormat, test.KeyFormatVersions)
		is.NoErr(e)                         // Set key to a media playlist should be successful
		is.Equal(p.ver, test.ExpectVersion) // Check key playlist version
	}
}

// Create new media playlist
// Add segment to media playlist
// Set encryption key
func TestSetDefaultKeyForMediaPlaylist(t *testing.T) {
	is := is.New(t)
	tests := []struct {
		KeyFormat         string
		KeyFormatVersions string
		ExpectVersion     uint8
	}{
		{"", "", 3},
		{"Format", "", 5},
		{"", "Version", 5},
		{"Format", "Version", 5},
	}

	for _, test := range tests {
		p, e := NewMediaPlaylist(3, 5)
		is.NoErr(e) // Create media playlist should be successful
		e = p.SetDefaultKey("AES-128", "https://example.com", "iv", test.KeyFormat,
			test.KeyFormatVersions)
		is.NoErr(e)                         // Set key to a media playlist should be successful
		is.Equal(p.ver, test.ExpectVersion) // Check key playlist version
	}
}

// Create new media playlist
// Set default map with byte range
func TestSetDefaultMapForMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, e := NewMediaPlaylist(3, 5)
	is.NoErr(e) // Create media playlist should be successful
	p.SetDefaultMap("https://example.com", 1000*1024, 1024*1024)

	expected := `EXT-X-MAP:URI="https://example.com",BYTERANGE=1024000@1048576`
	is.True(strings.Contains(p.String(), expected)) // map is not included in the playlist
}

// Create new media playlist
// Add segment to media playlist
// Set map on segment
func TestSetMapForMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, e := NewMediaPlaylist(3, 5)
	is.NoErr(e) // Create media playlist should be successful
	e = p.Append("test01.ts", 5.0, "")
	is.NoErr(e) // Add 1st segment to a media playlist should be successful
	e = p.SetMap("https://example.com", 1000*1024, 1024*1024)
	is.NoErr(e) // Set map to a media playlist should be successful

	expected := `EXT-X-MAP:URI="https://example.com",BYTERANGE=1024000@1048576
#EXTINF:5.000,
test01.ts`
	is.True(strings.Contains(p.String(), expected)) // map is included in the playlist with segment
}

// Create new media playlist
// Set default map
// Add segment to media playlist with two different maps.
// Only the second should be included in the playlist.
func TestEncodeMediaPlaylistWithDefaultMap(t *testing.T) {
	is := is.New(t)
	p, err := NewMediaPlaylist(3, 5)
	is.NoErr(err) // Create media playlist should be successful
	p.SetDefaultMap("https://example.com", 1000*1024, 1024*1024)

	err = p.Append("test01.ts", 5.0, "")
	is.NoErr(err) // Add 1st segment to a media playlist should be successful
	err = p.SetMap("https://example.com", 1000*1024, 1024*1024)
	is.NoErr(err) // Set map to a media playlist should be successful, but not set since same as default.

	err = p.Append("test02.ts", 5.0, "")
	is.NoErr(err) // Add 1st segment to a media playlist should be successful
	err = p.SetMap("https://example2.com", 1000*1024, 1024*1024)
	is.NoErr(err) // Set map to a media playlist should be successful, but not set since same as already set.

	err = p.SetDiscontinuity()
	is.NoErr(err) // Set discontinuity tag should be successful

	err = p.Append("test03.ts", 5.0, "")
	is.NoErr(err) // Add 1st segment to a media playlist should be successful
	err = p.SetMap("https://example2.com", 1000*1024, 1024*1024)
	is.NoErr(err) // Set map to a media playlist should be successful, but not set since same as already set.

	encoded := p.String()
	expected := `EXT-X-MAP:URI="https://example.com",BYTERANGE=1024000@1048576`
	is.Equal(1, strings.Count(encoded, expected)) // default map is included in the playlist just once

	expected = `EXT-X-MAP:URI="https://example2.com",BYTERANGE=1024000@1048576`
	is.Equal(1, strings.Count(encoded, expected)) // new map is included in the playlist just once

	split := strings.Split(encoded, "#EXT-X-DISCONTINUITY")
	is.Equal(2, len(split)) // discontinuity tag is included one time

	expected = `EXT-X-MAP:URI="https://example2.com",BYTERANGE=1024000@1048576`
	is.Equal(1, strings.Count(split[1], expected)) // new map should be after discontinuity tag
}

// Create new media playlist
// Add custom playlist tag
// Add segment with custom tag
func TestEncodeMediaPlaylistWithCustomTags(t *testing.T) {
	is := is.New(t)
	p, e := NewMediaPlaylist(1, 1)
	is.NoErr(e) // Create media playlist should be successful

	customPTag := &MockCustomTag{
		name:          "#CustomPTag",
		encodedString: "#CustomPTag",
	}
	p.SetCustomTag(customPTag)

	customEmptyPTag := &MockCustomTag{
		name:          "#CustomEmptyPTag",
		encodedString: "",
	}
	p.SetCustomTag(customEmptyPTag)

	e = p.Append("test01.ts", 5.0, "")
	is.NoErr(e) // Add 1st segment should be successful

	customSTag := &MockCustomTag{
		name:          "#CustomSTag",
		encodedString: "#CustomSTag",
	}
	e = p.SetCustomSegmentTag(customSTag)
	is.NoErr(e) // Set CustomTag to segment should be successful

	customEmptySTag := &MockCustomTag{
		name:          "#CustomEmptySTag",
		encodedString: "",
	}
	e = p.SetCustomSegmentTag(customEmptySTag)
	is.NoErr(e) // Set CustomTag to segment should be successful

	encoded := p.String()
	expectedStrings := []string{"#CustomPTag", "#CustomSTag"}
	for _, expected := range expectedStrings {
		is.True(strings.Contains(encoded, expected)) // custom tags should be included in the playlist
	}
	unexpectedStrings := []string{"#CustomEmptyPTag", "#CustomEmptySTag"}
	for _, unexpected := range unexpectedStrings {
		is.True(!strings.Contains(encoded, unexpected)) // empty custom tags should not be included in the playlist
	}
}

// Create new media playlist
// Add two segments to media playlist
// Encode structures to HLS
func TestEncodeMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, e := NewMediaPlaylist(3, 5)
	is.NoErr(e) // Create media playlist should be successful
	e = p.Append("test01.ts", 5.0, "")
	is.NoErr(e) // Add 1st segment to a media playlist should be successful
	expected := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:5
#EXTINF:5.000,
test01.ts
`
	out := p.String()
	is.Equal(out, expected) // Encode media playlist does not match expected
}

func TestEncodeMediaPlaylistWithGaps(t *testing.T) {
	is := is.New(t)
	p, e := NewMediaPlaylist(3, 5)
	is.NoErr(e) // Create media playlist should be successful
	p.SetVersion(8)
	e = p.Append("test01.ts", 5.0, "")
	is.NoErr(e) // Add 1st segment to a media playlist should be successful
	e = p.SetGap()
	is.NoErr(e) // Set gap tag should be successful
	expected := `#EXTM3U
#EXT-X-VERSION:8
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:5
#EXT-X-GAP
#EXTINF:5.000,
test01.ts
`
	out := p.String()
	is.Equal(out, expected) // Encode media playlist does not match expected
}

func TestEncodeLowLatencyMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, e := NewMediaPlaylist(5, 10)
	is.NoErr(e)                  // Create media playlist should be successful
	p.PartTargetDuration = 1.002 // Set tag #EXT-X-PART-INF:PART-TARGET

	e = p.Append("test00.m4s", 4.0, "")
	is.NoErr(e) // Add segment to a media playlist should be successful

	segments := [][]string{
		{"test01.1.m4s", "test01.2.m4s", "test01.3.m4s", "test01.4.m4s", "test01.m4s"},
		{"test02.1.m4s", "test02.2.m4s", "test02.3.m4s", "test02.4.m4s", "test02.m4s"},
		{"test03.1.m4s", "test03.2.m4s", "test03.3.m4s", "test03.4.m4s", "test03.m4s"},
		{"test04.1.m4s", "test04.2.m4s", "test04.3.m4s", "test04.4.m4s", "test04.m4s"},
		{"test05.1.m4s"}}
	for _, psList := range segments {
		for index, ps := range psList {
			if index > 0 && index == len(psList)-1 {
				e = p.Append(ps, 4.0, "")
				is.NoErr(e) // Add segment to a media playlist should be successful
			} else {
				e = p.AppendPartial(ps, 1.0, true)
				is.NoErr(e) // Add partial segment to a media playlist should be successful
			}
		}
	}

	for seqNo, psList := range segments {
		for partIndex := range psList {
			if partIndex > 0 && partIndex == len(psList)-1 {
				// ignore full segment
				continue
			}
			if seqNo == 0 {
				// ignore partial segment of first segment (they are removed)
				continue
			}

			partialSegment := p.PartialSegments[(seqNo-1)*4+partIndex]

			is.Equal(partialSegment.URI, psList[partIndex]) // Partial segment URI does not match expected
			is.Equal(partialSegment.SeqID, uint64(seqNo+1)) // Partial segment SeqID does not match expected
		}
	}

	p.SetPreloadHint("PART", "test05.2.m4s")

	partTargetDuration := p.PartTargetDuration
	serverControl := ServerControl{0.0, false, 0.0, partTargetDuration * 3, true}
	e = p.SetServerControl(&serverControl)
	is.NoErr(e) // Set server control should be successful

	// Output only partial segments from last 3 full segments
	expected := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-SERVER-CONTROL:PART-HOLD-BACK=3.006,CAN-BLOCK-RELOAD=YES
#EXT-X-PART-INF:PART-TARGET=1.002
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:4
#EXTINF:4.000,
test00.m4s
#EXTINF:4.000,
test01.m4s
#EXT-X-PART:DURATION=1.000,INDEPENDENT=YES,URI="test02.1.m4s"
#EXT-X-PART:DURATION=1.000,INDEPENDENT=YES,URI="test02.2.m4s"
#EXT-X-PART:DURATION=1.000,INDEPENDENT=YES,URI="test02.3.m4s"
#EXT-X-PART:DURATION=1.000,INDEPENDENT=YES,URI="test02.4.m4s"
#EXTINF:4.000,
test02.m4s
#EXT-X-PART:DURATION=1.000,INDEPENDENT=YES,URI="test03.1.m4s"
#EXT-X-PART:DURATION=1.000,INDEPENDENT=YES,URI="test03.2.m4s"
#EXT-X-PART:DURATION=1.000,INDEPENDENT=YES,URI="test03.3.m4s"
#EXT-X-PART:DURATION=1.000,INDEPENDENT=YES,URI="test03.4.m4s"
#EXTINF:4.000,
test03.m4s
#EXT-X-PART:DURATION=1.000,INDEPENDENT=YES,URI="test04.1.m4s"
#EXT-X-PART:DURATION=1.000,INDEPENDENT=YES,URI="test04.2.m4s"
#EXT-X-PART:DURATION=1.000,INDEPENDENT=YES,URI="test04.3.m4s"
#EXT-X-PART:DURATION=1.000,INDEPENDENT=YES,URI="test04.4.m4s"
#EXTINF:4.000,
test04.m4s
#EXT-X-PART:DURATION=1.000,INDEPENDENT=YES,URI="test05.1.m4s"
#EXT-X-PRELOAD-HINT:TYPE=PART,URI="test05.2.m4s"
`
	out := p.String()
	is.Equal(out, expected) // Encode media playlist does not match expected
}

func TestEncodeMediaPlaylistWithSkipUntil(t *testing.T) {
	is := is.New(t)
	p, e := NewMediaPlaylist(10, 10)
	is.NoErr(e) // Create media playlist should be successful

	p.SetVersion(9)                   // Version 9 is required for EXT-X-SKIP tag
	p.SetDefaultMap("init.mp4", 0, 0) // Set init segment (will be ignored)

	for i := 0; i < 10; i++ {
		e = p.Append(fmt.Sprintf("test%02d.m4s", i), 4.0, "")
		is.NoErr(e) // Add segment to a media playlist should be successful
	}

	skipped := uint64(6)
	canSkipUntil := float64(4.0 * skipped) // skip 6 segment
	holdBack := 4.0 * 3                    // hold back 3 segment
	serverControl := ServerControl{canSkipUntil, false, holdBack, 0.0, true}
	e = p.SetServerControl(&serverControl)
	is.NoErr(e) // Set server control should be successful

	expected := `#EXTM3U
#EXT-X-VERSION:9
#EXT-X-SERVER-CONTROL:CAN-SKIP-UNTIL=24.000,HOLD-BACK=12.000,CAN-BLOCK-RELOAD=YES
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-TARGETDURATION:4
#EXT-X-SKIP:SKIPPED-SEGMENTS=6
#EXTINF:4.000,
test06.m4s
#EXTINF:4.000,
test07.m4s
#EXTINF:4.000,
test08.m4s
#EXTINF:4.000,
test09.m4s
`
	out, err := p.EncodeWithSkip(skipped)
	is.NoErr(err)                    // Encode with skipped should be successful
	is.Equal(out.String(), expected) // Encode media playlist does not match expected
}

// Create new media playlist
// Add 10 segments to media playlist
// Test iterating over segments
func TestLoopSegmentsOfMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, e := NewMediaPlaylist(3, 5)
	is.NoErr(e) // Create media playlist should be successful
	for i := 0; i < 5; i++ {
		e = p.Append(fmt.Sprintf("test%d.ts", i), 5.0, "")
		is.NoErr(e) // Add segment to a media playlist should be successful
	}
	out := p.String()
	is.Equal(strings.Count(out, `#EXTINF:5.000,`), 3) // EXTINF not set to 5 on all segments
}

// Create new media playlist with capacity 5
// Add 5 segments and 5 unique keys
// Test correct keys set on correct segments
func TestEncryptionKeysInMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, _ := NewMediaPlaylist(5, 5)
	// Add 5 segments and set custom encryption key
	for i := uint(0); i < 5; i++ {
		uri := fmt.Sprintf("uri-%d", i)
		expected := &Key{
			Method:            "AES-128",
			URI:               uri,
			IV:                fmt.Sprintf("%d", i),
			Keyformat:         "identity",
			Keyformatversions: "1",
		}
		_ = p.Append(uri+".ts", 4, "")
		_ = p.SetKey(expected.Method, expected.URI, expected.IV, expected.Keyformat, expected.Keyformatversions)

		is.True(p.Segments[i].Key != nil)     // Key was not set on segment
		is.Equal(p.Segments[i].Key, expected) // Key does not match expected
	}
}

func TestEncryptionKeyMethodNoneInMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, e := NewMediaPlaylist(5, 5)
	is.NoErr(e) // Create media playlist should be successful
	_ = p.Append("segment-1.ts", 4, "")
	_ = p.SetKey("AES-128", "key-uri", "iv", "identity", "1")
	_ = p.Append("segment-2.ts", 4, "")
	_ = p.SetKey("NONE", "", "", "", "")
	expected := `#EXT-X-KEY:METHOD=NONE
#EXTINF:4.000,
segment-2.ts`
	is.True(strings.Contains(p.String(), expected)) // Key method NONE is not included in the playlist
}

// Create new media playlist
// Add 10 segments to media playlist
// Encode structure to HLS with integer target durations
func TestMediaPlaylistWithIntegerDurations(t *testing.T) {
	is := is.New(t)
	p, _ := NewMediaPlaylist(3, 10)
	for i := 0; i < 9; i++ {
		e := p.Append(fmt.Sprintf("test%d.ts", i), 5.6, "")
		is.NoErr(e) // Add segment to a media playlist should be successful
	}
	out := p.String()
	is.Equal(strings.Count(out, `#EXTINF:5.600,`), 3) // EXTINF not set to 5.600 on all segments
}

// Create new media playlist
// Add 9 segments to media playlist
// 11 times encode structure to HLS with integer target durations
// Last playlist must be empty
func TestMediaPlaylistWithEmptyMedia(t *testing.T) {
	is := is.New(t)
	p, _ := NewMediaPlaylist(3, 10)
	for i := 1; i < 10; i++ {
		err := p.Append(fmt.Sprintf("test%d.ts", i), 5.6, "")
		is.NoErr(err) // Add segment to a media playlist should be successful
	}
	for i := 1; i < 10; i++ {
		// fmt.Println(p.Encode().String())
		err := p.Remove()
		is.NoErr(err) // Remove segment from a media playlist should be successful
	}
	err := p.Remove()
	is.True(err != nil) // Remove segment from an empty media playlist should fail
	// TODO add check for buffers equality
}

// Create new media playlist with winsize == capacity
func TestMediaPlaylistWinsize(t *testing.T) {
	is := is.New(t)
	p, _ := NewMediaPlaylist(6, 6)
	for i := 1; i < 10; i++ {
		p.Slide(fmt.Sprintf("test%d.ts", i), 5.6, "")
	}
	is.Equal(p.Count(), uint(6)) // Count of segments does not match expected 6
	is.Equal(p.SeqNo, uint64(3)) // SeqNo of media playlist does not match expected 3
}

// Create new media playlist as sliding playlist.
// Close it.
func TestClosedMediaPlaylist(t *testing.T) {
	is := is.New(t)
	p, _ := NewMediaPlaylist(1, 10)
	for i := 0; i < 10; i++ {
		e := p.Append(fmt.Sprintf("test%d.ts", i), 5.0, "")
		is.NoErr(e) // Add 10 segments to capacity 10 list should be successful
	}
	p.Close()
}

// Create new media playlist as sliding playlist.
func TestLargeMediaPlaylistWithParallel(t *testing.T) {
	is := is.New(t)
	testCount := 10
	expected, err := os.ReadFile("sample-playlists/media-playlist-large.m3u8")
	is.NoErr(err) // Read expected playlist should be successful
	// Fix potential CRLF issues on Windows
	expected = bytes.Replace(expected, []byte("\r\n"), []byte("\n"), -1)
	var wg sync.WaitGroup
	for i := 0; i < testCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f, err := os.Open("sample-playlists/media-playlist-large.m3u8")
			is.NoErr(err) // Open playlist file in parallel go routine should be successful
			p, err := NewMediaPlaylist(50000, 50000)
			is.NoErr(err) // Create media playlist in parallel go routine should be successful
			err = p.DecodeFrom(bufio.NewReader(f), true)
			is.NoErr(err) // Decode media playlist in parallel go routine should be successful

			actual := p.Encode().Bytes() // disregard output
			// Fix potential CRLF issues on Windows
			actual = bytes.Replace(actual, []byte("\r\n"), []byte("\n"), -1)
			is.Equal(expected, actual) // Expected playlist does not match actual
		}()
		wg.Wait()
	}
}

func TestMediaVersion(t *testing.T) {
	is := is.New(t)
	m, err := NewMediaPlaylist(3, 3)
	is.NoErr(err) // Create media playlist should be successful
	m.ver = 5
	is.Equal(m.Version(), uint8(5)) // Version does not match expected 5
}

func TestMediaSetVersion(t *testing.T) {
	is := is.New(t)
	m, _ := NewMediaPlaylist(3, 3)
	m.ver = 3
	is.Equal(m.Version(), uint8(3)) // Version does not match expected 3
	m.SetVersion(5)
	is.Equal(m.ver, uint8(5)) // Version does not match expected 5
}

func TestMediaWinSize(t *testing.T) {
	is := is.New(t)
	winSize := uint(3)
	m, err := NewMediaPlaylist(winSize, 3)
	is.NoErr(err)                  // Create media playlist should be successful
	is.Equal(m.WinSize(), winSize) // WinSize does not match expected 3
}

func TestMediaSetWinSize(t *testing.T) {
	is := is.New(t)
	m, _ := NewMediaPlaylist(3, 5)
	is.Equal(m.WinSize(), uint(3)) // WinSize does not match expected 3
	err := m.SetWinSize(5)
	is.NoErr(err)                  // Set winsize to 5 failed
	is.Equal(m.WinSize(), uint(5)) // WinSize does not match expected 5
	// Check winsize cannot exceed capacity
	err = m.SetWinSize(99999)
	is.True(err != nil) // Set winsize to 99999 did not fail
	// Ensure winsize didn't change
	is.Equal(m.WinSize(), uint(5)) // WinSize did not stay 5
}

func TestIndependentSegments(t *testing.T) {
	is := is.New(t)
	m := NewMasterPlaylist()
	is.True(!m.IndependentSegments()) // independent segments not false by default
	m.SetIndependentSegments(true)
	is.True(m.IndependentSegments())                                     // independent segments not set  to true
	is.True(strings.Contains(m.String(), "#EXT-X-INDEPENDENT-SEGMENTS")) // independent segments not in playlist
}

// Create new media playlist
// Set default map
func TestStartTimeOffset(t *testing.T) {
	is := is.New(t)
	p, _ := NewMediaPlaylist(3, 5)
	p.StartTime = 3.4

	expected := `#EXT-X-START:TIME-OFFSET=3.4`
	is.True(strings.Contains(p.String(), expected)) // start time offset is not included in the playlist
}

func TestMediaPlaylist_Slide(t *testing.T) {
	is := is.New(t)
	m, e := NewMediaPlaylist(3, 4)
	is.NoErr(e) // NewMediaPlaylist failed

	_ = m.Append("t00.ts", 10, "")
	_ = m.Append("t01.ts", 10, "")
	_ = m.Append("t02.ts", 10, "")
	_ = m.Append("t03.ts", 10, "")
	is.Equal(m.Count(), uint(4)) // Count of segments not 4
	is.Equal(m.SeqNo, uint64(0)) // SeqNo of media playlist not 0
	var seqId, idx uint
	for idx, seqId = 0, 0; idx < 3; idx, seqId = idx+1, seqId+1 {
		segIdx := (m.head + idx) % m.capacity
		segUri := fmt.Sprintf("t%02d.ts", seqId)
		seg := m.Segments[segIdx]
		if seg.URI != segUri || seg.SeqId != uint64(seqId) {
			t.Errorf("Excepted segment: %s with SeqId: %v, got: %v/%v", segUri, seqId, seg.URI, seg.SeqId)
		}
	}

	m.Slide("t04.ts", 10, "")
	is.Equal(m.Count(), uint(4)) // Count of segments changed from 4
	is.Equal(m.SeqNo, uint64(1)) // SeqNo of media playlist not changed to 1
	for idx, seqId = 0, 1; idx < 3; idx, seqId = idx+1, seqId+1 {
		segIdx := (m.head + idx) % m.capacity
		segUri := fmt.Sprintf("t%02d.ts", seqId)
		seg := m.Segments[segIdx]
		if seg.URI != segUri || seg.SeqId != uint64(seqId) {
			t.Errorf("Excepted segment: %s with SeqId: %v, got: %v/%v", segUri, seqId, seg.URI, seg.SeqId)
		}
	}

	m.Slide("t05.ts", 10, "")
	m.Slide("t06.ts", 10, "")
	is.Equal(m.Count(), uint(4)) // Count of segments changed from 4
	is.Equal(m.SeqNo, uint64(3)) // SeqNo of media playlist not changed to 3
	for idx, seqId = 0, 3; idx < 3; idx, seqId = idx+1, seqId+1 {
		segIdx := (m.head + idx) % m.capacity
		segUri := fmt.Sprintf("t%02d.ts", seqId)
		seg := m.Segments[segIdx]
		if seg.URI != segUri || seg.SeqId != uint64(seqId) {
			t.Errorf("Excepted segment: %s with SeqId: %v, got: %v/%v", segUri, seqId, seg.URI, seg.SeqId)
		}
	}
}

// Create new master playlist without params
// Add media playlist
func TestNewMasterPlaylist(t *testing.T) {
	is := is.New(t)
	m := NewMasterPlaylist()
	p, _ := NewMediaPlaylist(3, 5)
	for i := 0; i < 5; i++ {
		e := p.Append(fmt.Sprintf("test%d.ts", i), 5.0, "")
		is.NoErr(e) // Add segment to a media playlist should be successful
	}
	m.Append("chunklist1.m3u8", p, VariantParams{})
}

// Create new master playlist without params
// Add media playlist with Alternatives
func TestNewMasterPlaylistWithAlternatives(t *testing.T) {
	is := is.New(t)
	m := NewMasterPlaylist()
	audioUri := fmt.Sprintf("%s/rendition.m3u8", "800")
	audioAlt := &Alternative{
		GroupId:    "audio",
		URI:        audioUri,
		Type:       "AUDIO",
		Name:       "main",
		Default:    true,
		Autoselect: true,
		Language:   "english",
	}
	p, _ := NewMediaPlaylist(3, 5)
	for i := 0; i < 5; i++ {
		e := p.Append(fmt.Sprintf("test%d.ts", i), 5.0, "")
		is.NoErr(e) // Add segment to a media playlist should be successful
	}
	m.Append("chunklist1.m3u8", p, VariantParams{Alternatives: []*Alternative{audioAlt}})

	is.Equal(m.Version(), uint8(4)) // Version does not match expected 4
	expected := `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="main",LANGUAGE="english",DEFAULT=YES,` +
		`AUTOSELECT=YES,URI="800/rendition.m3u8"`
	is.True(strings.Contains(m.String(), expected)) // Master playlist does not contain EXT-X-MEDIA
}

// Create new master playlist supporting CLOSED-CAPTIONS=NONE
func TestNewMasterPlaylistWithClosedCaptionEqNone(t *testing.T) {
	is := is.New(t)
	m := NewMasterPlaylist()

	vp := &VariantParams{
		Bandwidth:  8000,
		Codecs:     "avc1",
		Resolution: "1280x720",
		Audio:      "audio0",
		Captions:   "NONE",
	}

	p, _ := NewMediaPlaylist(1, 1)
	m.Append("eng_rendition_rendition.m3u8", p, *vp)

	expected := "CLOSED-CAPTIONS=NONE"
	is.True(strings.Contains(m.String(), expected)) // master playlist lacks CLOSED-CAPTIONS=NONE
	// quotes need to be include if not eq NONE
	vp.Captions = "CC1"
	m2 := NewMasterPlaylist()
	m2.Append("eng_rendition_rendition.m3u8", p, *vp)
	expected = `CLOSED-CAPTIONS="CC1"`
	is.True(strings.Contains(m2.String(), expected)) // master playlist lacks CLOSED-CAPTIONS="CC1"
}

// Create new master playlist with params
// Add media playlist
func TestNewMasterPlaylistWithParams(t *testing.T) {
	is := is.New(t)
	m := NewMasterPlaylist()
	p, _ := NewMediaPlaylist(3, 5)
	for i := 0; i < 5; i++ {
		e := p.Append(fmt.Sprintf("test%d.ts", i), 5.0, "")
		is.NoErr(e) // Add segment to a media playlist should be successful
	}
	m.Append("chunklist1.m3u8", p, VariantParams{Bandwidth: 1500000, Resolution: "576x480"})
	is.Equal(len(m.Variants), 1) // Number of variants does not match expected 1
}

// Create new master playlist
// Add media playlist with existing query params in URI
// Append more query params and ensure it encodes correctly
func TestEncodeMasterPlaylistWithExistingQuery(t *testing.T) {
	is := is.New(t)
	m := NewMasterPlaylist()
	p, _ := NewMediaPlaylist(3, 5)
	for i := 0; i < 5; i++ {
		e := p.Append(fmt.Sprintf("test%d.ts", i), 5.0, "")
		is.NoErr(e) // Add segment to a media playlist should be successful
	}
	m.Append("chunklist1.m3u8?k1=v1&k2=v2", p, VariantParams{Bandwidth: 1500000, Resolution: "576x480"})
	m.Args = "k3=v3"
	is.True(strings.Contains(m.String(), `chunklist1.m3u8?k1=v1&k2=v2&k3=v3`)) //
}

// Create new master playlist
// Add media playlist
// Encode structures to HLS
func TestEncodeMasterPlaylist(t *testing.T) {
	is := is.New(t)
	m := NewMasterPlaylist()
	p, _ := NewMediaPlaylist(3, 5)
	for i := 0; i < 5; i++ {
		e := p.Append(fmt.Sprintf("test%d.ts", i), 5.0, "")
		is.NoErr(e) // Add segment to a media playlist should be successful
	}
	m.Append("chunklist1.m3u8", p, VariantParams{Bandwidth: 1500000, Resolution: "576x480"})
	m.Append("chunklist2.m3u8", p, VariantParams{Bandwidth: 1500000, Resolution: "576x480"})
	nrVariants := len(m.Variants)
	is.Equal(nrVariants, 2) // Number of variants does not match expected 2
}

// Create new master playlist with Name tag in EXT-X-STREAM-INF
func TestEncodeMasterPlaylistWithStreamInfName(t *testing.T) {
	is := is.New(t)
	m := NewMasterPlaylist()
	p, _ := NewMediaPlaylist(3, 5)
	for i := 0; i < 5; i++ {
		e := p.Append(fmt.Sprintf("test%d.ts", i), 5.0, "")
		is.NoErr(e) // Add segment to a media playlist should be successful
	}
	m.Append("chunklist1.m3u8", p, VariantParams{Bandwidth: 3000, Resolution: "1152x960", Name: "HD 960p"})

	is.Equal(m.Variants[0].Name, "HD 960p")                 //  Bad variant name
	is.True(strings.Contains(m.String(), `NAME="HD 960p"`)) // Master playlist does not contain Name in EXT-X-STREAM-INF
}

func TestEncodeMasterPlaylistWithCustomTags(t *testing.T) {
	is := is.New(t)
	m := NewMasterPlaylist()
	customMTag := &MockCustomTag{
		name:          "#CustomMTag",
		encodedString: "#CustomMTag",
	}
	m.SetCustomTag(customMTag)

	encoded := m.String()
	expected := "#CustomMTag"

	is.True(strings.Contains(encoded, expected)) // Master playlist does not contain custom tag
}

func TestMasterVersion(t *testing.T) {
	is := is.New(t)
	m := NewMasterPlaylist()
	m.ver = 5
	is.Equal(m.Version(), uint8(5)) // Version does not match expected 5
	m.SetVersion(7)
	is.Equal(m.Version(), uint8(7)) // Version does not match expected 7
}

func TestKeyIsNotDuplicated(t *testing.T) {
	encoded := decodeEncode(t, "sample-playlists/media-playlist-with-key.m3u8")
	count := strings.Count(encoded, "#EXT-X-KEY")
	if count != 1 {
		t.Errorf("Expected number of EXT-X-KEY: 1 actual: %d", count)
	}
}

func decodeEncode(t *testing.T, fileName string) string {
	f, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	p, _, err := DecodeFrom(bufio.NewReader(f), true)
	if err != nil {
		t.Fatal(err)
	}
	pp := p.(*MediaPlaylist)
	return pp.Encode().String()
}

// TestCalculateTargetDuration tests the calculation of the target duration.
// It should be rounded up to an integer if the version is 5 or lower.
// If should be rounded to nearest integer if the version is 6 or higher.
// With nrSlides, we check that it works when the circular buffer has wrapped around.
// With lockedTargetDur, we check that it works when the target duration is locked.
func TestCalculateTargetDuration(t *testing.T) {
	is := is.New(t)
	cases := []struct {
		desc            string
		hlsVersion      uint8
		segDur          float64
		nrSlides        uint
		lockedTargetDur uint
		wantedTargetDur uint
	}{
		{desc: "HLSv5Locked", hlsVersion: 5, segDur: 5.1, nrSlides: 1, lockedTargetDur: 4, wantedTargetDur: 4},
		{desc: "HLSv5", hlsVersion: 5, segDur: 5.1, nrSlides: 2, wantedTargetDur: 6},
		{desc: "HLSv6", hlsVersion: 6, segDur: 5.1, nrSlides: 2, wantedTargetDur: 5},
		{desc: "HLSv5Wrap", hlsVersion: 5, segDur: 5.1, nrSlides: 6, wantedTargetDur: 6},
		{desc: "HLSv6Wrap", hlsVersion: 6, segDur: 5.1, nrSlides: 6, wantedTargetDur: 6},
		{desc: "Zero segments", hlsVersion: 5, segDur: 5.1, nrSlides: 0, wantedTargetDur: 0},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			p, err := NewMediaPlaylist(3, 5)
			if c.lockedTargetDur != 0 {
				p.SetTargetDuration(c.lockedTargetDur)
			}
			is.NoErr(err) // Create media playlist should be successful
			p.ver = c.hlsVersion
			for i := 0; i < int(c.nrSlides); i++ {
				segDur := c.segDur + 0.1*float64(i)
				p.Slide(fmt.Sprintf("test%d.ts", i), segDur, "")
			}
			is.Equal(p.TargetDuration, c.wantedTargetDur) // Target duration does not match expected
			if c.lockedTargetDur == 0 {
				calcTargetDur := p.CalculateTargetDuration(c.hlsVersion)
				is.Equal(calcTargetDur, c.wantedTargetDur) // Calculate target duration does not match expected
			}
		})
	}
}

// Create new master and media playlist
// Add define to playlists
func TestAppendDefine(t *testing.T) {
	is := is.New(t)
	tests := []struct {
		define   Define
		Expected string
	}{
		{Define{Name: "Define1", Type: VALUE, Value: "Value1"}, `#EXT-X-DEFINE:NAME="Define1",VALUE="Value1"` + "\n"},
		{Define{Name: "Define2", Type: IMPORT}, `#EXT-X-DEFINE:IMPORT="Define2"` + "\n"},
		{Define{Name: "Define3", Type: QUERYPARAM}, `#EXT-X-DEFINE:QUERYPARAM="Define3"` + "\n"},
	}

	for _, test := range tests {
		p := NewMasterPlaylist()
		e := p.AppendDefine(test.define)
		if test.define.Type != IMPORT {
			is.NoErr(e)
			is.True(strings.Contains(p.String(), test.Expected))
		}

		mp, e := NewMediaPlaylist(1, 1)
		is.NoErr(e) // Create media playlist should be successful
		mp.AppendDefine(test.define)
		is.True(strings.Contains(mp.String(), test.Expected))
	}
}

func TestIsPartOf(t *testing.T) {
	tests := []struct {
		partialSegUri string
		segUri        string
		expected      bool
	}{
		{"filePart249.1.m4s", "fileSequence249.m4s", true},
		{"filePart249.2.m4s", "fileSequence249.m4s", true},
		{"chunk249.1.m4s", "fileSequence249.m4s", true},
		{"filePart0249.1.m4s", "fileSequence249.m4s", true},

		{"filePart249.1.m4s", "fileSequence2490.m4s", false},
		{"filePart2490.1.m4s", "fileSequence249.m4s", false},
		{"filePart1249.1.m4s", "fileSequence249.m4s", false},
		{"filePart249.1.ts", "fileSequence249.m4s", false},
		{"filePart249.m4s", "fileSequence249.m4s", false},
		{"filePart249.m4s", "fileSequence249.1.m4s", false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%s", test.partialSegUri, test.segUri), func(t *testing.T) {
			result := IsPartOf(test.partialSegUri, test.segUri)
			if result != test.expected {
				t.Errorf("IsPartOf(%s, %s) = %v; want %v", test.partialSegUri, test.segUri, result, test.expected)
			}
		})
	}
}

/******************************
 *  Code generation examples  *
 ******************************/

// Create new media playlist
// Add two segments to media playlist
// Print it
func ExampleNewMediaPlaylist_string() {
	p, _ := NewMediaPlaylist(1, 2)
	_ = p.Append("test01.ts", 5.0, "")
	_ = p.Append("test02.ts", 6.0, "")
	fmt.Printf("%s\n", p)

	// Skip this for now as to be discussed in a separate PR
	// #EXTM3U
	// #EXT-X-VERSION:3
	// #EXT-X-MEDIA-SEQUENCE:0
	// #EXT-X-TARGETDURATION:6
	// #EXTINF:6.000,
	// test02.ts
}

// Create new media playlist
// Add two segments to media playlist
// Print it
func ExampleNewMediaPlaylist_stringWinsize0() {
	p, _ := NewMediaPlaylist(0, 2)
	_ = p.Append("test01.ts", 5.0, "")
	_ = p.Append("test02.ts", 6.0, "")
	fmt.Printf("%s\n", p)
	// Output:
	// #EXTM3U
	// #EXT-X-VERSION:3
	// #EXT-X-MEDIA-SEQUENCE:0
	// #EXT-X-TARGETDURATION:6
	// #EXTINF:5.000,
	// test01.ts
	// #EXTINF:6.000,
	// test02.ts
}

// Create new media playlist
// Add two segments to media playlist
// Print it
func ExampleNewMediaPlaylist_stringWinsize0VOD() {
	p, _ := NewMediaPlaylist(0, 2)
	_ = p.Append("test01.ts", 5.0, "")
	_ = p.Append("test02.ts", 6.0, "")
	p.Close()
	fmt.Printf("%s\n", p)
	// Output:
	// #EXTM3U
	// #EXT-X-VERSION:3
	// #EXT-X-MEDIA-SEQUENCE:0
	// #EXT-X-TARGETDURATION:6
	// #EXTINF:5.000,
	// test01.ts
	// #EXTINF:6.000,
	// test02.ts
	// #EXT-X-ENDLIST
}

// Create new master playlist
// Add media playlist
// Encode structures to HLS
func ExampleNewMasterPlaylist_string() {
	m := NewMasterPlaylist()
	p, _ := NewMediaPlaylist(3, 5)
	for i := 0; i < 5; i++ {
		_ = p.Append(fmt.Sprintf("test%d.ts", i), 5.0, "")
	}
	m.Append("chunklist1.m3u8", p, VariantParams{Bandwidth: 1500000, AverageBandwidth: 1500000,
		Resolution: "576x480", FrameRate: 25.000})
	m.Append("chunklist2.m3u8", p, VariantParams{Bandwidth: 1500000, AverageBandwidth: 1500000,
		Resolution: "576x480", FrameRate: 25.000})
	fmt.Printf("%s", m)
	// Output:
	// #EXTM3U
	// #EXT-X-VERSION:3
	// #EXT-X-STREAM-INF:BANDWIDTH=1500000,AVERAGE-BANDWIDTH=1500000,RESOLUTION=576x480,FRAME-RATE=25.000
	// chunklist1.m3u8
	// #EXT-X-STREAM-INF:BANDWIDTH=1500000,AVERAGE-BANDWIDTH=1500000,RESOLUTION=576x480,FRAME-RATE=25.000
	// chunklist2.m3u8
}

func ExampleNewMasterPlaylist_stringWithHLSv7() {
	m := NewMasterPlaylist()
	m.SetVersion(7)
	m.SetIndependentSegments(true)
	p, _ := NewMediaPlaylist(3, 5)
	m.Append("hdr10_1080/prog_index.m3u8", p, VariantParams{AverageBandwidth: 7964551, Bandwidth: 12886714, VideoRange: "PQ", Codecs: "hvc1.2.4.L123.B0", Resolution: "1920x1080", FrameRate: 23.976, Captions: "NONE", HDCPLevel: "TYPE-0"})
	m.Append("hdr10_1080/iframe_index.m3u8", p, VariantParams{Iframe: true, AverageBandwidth: 364552, Bandwidth: 905053, VideoRange: "PQ", Codecs: "hvc1.2.4.L123.B0", Resolution: "1920x1080", HDCPLevel: "TYPE-0"})
	fmt.Printf("%s", m)
	// Output:
	// #EXTM3U
	// #EXT-X-VERSION:7
	// #EXT-X-INDEPENDENT-SEGMENTS
	// #EXT-X-STREAM-INF:BANDWIDTH=12886714,AVERAGE-BANDWIDTH=7964551,CODECS="hvc1.2.4.L123.B0",RESOLUTION=1920x1080,FRAME-RATE=23.976,HDCP-LEVEL=TYPE-0,VIDEO-RANGE=PQ,CLOSED-CAPTIONS=NONE
	// hdr10_1080/prog_index.m3u8
	// #EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=905053,AVERAGE-BANDWIDTH=364552,CODECS="hvc1.2.4.L123.B0",RESOLUTION=1920x1080,HDCP-LEVEL=TYPE-0,VIDEO-RANGE=PQ,URI="hdr10_1080/iframe_index.m3u8"
}

func ExampleDecode_mediaPlaylistSegmentsSCTE35OATCLS() {
	f, _ := os.Open("sample-playlists/media-playlist-with-oatcls-scte35.m3u8")
	p, _, _ := DecodeFrom(bufio.NewReader(f), true)
	pp := p.(*MediaPlaylist)
	fmt.Print(pp)
	// Output:
	// #EXTM3U
	// #EXT-X-VERSION:3
	// #EXT-X-MEDIA-SEQUENCE:0
	// #EXT-X-TARGETDURATION:10
	// #EXT-OATCLS-SCTE35:/DAlAAAAAAAAAP/wFAUAAAABf+/+ANgNkv4AFJlwAAEBAQAA5xULLA==
	// #EXT-X-CUE-OUT:15
	// #EXTINF:8.844,
	// media0.ts
	// #EXT-X-CUE-OUT-CONT:ElapsedTime=8.844,Duration=15,SCTE35=/DAlAAAAAAAAAP/wFAUAAAABf+/+ANgNkv4AFJlwAAEBAQAA5xULLA==
	// #EXTINF:6.156,
	// media1.ts
	// #EXT-X-CUE-IN
	// #EXTINF:3.844,
	// media2.ts
}

func ExampleMediaPlaylist_Segments_scte35_67_2014() {
	f, _ := os.Open("sample-playlists/media-playlist-with-scte35.m3u8")
	p, _, _ := DecodeFrom(bufio.NewReader(f), true)
	pp := p.(*MediaPlaylist)
	fmt.Print(pp)
	// Output:
	// #EXTM3U
	// #EXT-X-VERSION:3
	// #EXT-X-MEDIA-SEQUENCE:0
	// #EXT-X-TARGETDURATION:10
	// #EXTINF:10.000,
	// media0.ts
	// #EXTINF:10.000,
	// media1.ts
	// #EXT-SCTE35:CUE="/DAIAAAAAAAAAAAQAAZ/I0VniQAQAgBDVUVJQAAAAH+cAAAAAA==",ID="123",TIME=123.12
	// #EXTINF:10.000,
	// media2.ts
}

// Range over segments of media playlist. Check for ring buffer corner
// cases.
func ExampleNewMediaPlaylist_getAllSegments() {
	m, _ := NewMediaPlaylist(3, 3)
	_ = m.Append("t00.ts", 10, "")
	_ = m.Append("t01.ts", 10, "")
	_ = m.Append("t02.ts", 10, "")
	for _, v := range m.GetAllSegments() {
		fmt.Printf("%s\n", v.URI)
	}
	_ = m.Remove()
	_ = m.Remove()
	_ = m.Remove()
	_ = m.Append("t03.ts", 10, "")
	_ = m.Append("t04.ts", 10, "")
	for _, v := range m.GetAllSegments() {
		fmt.Printf("%s\n", v.URI)
	}
	_ = m.Remove()
	_ = m.Remove()
	_ = m.Append("t05.ts", 10, "")
	_ = m.Append("t06.ts", 10, "")
	_ = m.Remove()
	_ = m.Remove()
	// empty because removed two elements
	for _, v := range m.GetAllSegments() {
		fmt.Printf("%s\n", v.URI)
	}
	// Output:
	// t00.ts
	// t01.ts
	// t02.ts
	// t03.ts
	// t04.ts
}

/****************
 *  Benchmarks  *
 ****************/

func BenchmarkEncodeMasterPlaylist(b *testing.B) {
	f, err := os.Open("sample-playlists/master.m3u8")
	if err != nil {
		b.Fatal(err)
	}
	p := NewMasterPlaylist()
	if err := p.DecodeFrom(bufio.NewReader(f), true); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		p.ResetCache()
		_ = p.Encode() // disregard output
	}
}

func BenchmarkEncodeMediaPlaylist(b *testing.B) {
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
	for i := 0; i < b.N; i++ {
		p.ResetCache()
		_ = p.Encode() // disregard output
	}
}

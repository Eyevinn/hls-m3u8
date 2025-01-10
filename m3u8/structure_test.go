package m3u8

/*
 Playlist structures tests.
*/

import (
	"bytes"
	"testing"
)

func CheckType(t *testing.T, p Playlist) {
	t.Logf("%T implements Playlist interface OK\n", p)
}

// Create new media playlist.
func TestNewMediaPlaylist(t *testing.T) {
	_, e := NewMediaPlaylist(1, 2)
	if e != nil {
		t.Fatalf("Create media playlist failed: %s", e)
	}
}

func TestSCTE35String(t *testing.T) {
	data := []struct {
		syntax   SCTE35Syntax
		expected string
	}{{SCTE35_NONE, "None"}, {SCTE35_67_2014, "SCTE35_67_2014"}, {SCTE35_OATCLS, "SCTE35_OATCLS"}, {SCTE35_DATERANGE, "SCTE35_DATERANGE"}}

	for _, d := range data {
		if d.syntax.String() != d.expected {
			t.Fatalf("Expected %s, got %s", d.expected, d.syntax.String())
		}
	}
}

type MockCustomTag struct {
	name          string
	err           error
	segment       bool
	encodedString string
}

func (t *MockCustomTag) TagName() string {
	return t.name
}

func (t *MockCustomTag) Decode(line string) (CustomTag, error) {
	return t, t.err
}

func (t *MockCustomTag) Encode() *bytes.Buffer {
	if t.encodedString == "" {
		return nil
	}

	buf := new(bytes.Buffer)

	buf.WriteString(t.encodedString)

	return buf
}

func (t *MockCustomTag) String() string {
	return t.encodedString
}

func (t *MockCustomTag) SegmentTag() bool {
	return t.segment
}

package m3u8

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestCalcMinVersionMasterPlaylist(t *testing.T) {
	is := is.New(t)
	pl3 := NewMasterPlaylist()

	pl7 := NewMasterPlaylist()
	pl7.Variants = append(pl7.Variants, &Variant{
		VariantParams: VariantParams{
			Alternatives: []*Alternative{{InstreamId: "SERVICE1", Type: "CLOSED-CAPTIONS"}},
		},
	})

	pl11, err := readTestMasterPlaylist(t, "sample-playlists/master-with-defines.m3u8")
	is.NoErr(err) // must decode sample-playlists/master-with-defines.m3u8

	pl12, err := readTestMasterPlaylist(t, "sample-playlists/master-with-req-video-layout.m3u8")
	is.NoErr(err) // must decode sample-playlists/master-with-req-video-layout.m3u8

	pl13, err := readTestMasterPlaylist(t, "sample-playlists/master-with-non-closed-captions-instream-id.m3u8")
	is.NoErr(err) // must decode sample-playlists/master-with-non-closed-captions-instream-id.m3u8

	cases := []struct {
		playlist        Playlist
		expectedVersion uint8
		expectedReason  string
	}{
		{pl3, minVer, "minimal version supported by this library"},
		{pl7, 7, "SERVICE value for the INSTREAM-ID attribute of the EXT-X-MEDIA"},
		{pl11, 11, "EXT-X-DEFINE tag with a QUERYPARAM attribute"},
		{pl12, 12, "REQ- attribute"},
		{pl13, 13, "EXT-X-MEDIA tag with INSTREAM-ID attribute for non CLOSED-CAPTIONS TYPE"},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			is := is.New(t)
			ver, reason := c.playlist.CalcMinVersion()
			is.Equal(ver, c.expectedVersion)
			is.Equal(reason, c.expectedReason)
		})
	}
}

func TestCalcMinVersionMediaPlaylist(t *testing.T) {

	is := is.New(t)

	pl3, err := NewMediaPlaylist(10, 10)
	is.NoErr(err) // must create media playlist

	pl4ByteRange, err := readTestMediaPlaylist(t, "sample-playlists/media-playlist-with-byterange.m3u8")
	is.NoErr(err) // must decode sample-playlists/media-playlist-with-byterange.m3u8

	pl4IframesOnly, err := readTestMediaPlaylist(t, "sample-playlists/media-playlist-with-iframes-only.m3u8")
	is.NoErr(err) // must decode sample-playlists/media-playlist-with-iframes-only.m3u8

	pl5IframesOnlyAndMap, err := readTestMediaPlaylist(t, "sample-playlists/media-playlist-with-iframes-only-and-map.m3u8")
	is.NoErr(err) // must decode sample-playlists/media-playlist-with-iframes-only-and-map.m3u8

	pl5SampleAES, err := readTestMediaPlaylist(t, "sample-playlists/media-playlist-with-key.m3u8")
	is.NoErr(err) // must decode sample-playlists/media-playlist-with-key.m3u8

	pl6Fmp4, err := readTestMediaPlaylist(t, "sample-playlists/media-playlist-fmp4.m3u8")
	is.NoErr(err) // must decode sample-playlists/media-playlist-fmp4.m3u8

	pl8VariableSubstitution, err := readTestMediaPlaylist(t, "sample-playlists/media-playlist-with-defines.m3u8")
	is.NoErr(err) // must decode sample-playlists/media-playlist-with-defines.m3u8

	pl11QueryParam, err := readTestMediaPlaylist(t, "sample-playlists/media-playlist-with-queryparam.m3u8")
	is.NoErr(err) // must decode sample-playlists/media-playlist-with-queryparam.m3u8

	cases := []struct {
		playlist        Playlist
		expectedVersion uint8
		expectedReason  string
	}{
		{pl3, minVer, "minimal version supported by this library"},
		{pl4ByteRange, 4, "EXT-X-BYTERANGE tag"},
		{pl4IframesOnly, 4, "EXT-X-I-FRAMES-ONLY tag"},
		{pl5IframesOnlyAndMap, 5, "EXT-X-MAP tag"},
		{pl5SampleAES, 5, "EXT-X-KEY tag with a METHOD of SAMPLE-AES, KEYFORMAT or KEYFORMATVERSIONS attributes"},
		{pl6Fmp4, 6, "EXT-X-MAP tag in a Media Playlist that does not contain EXT-X-I-FRAMES-ONLY"},
		{pl8VariableSubstitution, 8, "Variable substitution"},
		{pl11QueryParam, 11, "EXT-X-DEFINE tag with a QUERYPARAM attribute"},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			is := is.New(t)
			ver, reason := c.playlist.CalcMinVersion()
			is.Equal(ver, c.expectedVersion)
			is.Equal(reason, c.expectedReason)
		})
	}
}

func readTestPlaylist(t *testing.T, fileName string) Playlist {
	t.Helper()
	f, err := os.Open(fileName)
	if err != nil {
		t.Fail()
	}
	defer f.Close()

	p, _, err := DecodeFrom(bufio.NewReader(f), false)
	if err != nil {
		t.Fail()
	}
	return p
}

func TestAllPlaylistVersions(t *testing.T) {
	is := is.New(t)

	// Read all m3u8 files in sample-playlists directory
	files, err := os.ReadDir("sample-playlists")
	is.NoErr(err)

	for _, file := range files {
		fName := file.Name()
		if !strings.HasSuffix(fName, ".m3u8") {
			continue
		}

		t.Run(fName, func(t *testing.T) {
			path := "sample-playlists/" + fName

			p := readTestPlaylist(t, path)

			minVer, reason := p.CalcMinVersion()
			actualVer := p.Version()
			if minVer > actualVer {
				t.Errorf("Playlist %s: CalcMinVersion=%d but Version=%d (reason: %s)",
					fName, minVer, actualVer, reason)
			}
		})
	}
}

// TestEncodeRaisesVersionToMinVersion checks that the version signaled by EXT-X-VERSION
// is raised to the minimum version required by the playlist content (issue #95).
func TestEncodeRaisesVersionToMinVersion(t *testing.T) {
	mapNoIframes := func(t *testing.T) Playlist {
		t.Helper()
		p, err := NewMediaPlaylist(0, 2)
		if err != nil {
			t.Fatal(err)
		}
		p.SetDefaultMap("init.mp4", 0, 0)
		if err := p.Append("seg1.m4s", 4.0, ""); err != nil {
			t.Fatal(err)
		}
		return p
	}

	cases := []struct {
		desc            string
		makePlaylist    func(t *testing.T) Playlist
		expectedVersion uint8
	}{
		{
			desc:            "EXT-X-MAP without EXT-X-I-FRAMES-ONLY requires version 6",
			makePlaylist:    mapNoIframes,
			expectedVersion: 6,
		},
		{
			desc: "EXT-X-MAP with EXT-X-I-FRAMES-ONLY requires version 5",
			makePlaylist: func(t *testing.T) Playlist {
				t.Helper()
				p := mapNoIframes(t).(*MediaPlaylist)
				p.SetIframeOnly()
				return p
			},
			expectedVersion: 5,
		},
		{
			desc: "higher version set with SetVersion is kept",
			makePlaylist: func(t *testing.T) Playlist {
				t.Helper()
				p := mapNoIframes(t)
				p.SetVersion(9)
				return p
			},
			expectedVersion: 9,
		},
		{
			desc: "EXT-X-DEFINE with QUERYPARAM requires version 11",
			makePlaylist: func(t *testing.T) Playlist {
				t.Helper()
				p, err := NewMediaPlaylist(0, 2)
				if err != nil {
					t.Fatal(err)
				}
				p.AppendDefine(Define{Name: "token", Type: QUERYPARAM})
				if err := p.Append("seg1.ts", 4.0, ""); err != nil {
					t.Fatal(err)
				}
				return p
			},
			expectedVersion: 11,
		},
		{
			desc: "SERVICE value for INSTREAM-ID requires version 7 in master playlist",
			makePlaylist: func(t *testing.T) Playlist {
				t.Helper()
				p := NewMasterPlaylist()
				p.Append("variant.m3u8", nil, VariantParams{
					Bandwidth: 1000,
					Alternatives: []*Alternative{
						{GroupId: "cc", Type: "CLOSED-CAPTIONS", Name: "English", InstreamId: "SERVICE1"},
					},
				})
				return p
			},
			expectedVersion: 7,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			is := is.New(t)
			p := c.makePlaylist(t)
			out := p.Encode().String()
			is.Equal(p.Version(), c.expectedVersion) // version after encoding
			wantTag := fmt.Sprintf("#EXT-X-VERSION:%d\n", c.expectedVersion)
			is.True(strings.Contains(out, wantTag)) // EXT-X-VERSION tag in output
		})
	}
}

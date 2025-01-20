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
			Alternatives: []*Alternative{{InstreamId: "SERVICE1"}},
		},
	})

	pl11, err := readTestMasterPlaylist(t, "sample-playlists/master-with-defines.m3u8")
	is.NoErr(err) // must decode sample-playlists/master-with-defines.m3u8

	pl12, err := readTestMasterPlaylist(t, "sample-playlists/master-with-req-video-layout.m3u8")
	is.NoErr(err) // must decode sample-playlists/master-with-req-video-layout.m3u8

	cases := []struct {
		playlist        Playlist
		expectedVersion uint8
		expectedReason  string
	}{
		{pl3, minVer, "minimal version supported by this library"},
		{pl7, 7, "SERVICE value for the INSTREAM-ID attribute of the EXT-X-MEDIA"},
		{pl11, 11, "EXT-X-DEFINE tag with a QUERYPARAM attribute"},
		{pl12, 12, "REQ- attribute"},
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

package m3u8

import (
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

	cases := []struct {
		playlist        Playlist
		expectedVersion uint8
		expectedReason  string
	}{
		{pl3, 3, "minimal version supported by this library"},
		{pl7, 7, "SERVICE value for the INSTREAM-ID attribute of the EXT-X-MEDIA"},
	}

	for _, c := range cases {
		gotVersion, gotReason, err := c.playlist.CalcMinVersion()
		is.NoErr(err)                           // no error when checking the version
		is.Equal(gotVersion, c.expectedVersion) // version is as expected
		is.Equal(gotReason, c.expectedReason)   // reason is as expected

	}

}

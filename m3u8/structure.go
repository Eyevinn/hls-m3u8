package m3u8

/*
 This file defines data structures related to package.
*/

import (
	"bytes"
	"io"
	"time"
)

// Playlist interface applied to various playlist types.
type Playlist interface {
	Encode() *bytes.Buffer
	Decode(bytes.Buffer, bool) error
	DecodeFrom(reader io.Reader, strict bool) error
	WithCustomDecoders([]CustomDecoder) Playlist
	String() string
}

// CustomDecoder interface for decoding custom and unsupported tags
type CustomDecoder interface {
	// TagName should return the full identifier including the leading '#' as well as the
	// trailing ':' if the tag also contains a value or attribute list
	TagName() string
	// Decode parses a line from the playlist and returns the CustomTag representation
	Decode(line string) (CustomTag, error)
	// SegmentTag should return true if this CustomDecoder should apply per segment.
	// Should returns false if it a MediaPlaylist header tag.
	// This value is ignored for MasterPlaylists.
	SegmentTag() bool
}

// CustomTag interface for encoding custom and unsupported tags
type CustomTag interface {
	// TagName should return the full identifier including the leading '#' as well as the
	// trailing ':' if the tag also contains a value or attribute list
	TagName() string
	// Encode should return the complete tag string as a *bytes.Buffer. This will
	// be used by Playlist.Decode to write the tag to the m3u8.
	// Return nil to not write anything to the m3u8.
	Encode() *bytes.Buffer
	// String should return the encoded tag as a string.
	String() string
}

const (
	// minVer is the minimum version of the HLS protocol supported by this package.
	// Version 3, means that floating point EXTINF durations are used.
	// [Protocol Version Compatibility]
	minVer = uint8(3)

	// DATETIME represents format for EXT-X-PROGRAM-DATE-TIME timestamps.
	// Format is [ISO/IEC 8601:2004] according to the [HLS spec].
	DATETIME = time.RFC3339Nano
)

// ListType is type of the playlist.
type ListType uint

const (
	// use 0 for undefined type
	MASTER ListType = iota + 1
	MEDIA
)

// MediaType is the type for EXT-X-PLAYLIST-TYPE tag
type MediaType uint

const (
	// use 0 for undefined
	EVENT MediaType = iota + 1
	VOD
)

// SCTE35Syntax defines the format of the SCTE-35 cue points which do not use
// the the HLS EXT-X-DATERANGE tag and instead use custom tags.
type SCTE35Syntax uint

const (
	// SCTE35_67_2014 is the default due to backwards compatibility reasons.
	SCTE35_67_2014 SCTE35Syntax = iota // SCTE35_67_2014 defined in [scte67]
	SCTE35_OATCLS                      // SCTE35_OATCLS is a non-standard but common format
)

// SCTE35CueType defines the type of cue point
type SCTE35CueType uint

const (
	SCTE35Cue_Start SCTE35CueType = iota // SCTE35Cue_Start indicates an cue-out point
	SCTE35Cue_Mid                        // SCTE35Cue_Mid indicates a segment between start and end cue points
	SCTE35Cue_End                        // SCTE35Cue_End indicates an cue-in point
)

// MediaPlaylist structure represents a single bitrate playlist aka
// media playlist. It related to both a simple media playlists and a
// sliding window live media playlists with window size.
// URI lines in the Playlist point to media segments.
type MediaPlaylist struct {
	TargetDuration   float64 // TargetDuration is the maximum media segment duration in seconds (an integer)
	SeqNo            uint64  // EXT-X-MEDIA-SEQUENCE
	Segments         []*MediaSegment
	Args             string // optional query placed after URIs (URI?Args)
	Iframe           bool   // EXT-X-I-FRAMES-ONLY
	Closed           bool   // is this VOD (closed) or Live (sliding) playlist?
	MediaType        MediaType
	DiscontinuitySeq uint64 // EXT-X-DISCONTINUITY-SEQUENCE
	StartTime        float64
	StartTimePrecise bool
	winsize          uint         // max number of segments encoded sliding playlist, set to 0 for VOD and EVENT
	capacity         uint         // total capacity of slice used for the playlist
	head             uint         // head of FIFO, we add segments to head
	tail             uint         // tail of FIFO, we remove segments from tail
	count            uint         // number of segments added to the playlist
	buf              bytes.Buffer // buffer used for encoding and caching playlist
	ver              uint8        // protocol version of the playlist, 3 or higher
	Key              *Key         // Key correspnds to optional EXT-X-KEY tag for encrypted segments
	// Map = EXT-X-MAP (optional) provides address to a Media Initialization Section. Can be redefined in segments.
	Map            *Map
	Custom         map[string]CustomTag
	customDecoders []CustomDecoder
}

// MasterPlaylist structure represents a master playlist which
// combines media playlists for multiple bitrates. URI lines in the
// playlist identify media playlists.
type MasterPlaylist struct {
	Variants            []*Variant
	Args                string               // optional query placed after URI (URI?Args)
	buf                 bytes.Buffer         // buffer used for encoding and caching playlist
	ver                 uint8                // protocol version of the playlist, 3 or higher
	independentSegments bool                 // Global tag for EXT-X-INDEPENDENT-SEGMENTS
	Custom              map[string]CustomTag // Custom provided custom tags for encoding
	customDecoders      []CustomDecoder      // customDecoders provided custom tags for decoding
}

// Variant structure represents media playlist variants in master playlists.
type Variant struct {
	URI       string
	Chunklist *MediaPlaylist
	VariantParams
}

// VariantParams represents parameters for a Variant.
// Used in EXT-X-STREAM-INF and EXT-X-I-FRAME-STREAM-INF.
type VariantParams struct {
	ProgramId        uint32
	Bandwidth        uint32
	AverageBandwidth uint32 // EXT-X-STREAM-INF only
	Codecs           string
	Resolution       string
	Audio            string // EXT-X-STREAM-INF only
	Video            string
	Subtitles        string // EXT-X-STREAM-INF only
	Captions         string // EXT-X-STREAM-INF only
	Name             string
	Iframe           bool // EXT-X-I-FRAME-STREAM-INF
	VideoRange       string
	HDCPLevel        string
	FrameRate        float64        // EXT-X-STREAM-INF
	Alternatives     []*Alternative // EXT-X-MEDIA
}

// Alternative represents EXT-X-MEDIA tag in variants.
// Listed in same order as in specification for easy comparison.
type Alternative struct {
	Type              string
	URI               string
	GroupId           string
	Language          string
	AssocLanguage     string
	Name              string
	StableRenditionId string
	Default           bool
	Autoselect        bool
	Forced            bool
	InstreamId        string
	BitDepth          byte
	SampleRate        uint32
	Characteristics   string
	Channels          string
}

// MediaSegment structure represents a media segment included in a
// media playlist. Media segment may be encrypted.
type MediaSegment struct {
	SeqId uint64
	Title string // optional second parameter for EXTINF tag
	URI   string
	// Duration is the first parameter for EXTINF tag.
	// It provides the duration in seconds of the segment.
	// if  protocol version is 2 or less, its value must be an integer.
	Duration float64
	Limit    int64 // EXT-X-BYTERANGE <n> is length in bytes for the file under URI.
	Offset   int64 // EXT-X-BYTERANGE [@o] is offset from the start of the file under URI.
	// Key = EXT-X-KEY  changes the key for encryption until next EXT-X-KEY tag.
	Key *Key
	// Map = EXT-X-MAP changes the Media Initialization Section until next EXT-X-MAP tag.
	Map *Map
	// Discontinuity = EXT-X-DISCONTINUITY signals a discontinuity between the
	// following and preceding media segments.
	Discontinuity bool
	SCTE          *SCTE // SCTE-35 used for Ad signaling in HLS.
	// ProgramDateTime is EXT-X-PROGRAM-DATE-TIME tag .
	// It associates the first sample of a media segment with an absolute date and/or time.
	ProgramDateTime time.Time
	// Custom holds custom tags
	Custom map[string]CustomTag
}

// SCTE holds custom, non EXT-X-DATERANGE, SCTE-35 tags
type SCTE struct {
	Syntax  SCTE35Syntax  // Syntax defines the format of the SCTE-35 cue tag
	CueType SCTE35CueType // CueType defines whether the cue is a start, mid, end (if applicable)
	Cue     string
	ID      string
	Time    float64
	Elapsed float64
}

// Key structure represents information about stream encryption (EXT-X-KEY tag)
type Key struct {
	Method            string
	URI               string
	IV                string
	Keyformat         string
	Keyformatversions string
}

// Map (EXT-X-MAP tag) specifies how obtain the Media
// Initialization Section required to parse the applicable
// Media Segments.
//
// It applies to every Media Segment that appears after it in the
// Playlist until the next EXT-X-MAP tag or until the end of the
// playlist.
type Map struct {
	URI    string
	Limit  int64 // <n> is length in bytes for the file under URI
	Offset int64 // [@o] is offset from the start of the file under URI
}

// Internal structure for decoding a line of input stream with a list type detection
type decodingState struct {
	listType           ListType
	m3u                bool
	tagStreamInf       bool
	tagInf             bool
	tagSCTE35          bool
	tagRange           bool
	tagDiscontinuity   bool
	tagProgramDateTime bool
	tagKey             bool
	tagMap             bool
	tagCustom          bool
	programDateTime    time.Time
	limit              int64
	offset             int64
	duration           float64
	title              string
	variant            *Variant
	alternatives       []*Alternative
	xkey               *Key
	xmap               *Map
	scte               *SCTE
	custom             map[string]CustomTag
}

/*
[scte67]: http://www.scte.org/documents/pdf/standards/SCTE%2067%202014.pdf
[hls-spec]: https://datatracker.ietf.org/doc/html/draft-pantos-hls-rfc8216bis-16
[ISO/IEC 8601:2004]:http://www.iso.org/iso/catalogue_detail?csnumber=40874
[Protocol Version Compatibility]: https://datatracker.ietf.org/doc/html/draft-pantos-hls-rfc8216bis-16#section-8
*/

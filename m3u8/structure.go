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

// CustomMap maps custom tags names to CustomTag
type CustomMap map[string]CustomTag

const (
	// minVer is the minimum version of the HLS protocol supported by this package.
	// Version 3, means that floating point EXTINF durations are used.
	// [Protocol Version Compatibility]
	minVer = uint8(3)

	// DATETIME represents format for EXT-X-PROGRAM-DATE-TIME timestamps.
	// Format is [ISO/IEC 8601:2004] according to the [HLS spec].
	DATETIME = time.RFC3339Nano
)

// ListType is type of playlist.
type ListType uint

const (
	// use 0 for undefined type
	MASTER ListType = iota + 1
	MEDIA
)

// MediaType is EXT-X-PLAYLIST-TYPE tag
type MediaType uint

const (
	// use 0 for undefined
	EVENT MediaType = iota + 1
	VOD
)

// SCTE35Syntax defines the format of the SCTE-35 cue points including EXT-X-DATERANGE version.
type SCTE35Syntax uint

const (
	// SCTE35_67_2014 is the default due to backwards compatibility reasons.
	SCTE35_67_2014   SCTE35Syntax = iota // SCTE35_67_2014 defined in [scte67]
	SCTE35_OATCLS                        // SCTE35_OATCLS is a non-standard but common format
	SCTE35_DATERANGE                     // SCTE35_DATERANGE is standard format for HLS
)

// SCTE35CueType defines the type of cue point
type SCTE35CueType uint

const (
	SCTE35Cue_Start SCTE35CueType = iota // SCTE35Cue_Start indicates an cue-out point
	SCTE35Cue_Mid                        // SCTE35Cue_Mid indicates a segment between start and end cue points
	SCTE35Cue_End                        // SCTE35Cue_End indicates a cue-in point
	SCTE35Cue_Cmd                        // Not in, out, or mid. Indicates a command for splice.
)

// MediaPlaylist represents a single bitrate playlist aka media playlist.
// It is used for both VOD, EVENT and sliding window live media playlists with window size.
// URI lines in the Playlist point to media segments.
type MediaPlaylist struct {
	TargetDuration   uint            // TargetDuration is max media segment duration. Rounding depends on version.
	SeqNo            uint64          // EXT-X-MEDIA-SEQUENCE
	Segments         []*MediaSegment // List of segments in the playlist. Output may be limited by winsize.
	Args             string          // optional query placed after URIs (URI?Args)
	Iframe           bool            // EXT-X-I-FRAMES-ONLY
	Closed           bool            // is this VOD/EVENT (closed) or Live (sliding) playlist?
	MediaType        MediaType       // EXT-X-PLAYLIST-TYPE (EVENT, VOD or empty)
	DiscontinuitySeq uint64          // EXT-X-DISCONTINUITY-SEQUENCE
	StartTime        float64         // EXT-X-START:TIME-OFFSET=<n>
	StartTimePrecise bool            // EXT-X-START:PRECISE=YES
	Key              *Key            // EXT-X-KEY is initial key tag for encrypted segments
	Map              *Map            // EXT-X-MAP provides a Media Initialization Section. Segments can redefine.
	DateRanges       []*DateRange    // EXT-X-DATERANGE tags not associated with SCTE-35
	Custom           CustomMap       // Custom-provided tags for encoding
	customDecoders   []CustomDecoder // customDecoders provides custom tags for decoding
	winsize          uint            // max number of segments encoded sliding playlist, set to 0 for VOD and EVENT
	capacity         uint            // total capacity of slice used for the playlist
	head             uint            // head of FIFO, we add segments to head
	tail             uint            // tail of FIFO, we remove segments from tail
	count            uint            // number of segments added to the playlist
	buf              bytes.Buffer    // buffer used for encoding and caching playlist output
	ver              uint8           // protocol version of the playlist, 3 or higher
	targetDurLocked  bool            // target duration is locked and cannot be changed

}

// MasterPlaylist represents a master (multivariant) playlist which
// provides parameters and lists one or more media playlists. URI lines in the
// playlist identify media playlists.
type MasterPlaylist struct {
	Variants            []*Variant      // Variants is a list of media playlists
	Args                string          // optional query placed after URI (URI?Args)
	buf                 bytes.Buffer    // buffer used for encoding and caching playlist
	ver                 uint8           // protocol version of the playlist, 3 or higher
	independentSegments bool            // Global tag for EXT-X-INDEPENDENT-SEGMENTS
	Custom              CustomMap       // Custom-provided tags for encoding
	customDecoders      []CustomDecoder // customDecoders provided custom tags for decoding
}

// Variant structure represents media playlist variants in master playlists.
type Variant struct {
	URI       string         // URI is the path to the media playlist.
	Chunklist *MediaPlaylist // Chunklist is the media playlist for the variant.
	VariantParams
}

// VariantParams represents parameters for a Variant.
// Used in EXT-X-STREAM-INF and EXT-X-I-FRAME-STREAM-INF.
type VariantParams struct {
	ProgramId        uint32         // PROGRAM-ID parameter. Removed in version 6
	Bandwidth        uint32         // BANDWIDTH parameter
	AverageBandwidth uint32         // AVERAGE-BANDWIDTH parameter
	Codecs           string         // CODECS parameter
	Resolution       string         // RESOLUTION parameter
	Audio            string         // AUDIO parameter. EXT-X-STREAM-INF only
	Video            string         // VIDEO parameter. EXT-X-STREAM-INF only
	Subtitles        string         // SUBTITLES parameter. EXT-X-STREAM-INF only
	Captions         string         // CLOSED-CAPTIONS parameter. EXT-X-STREAM-INF only
	Name             string         // NAME parameter. EXT-X-STREAM-INF only. Non-standard Wowza/JWPlayer extension
	Iframe           bool           // EXT-X-I-FRAME-STREAM-INF flag.
	VideoRange       string         // VIDEO-RANGE parameter
	HDCPLevel        string         // HDCP-LEVEL parameter
	FrameRate        float64        // FRAME-RATE parameter. EXT-X-STREAM-INF only
	Alternatives     []*Alternative // EXT-X-MEDIA parameters
}

// Alternative represents an EXT-X-MEDIA tag.
// Attributes are listed in same order as in specification for easy comparison.
type Alternative struct {
	Type              string // TYPE parameter
	URI               string // URI parameter
	GroupId           string // GROUP-ID parameter
	Language          string // LANGUAGE parameter
	AssocLanguage     string // ASSOC-LANGUAGE parameter
	Name              string // NAME parameter
	StableRenditionId string // STABLE-RENDITION-ID parameter
	Default           bool   // DEFAULT parameter
	Autoselect        bool   // AUTOSELECT parameter
	Forced            bool   // FORCED parameter
	InstreamId        string // INSTREAM-ID parameter
	BitDepth          byte   // BIT-DEPTH parameter
	SampleRate        uint32 // SAMPLE-RATE parameter
	Characteristics   string // CHARACTERISTICS parameter
	Channels          string // CHANNELS parameter
}

// MediaSegment represents a media segment included in a
// media playlist. Media segment may be encrypted.
type MediaSegment struct {
	SeqId           uint64    // SeqId is the sequence number of the segment. Should be unique and consecutive.
	URI             string    // URI is the path to the media segment.
	Duration        float64   // EXTINF first parameter. Duration in seconds.
	Title           string    // EXTINF optional second parameter.
	Limit           int64     // EXT-X-BYTERANGE <n> is length in bytes for the file under URI.
	Offset          int64     // EXT-X-BYTERANGE [@o] is offset from the start of the file under URI.
	Key             *Key      // EXT-X-KEY  changes the key for encryption until next EXT-X-KEY tag.
	Map             *Map      // EXT-X-MAP changes the Media Initialization Section until next EXT-X-MAP tag.
	Discontinuity   bool      // EXT-X-DISCONTINUITY signals a discontinuity between the surrounding segments.
	SCTE            *SCTE     // SCTE-35 used for Ad signaling in HLS.
	ProgramDateTime time.Time // EXT-X-PROGRAM-DATE-TIME associates the first sample with an absolute date and/or time.
	Custom          CustomMap // Custom holds custom tags
}

// SCTE holds both custom and EXT-X-DATERANGE SCTE-35 tags.
type SCTE struct {
	Syntax          SCTE35Syntax  // Syntax defines the format of the SCTE-35 cue tag
	CueType         SCTE35CueType // CueType defines whether the cue is a start, mid, end, cmd (as applicable)
	Cue             string        // Base64 encoded SCTE-35 cue message
	ID              string        // Unique ID
	Time            float64       // TIME for SCTE-67 and OATCLS SCTE-35 signalling
	Elapsed         float64       // ELAPSED for OATCLS SCTE-35 signalling
	PlannedDuration *float64      // PLANNED-DURATION for DATERANGE SCTE-35 signalling
	Duration        *float64      // DURATION in seconds for OATCLS and DATERANGE SCTE-35 signalling
	StartDate       *time.Time    // START-DATE for DATERANGE SCTE-35 signalling
	EndDate         *time.Time    // END-DATE for DATERANGE SCTE-35 signalling
}

// Key structure represents information about stream encryption (EXT-X-KEY tag)
type Key struct {
	Method            string // METHOD parameter
	URI               string // URI parameter
	IV                string // IV parameter
	Keyformat         string // KEYFORMAT parameter
	Keyformatversions string // KEYFORMATVERSIONS parameter
}

// Map (EXT-X-MAP tag) specifies how obtain the Media
// Initialization Section required to parse the applicable
// Media Segments.
//
// It applies to every Media Segment that appears after it in the
// Playlist until the next EXT-X-MAP tag or until the end of the
// playlist.
type Map struct {
	URI    string // URI is the path to the Media Initialization Section.
	Limit  int64  // <n> is length in bytes for the file under URI
	Offset int64  // [@o] is offset from the start of the file under URI
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
	custom             CustomMap
}

// DateRange corresponds to EXT-X-DATERANGE tag.
// It is used for signaling SCTE-35 messages,interstitials, and other metadata events.
type DateRange struct {
	ID              string      // ID is mandatory quoted string ID
	Class           string      // CLASS is a client-defined quoted string
	StartDate       time.Time   // START-DATE is mandatory start time
	EndDate         *time.Time  // END-DATE is optional end time
	Cue             string      // CUE is an enumerated-string-list of Trigger Identifiers, PRE, POST, or ONCE.
	Duration        *float64    // DURATION is optional duration in seconds
	PlannedDuration *float64    // PLANNED-DURATION is optional planned duration in seconds
	XAttrs          []Attribute // XAttrs is a list of X-<client-attribute>
	SCTE35Cmd       string      // SCTE35-CMD is a optional hex value for SCTE35 command
	SCTE35Out       string      // SCTE35-OUT is a optional hex value for SCTE35 CUE-OUT command
	SCTE35In        string      // SCTE35-IN is a optional hex value for SCTE35 CUE-IN command
	EndOnNext       bool        // END-ON-NEXT is enumerated YES/NO
}

// Attribute provides a raw key-value pair for an attribute. Quotes and 0x are included
type Attribute struct {
	Key string // Name of the attribute
	Val string // Value including quotes if a quoted string, and 0x if hexadecimal value
}

/*
[scte67]: http://www.scte.org/documents/pdf/standards/SCTE%2067%202014.pdf
[hls-spec]: https://datatracker.ietf.org/doc/html/draft-pantos-hls-rfc8216bis-16
[ISO/IEC 8601:2004]:http://www.iso.org/iso/catalogue_detail?csnumber=40874
[Protocol Version Compatibility]: https://datatracker.ietf.org/doc/html/draft-pantos-hls-rfc8216bis-16#section-8
*/

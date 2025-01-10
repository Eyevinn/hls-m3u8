package m3u8

/*
 This file defines functions related to playlist generation.
*/

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

var ErrPlaylistFull = errors.New("playlist is full")
var ErrPlaylistEmpty = errors.New("playlist is empty")
var ErrWinSizeTooSmall = errors.New("window size must be >= capacity")

// updateVersion updates the version if it is higher than before.
func updateVersion(ver *uint8, newVer uint8) {
	if *ver < newVer {
		*ver = newVer
	}
}

func strVer(ver uint8) string {
	return strconv.FormatUint(uint64(ver), 10)
}

// NewMasterPlaylist creates a new empty master playlist.
func NewMasterPlaylist() *MasterPlaylist {
	p := new(MasterPlaylist)
	p.ver = minVer
	return p
}

// Append appends a variant to master playlist. This operation resets the cache.
func (p *MasterPlaylist) Append(uri string, chunklist *MediaPlaylist, params VariantParams) {
	v := new(Variant)
	v.URI = uri
	v.Chunklist = chunklist
	v.VariantParams = params
	p.Variants = append(p.Variants, v)
	if len(v.Alternatives) > 0 {
		// This is not needed according to [Protocol Version Compatibility]
		// but remains for backwards compatibility reasons. Set the version
		// manually by using SerVersion.
		updateVersion(&p.ver, 4)
	}
	p.buf.Reset()
}

// ResetCache resets the playlist's cache (its buffer).
func (p *MasterPlaylist) ResetCache() {
	p.buf.Reset()
}

// Encode generates the output in M3U8 format and provides a pointer to its buffer.
func (p *MasterPlaylist) Encode() *bytes.Buffer {
	if p.buf.Len() > 0 {
		return &p.buf
	}

	p.buf.WriteString("#EXTM3U\n#EXT-X-VERSION:")
	p.buf.WriteString(strVer(p.ver))
	p.buf.WriteRune('\n')

	if p.IndependentSegments() {
		p.buf.WriteString("#EXT-X-INDEPENDENT-SEGMENTS\n")
	}

	// Write any custom master tags
	if p.Custom != nil {
		for _, v := range p.Custom {
			if customBuf := v.Encode(); customBuf != nil {
				p.buf.WriteString(customBuf.String())
				p.buf.WriteRune('\n')
			}
		}
	}

	alts := p.GetAllAlternatives()
	for _, alt := range alts {
		writeExtXMedia(&p.buf, alt)
	}

	for _, vnt := range p.Variants {
		if vnt.Iframe {
			writeExtXIFrameStreamInf(&p.buf, vnt)
		} else {
			writeExtXStreamInf(&p.buf, vnt)
			p.buf.WriteString(vnt.URI)
			if p.Args != "" {
				if strings.Contains(vnt.URI, "?") {
					p.buf.WriteRune('&')
				} else {
					p.buf.WriteRune('?')
				}
				p.buf.WriteString(p.Args)
			}
			p.buf.WriteRune('\n')
		}
	}

	return &p.buf
}

// writeExtXMedia writes an EXT-X-MEDIA tag line including \n to the buffer.
// No checks are done that the date is valid.
func writeExtXMedia(buf *bytes.Buffer, alt *Alternative) {
	buf.WriteString("#EXT-X-MEDIA:")
	buf.WriteString("TYPE=")
	buf.WriteString(alt.Type)                 // Mandatory enumerated string
	writeQuoted(buf, "GROUP-ID", alt.GroupId) // Mandatory quoted-string
	writeQuoted(buf, "NAME", alt.Name)        // Mandatory quoted-string
	if alt.Language != "" {
		writeQuoted(buf, "LANGUAGE", alt.Language)
	}
	if alt.AssocLanguage != "" {
		writeQuoted(buf, "ASSOC-LANGUAGE", alt.AssocLanguage)
	}
	if alt.StableRenditionId != "" {
		writeQuoted(buf, "STABLE-RENDITION-ID", alt.StableRenditionId)
	}
	buf.WriteString(",DEFAULT=")
	writeYESorNO(buf, alt.Default)
	if alt.Autoselect {
		buf.WriteString(",AUTOSELECT=YES")
	}
	if alt.Forced {
		buf.WriteString(",FORCED=YES")
	}
	if alt.InstreamId != "" {
		writeQuoted(buf, "INSTREAM-ID", alt.InstreamId)
	}
	if alt.BitDepth != 0 {
		writeUint(buf, "BIT-DEPTH", uint(alt.BitDepth))
	}
	if alt.SampleRate != 0 {
		writeUint(buf, "SAMPLE-RATE", uint(alt.SampleRate))
	}
	if alt.Characteristics != "" {
		writeQuoted(buf, "CHARACTERISTICS", alt.Characteristics)
	}
	if alt.Channels != "" {
		writeQuoted(buf, "CHANNELS", alt.Channels)
	}
	if alt.URI != "" {
		writeQuoted(buf, "URI", alt.URI)
	}
	buf.WriteRune('\n')
}

func writeExtXStreamInf(buf *bytes.Buffer, vnt *Variant) {
	buf.WriteString("#EXT-X-STREAM-INF:BANDWIDTH=")
	buf.WriteString(strconv.FormatUint(uint64(vnt.Bandwidth), 10))
	if vnt.AverageBandwidth != 0 {
		buf.WriteString(",AVERAGE-BANDWIDTH=")
		buf.WriteString(strconv.FormatUint(uint64(vnt.AverageBandwidth), 10))
	}
	if vnt.Score > 0 {
		writeFloat(buf, "SCORE", vnt.Score)
	}
	if vnt.Codecs != "" {
		writeQuoted(buf, "CODECS", vnt.Codecs)
	}
	if vnt.SupplementalCodecs != "" {
		writeQuoted(buf, "SUPPLEMENTAL-CODECS", vnt.SupplementalCodecs)
	}
	if vnt.Resolution != "" {
		writeUnQuoted(buf, "RESOLUTION", vnt.Resolution)
	}
	if vnt.FrameRate != 0 {
		writeFloat(buf, "FRAME-RATE", vnt.FrameRate)
	}
	if vnt.HDCPLevel != "" {
		writeUnQuoted(buf, "HDCP-LEVEL", vnt.HDCPLevel)
	}
	if vnt.AllowedCPC != "" {
		writeQuoted(buf, "ALLOWED-CPC", vnt.AllowedCPC)
	}
	if vnt.VideoRange != "" {
		writeUnQuoted(buf, "VIDEO-RANGE", vnt.VideoRange)
	}
	if vnt.ReqVideoLayout != "" {
		writeQuoted(buf, "REQ-VIDEO-LAYOUT", vnt.ReqVideoLayout)
	}
	if vnt.StableVariantId != "" {
		writeQuoted(buf, "STABLE-VARIANT-ID", vnt.StableVariantId)
	}
	if vnt.Audio != "" {
		writeQuoted(buf, "AUDIO", vnt.Audio)
	}
	if vnt.Video != "" {
		writeQuoted(buf, "VIDEO", vnt.Video)
	}
	if vnt.Subtitles != "" {
		writeQuoted(buf, "SUBTITLES", vnt.Subtitles)
	}
	if vnt.Captions != "" {
		if vnt.Captions == "NONE" {
			writeUnQuoted(buf, "CLOSED-CAPTIONS", vnt.Captions)
		} else {
			writeQuoted(buf, "CLOSED-CAPTIONS", vnt.Captions)
		}
	}
	if vnt.PathwayId != "" {
		writeQuoted(buf, "PATHWAY-ID", vnt.PathwayId)
	}
	if vnt.ProgramId != nil { // Removed in version 6
		buf.WriteString(",PROGRAM-ID=")
		buf.WriteString(strconv.FormatUint(uint64(*vnt.ProgramId), 10))
	}
	if vnt.Name != "" {
		writeQuoted(buf, "NAME", vnt.Name)
	}
	buf.WriteRune('\n')
}

func writeExtXIFrameStreamInf(buf *bytes.Buffer, vnt *Variant) {
	buf.WriteString("#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=")
	buf.WriteString(strconv.FormatUint(uint64(vnt.Bandwidth), 10))
	if vnt.AverageBandwidth != 0 {
		buf.WriteString(",AVERAGE-BANDWIDTH=")
		buf.WriteString(strconv.FormatUint(uint64(vnt.AverageBandwidth), 10))
	}
	if vnt.Score > 0 {
		writeFloat(buf, "SCORE", vnt.Score)
	}
	if vnt.Codecs != "" {
		writeQuoted(buf, "CODECS", vnt.Codecs)
	}
	if vnt.SupplementalCodecs != "" {
		writeQuoted(buf, "SUPPLEMENTAL-CODECS", vnt.SupplementalCodecs)
	}
	if vnt.Resolution != "" {
		writeUnQuoted(buf, "RESOLUTION", vnt.Resolution)
	}
	if vnt.HDCPLevel != "" {
		writeUnQuoted(buf, "HDCP-LEVEL", vnt.HDCPLevel)
	}
	if vnt.AllowedCPC != "" {
		writeQuoted(buf, "ALLOWED-CPC", vnt.AllowedCPC)
	}
	if vnt.VideoRange != "" {
		writeUnQuoted(buf, "VIDEO-RANGE", vnt.VideoRange)
	}
	if vnt.ReqVideoLayout != "" {
		writeQuoted(buf, "REQ-VIDEO-LAYOUT", vnt.ReqVideoLayout)
	}
	if vnt.StableVariantId != "" {
		writeQuoted(buf, "STABLE-VARIANT-ID", vnt.StableVariantId)
	}
	if vnt.Video != "" {
		writeQuoted(buf, "VIDEO", vnt.Video)
	}
	if vnt.PathwayId != "" {
		writeQuoted(buf, "PATHWAY-ID", vnt.PathwayId)
	}
	if vnt.ProgramId != nil { // Removed in version 6
		buf.WriteString(",PROGRAM-ID=")
		buf.WriteString(strconv.FormatInt(int64(*vnt.ProgramId), 10))
	}
	if vnt.Name != "" {
		writeQuoted(buf, "NAME", vnt.Name)
	}
	writeQuoted(buf, "URI", vnt.URI) // Mandatory
	buf.WriteRune('\n')
}

// writeDateRange writes an EXT-X-DATERANGE tag line including \n to the buffer.
func writeDateRange(buf *bytes.Buffer, dr *DateRange) {
	buf.WriteString(`#EXT-X-DATERANGE:ID="`)
	buf.WriteString(dr.ID)
	buf.WriteRune('"')
	if dr.Class != "" {
		writeQuoted(buf, "CLASS", dr.Class)
	}
	str := dr.StartDate.Format(DATETIME)
	writeQuoted(buf, "START-DATE", str)
	if dr.Cue != "" {
		writeUnQuoted(buf, "CUE", dr.Cue)
	}
	if dr.EndDate != nil {
		str = dr.EndDate.Format(DATETIME)
		writeQuoted(buf, "END-DATE", str)
	}
	if dr.Duration != nil {
		writeFloat(buf, "DURATION", *dr.Duration)
	}
	if dr.PlannedDuration != nil {
		writeFloat(buf, "PLANNED-DURATION", *dr.PlannedDuration)
	}
	if dr.SCTE35Cmd != "" {
		writeUnQuoted(buf, "SCTE35-CMD", dr.SCTE35Cmd)
	}
	if dr.SCTE35Out != "" {
		writeUnQuoted(buf, "SCTE35-OUT", dr.SCTE35Out)
	}
	if dr.SCTE35In != "" {
		writeUnQuoted(buf, "SCTE35-IN", dr.SCTE35In)
	}
	if dr.EndOnNext {
		buf.WriteString(",END-ON-NEXT=YES")
	}
	for _, xa := range dr.XAttrs {
		writeUnQuoted(buf, xa.Key, xa.Val)
	}
	buf.WriteRune('\n')
}

// writeQuoted writes a quoted key-value pair to the buffer preceded by a comma.
func writeQuoted(buf *bytes.Buffer, key, value string) {
	buf.WriteRune(',')
	buf.WriteString(key)
	buf.WriteString(`="`)
	buf.WriteString(value)
	buf.WriteRune('"')
}

func writeUnQuoted(buf *bytes.Buffer, key, value string) {
	buf.WriteRune(',')
	buf.WriteString(key)
	buf.WriteString(`=`)
	buf.WriteString(value)
}

// writeUint writes a key-value pair to the buffer preceded by a comma.
func writeUint(buf *bytes.Buffer, key string, value uint) {
	buf.WriteRune(',')
	buf.WriteString(key)
	buf.WriteRune('=')
	buf.WriteString(strconv.FormatUint(uint64(value), 10))
}

func writeFloat(buf *bytes.Buffer, key string, value float64) {
	buf.WriteRune(',')
	buf.WriteString(key)
	buf.WriteRune('=')
	buf.WriteString(strconv.FormatFloat(value, 'f', 3, 64))
}

func writeYESorNO(buf *bytes.Buffer, b bool) {
	if b {
		buf.WriteString("YES")
	} else {
		buf.WriteString("NO")
	}
}

// SetCustomTag sets the provided tag on the master playlist for its TagName
func (p *MasterPlaylist) SetCustomTag(tag CustomTag) {
	if p.Custom == nil {
		p.Custom = make(CustomMap)
	}

	p.Custom[tag.TagName()] = tag
}

// Version returns the current playlist version number
func (p *MasterPlaylist) Version() uint8 {
	return p.ver
}

// SetVersion sets the playlist version number, note the version maybe changed
// automatically by other Set methods.
func (p *MasterPlaylist) SetVersion(ver uint8) {
	p.ver = ver
}

// IndependentSegments returns true if all media samples in a segment can be
// decoded without information from other buf.
func (p *MasterPlaylist) IndependentSegments() bool {
	return p.independentSegments
}

// SetIndependentSegments sets the master playlist #EXT-X--INDEPENDENT-SEGMENTS tag.
func (p *MasterPlaylist) SetIndependentSegments(b bool) {
	p.independentSegments = b
}

// String provides the playlist fulfilling the Stringer interface.
func (p *MasterPlaylist) String() string {
	return p.Encode().String()
}

// GetAllAlternatives returns all alternative renditions sorted by
// groupID, type, name, and language.
func (p *MasterPlaylist) GetAllAlternatives() []*Alternative {
	added := make(map[string]*Alternative)

	for _, v := range p.Variants {
		for _, alt := range v.Alternatives {
			key := fmt.Sprintf("%s-%s-%s-%s", alt.GroupId, alt.Type, alt.Name, alt.Language)
			if _, ok := added[key]; !ok {
				added[key] = alt
			}
		}
	}
	alts := make([]*Alternative, 0, len(added))
	keys := make([]string, 0, len(added))
	for k := range added {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		alts = append(alts, added[key])
	}
	return alts
}

// NewMediaPlaylist creates a new media playlist structure.
// Winsize defines live window for playlist generation. Set to zero for VOD or EVENT
// playlists.  Capacity is the total size of the backing segment list..
// For VOD playlists, call Close() after the last segment is added.
func NewMediaPlaylist(winsize uint, capacity uint) (*MediaPlaylist, error) {
	p := new(MediaPlaylist)
	p.ver = minVer
	p.capacity = capacity
	if err := p.SetWinSize(winsize); err != nil {
		return nil, err
	}
	p.Segments = make([]*MediaSegment, capacity)
	return p, nil
}

// last returns the previously written segment's index
func (p *MediaPlaylist) last() uint {
	if p.tail == 0 {
		return p.capacity - 1
	}
	return p.tail - 1
}

func (p *MediaPlaylist) SetIndependentSegments(b bool) {
	p.independentSegments = b
}

func (p *MediaPlaylist) IndependentSegments() bool {
	return p.independentSegments
}

// Remove current segment from the head of chunk slice form a media playlist. Useful for sliding playlists.
// This operation resets playlist cache.
func (p *MediaPlaylist) Remove() (err error) {
	if p.count == 0 {
		return ErrPlaylistEmpty
	}
	p.head = (p.head + 1) % p.capacity
	p.count--
	if !p.Closed {
		p.SeqNo++
	}
	p.buf.Reset()
	return nil
}

// Append general chunk to the tail of chunk slice for a media playlist.
// This operation resets playlist cache.
func (p *MediaPlaylist) Append(uri string, duration float64, title string) error {
	seg := new(MediaSegment)
	seg.URI = uri
	seg.Duration = duration
	seg.Title = title
	return p.AppendSegment(seg)
}

// AppendSegment appends a MediaSegment to the tail of chunk slice for
// a media playlist.  This operation resets playlist cache.
func (p *MediaPlaylist) AppendSegment(seg *MediaSegment) error {
	if p.head == p.tail && p.count > 0 {
		return ErrPlaylistFull
	}
	seg.SeqId = p.SeqNo
	if p.count > 0 {
		seg.SeqId = p.Segments[(p.capacity+p.tail-1)%p.capacity].SeqId + 1
	}
	p.Segments[p.tail] = seg
	p.tail = (p.tail + 1) % p.capacity
	p.count++
	if !p.targetDurLocked {
		p.TargetDuration = calcNewTargetDuration(seg.Duration, p.ver, p.TargetDuration)
	}
	if seg.SCTE != nil {
		p.scte35Syntax = seg.SCTE.Syntax
	}
	if len(seg.SCTE35DateRanges) > 0 {
		p.scte35Syntax = SCTE35_DATERANGE
	}
	p.buf.Reset()
	return nil
}

// Slide combines two operations: first it removes one chunk from
// the head of chunk slice and move pointer to next chunk. Secondly it
// appends one chunk to the tail of chunk slice. Useful for sliding
// playlists.  This operation resets the cache.
func (p *MediaPlaylist) Slide(uri string, duration float64, title string) {
	if !p.Closed && p.count >= p.winsize {
		_ = p.Remove()
	}
	_ = p.Append(uri, duration, title)
}

// ResetCache resets playlist cache (internal buffer).
// Next call to Encode() fills buffer/cache again.
func (p *MediaPlaylist) ResetCache() {
	p.buf.Reset()
}

// Encode generates output and returns a pointer to an internal buffer.
// If winsize > 0, encoded the last `winsize` segments, otherwise encode all segments.
// If already encoded, and not changed, the cached buffer will be returned.
// Don't change the buffer externally, e.g. by using the Write() method
// if you want to use the cached value. Instead use the String() or Bytes() methods.
func (p *MediaPlaylist) Encode() *bytes.Buffer {
	if p.buf.Len() > 0 {
		return &p.buf
	}

	p.buf.WriteString("#EXTM3U\n#EXT-X-VERSION:")
	p.buf.WriteString(strVer(p.ver))
	p.buf.WriteRune('\n')

	if p.IndependentSegments() {
		p.buf.WriteString("#EXT-X-INDEPENDENT-SEGMENTS\n")
	}

	// Write any custom master tags
	if p.Custom != nil {
		for _, v := range p.Custom {
			if customBuf := v.Encode(); customBuf != nil {
				p.buf.WriteString(customBuf.String())
				p.buf.WriteRune('\n')
			}
		}
	}

	if p.AllowCache != nil {
		p.buf.WriteString("#EXT-X-ALLOW-CACHE:")
		writeYESorNO(&p.buf, *p.AllowCache)
		p.buf.WriteRune('\n')
	}

	// default key before any segment
	if p.Key != nil {
		p.buf.WriteString("#EXT-X-KEY:")
		p.buf.WriteString("METHOD=")
		p.buf.WriteString(p.Key.Method)
		if p.Key.Method != "NONE" {
			p.buf.WriteString(",URI=\"")
			p.buf.WriteString(p.Key.URI)
			p.buf.WriteRune('"')
			if p.Key.IV != "" {
				p.buf.WriteString(",IV=")
				p.buf.WriteString(p.Key.IV)
			}
			if p.Key.Keyformat != "" {
				p.buf.WriteString(",KEYFORMAT=\"")
				p.buf.WriteString(p.Key.Keyformat)
				p.buf.WriteRune('"')
			}
			if p.Key.Keyformatversions != "" {
				p.buf.WriteString(",KEYFORMATVERSIONS=\"")
				p.buf.WriteString(p.Key.Keyformatversions)
				p.buf.WriteRune('"')
			}
		}
		p.buf.WriteRune('\n')
	}
	// default MAP before any segment
	if p.Map != nil {
		p.buf.WriteString("#EXT-X-MAP:")
		p.buf.WriteString("URI=\"")
		p.buf.WriteString(p.Map.URI)
		p.buf.WriteRune('"')
		if p.Map.Limit > 0 {
			p.buf.WriteString(",BYTERANGE=")
			p.buf.WriteString(strconv.FormatInt(p.Map.Limit, 10))
			p.buf.WriteRune('@')
			p.buf.WriteString(strconv.FormatInt(p.Map.Offset, 10))
		}
		p.buf.WriteRune('\n')
	}
	if p.MediaType > 0 {
		p.buf.WriteString("#EXT-X-PLAYLIST-TYPE:")
		switch p.MediaType {
		case EVENT:
			p.buf.WriteString("EVENT\n")
		case VOD:
			p.buf.WriteString("VOD\n")
		}
	}
	p.buf.WriteString("#EXT-X-MEDIA-SEQUENCE:")
	p.buf.WriteString(strconv.FormatUint(p.SeqNo, 10))
	p.buf.WriteRune('\n')
	p.buf.WriteString("#EXT-X-TARGETDURATION:")
	p.buf.WriteString(strconv.FormatInt(int64(p.TargetDuration), 10))
	p.buf.WriteRune('\n')
	if p.StartTime > 0.0 {
		p.buf.WriteString("#EXT-X-START:TIME-OFFSET=")
		p.buf.WriteString(strconv.FormatFloat(p.StartTime, 'f', -1, 64))
		if p.StartTimePrecise {
			p.buf.WriteString(",PRECISE=YES")
		}
		p.buf.WriteRune('\n')
	}
	if p.DiscontinuitySeq != 0 {
		p.buf.WriteString("#EXT-X-DISCONTINUITY-SEQUENCE:")
		p.buf.WriteString(strconv.FormatUint(uint64(p.DiscontinuitySeq), 10))
		p.buf.WriteRune('\n')
	}
	if p.Iframe {
		p.buf.WriteString("#EXT-X-I-FRAMES-ONLY\n")
	}

	var (
		seg           *MediaSegment
		durationCache = make(map[float64]string)
	)

	head := p.head
	count := p.count
	for i := uint(0); (i < p.winsize || p.winsize == 0) && count > 0; count-- {
		seg = p.Segments[head]
		head = (head + 1) % p.capacity
		if seg == nil { // protection from badly filled chunklists
			continue
		}
		if p.winsize > 0 { // skip for VOD playlists, where winsize = 0
			i++
		}
		if seg.SCTE != nil {
			switch seg.SCTE.Syntax {
			case SCTE35_67_2014:
				p.buf.WriteString("#EXT-SCTE35:")
				p.buf.WriteString("CUE=\"")
				p.buf.WriteString(seg.SCTE.Cue)
				p.buf.WriteRune('"')
				if seg.SCTE.ID != "" {
					p.buf.WriteString(",ID=\"")
					p.buf.WriteString(seg.SCTE.ID)
					p.buf.WriteRune('"')
				}
				if seg.SCTE.Time != 0 {
					p.buf.WriteString(",TIME=")
					p.buf.WriteString(strconv.FormatFloat(seg.SCTE.Time, 'f', -1, 64))
				}
				p.buf.WriteRune('\n')
			case SCTE35_OATCLS:
				switch seg.SCTE.CueType {
				case SCTE35Cue_Start:
					if seg.SCTE.Cue != "" {
						p.buf.WriteString("#EXT-OATCLS-SCTE35:")
						p.buf.WriteString(seg.SCTE.Cue)
						p.buf.WriteRune('\n')
					}
					p.buf.WriteString("#EXT-X-CUE-OUT:")
					p.buf.WriteString(strconv.FormatFloat(seg.SCTE.Time, 'f', -1, 64))
					p.buf.WriteRune('\n')
				case SCTE35Cue_Mid:
					p.buf.WriteString("#EXT-X-CUE-OUT-CONT:")
					p.buf.WriteString("ElapsedTime=")
					p.buf.WriteString(strconv.FormatFloat(seg.SCTE.Elapsed, 'f', -1, 64))
					p.buf.WriteString(",Duration=")
					p.buf.WriteString(strconv.FormatFloat(seg.SCTE.Time, 'f', -1, 64))
					p.buf.WriteString(",SCTE35=")
					p.buf.WriteString(seg.SCTE.Cue)
					p.buf.WriteRune('\n')
				case SCTE35Cue_End:
					p.buf.WriteString("#EXT-X-CUE-IN")
					p.buf.WriteRune('\n')
				}
			}
		}
		for i := range seg.SCTE35DateRanges {
			writeDateRange(&p.buf, seg.SCTE35DateRanges[i])
		}
		// check for key change
		if seg.Key != nil && (p.Key == nil || *seg.Key != *p.Key) {
			p.buf.WriteString("#EXT-X-KEY:")
			p.buf.WriteString("METHOD=")
			p.buf.WriteString(seg.Key.Method)
			if seg.Key.Method != "NONE" {
				p.buf.WriteString(",URI=\"")
				p.buf.WriteString(seg.Key.URI)
				p.buf.WriteRune('"')
				if seg.Key.IV != "" {
					p.buf.WriteString(",IV=")
					p.buf.WriteString(seg.Key.IV)
				}
				if seg.Key.Keyformat != "" {
					p.buf.WriteString(",KEYFORMAT=\"")
					p.buf.WriteString(seg.Key.Keyformat)
					p.buf.WriteRune('"')
				}
				if seg.Key.Keyformatversions != "" {
					p.buf.WriteString(",KEYFORMATVERSIONS=\"")
					p.buf.WriteString(seg.Key.Keyformatversions)
					p.buf.WriteRune('"')
				}
			}
			p.buf.WriteRune('\n')
		}
		if seg.Discontinuity {
			p.buf.WriteString("#EXT-X-DISCONTINUITY\n")
		}
		// ignore segment Map if default playlist Map is present
		if p.Map == nil && seg.Map != nil {
			p.buf.WriteString("#EXT-X-MAP:")
			p.buf.WriteString("URI=\"")
			p.buf.WriteString(seg.Map.URI)
			p.buf.WriteRune('"')
			if seg.Map.Limit > 0 {
				p.buf.WriteString(",BYTERANGE=")
				p.buf.WriteString(strconv.FormatInt(seg.Map.Limit, 10))
				p.buf.WriteRune('@')
				p.buf.WriteString(strconv.FormatInt(seg.Map.Offset, 10))
			}
			p.buf.WriteRune('\n')
		}
		if !seg.ProgramDateTime.IsZero() {
			p.buf.WriteString("#EXT-X-PROGRAM-DATE-TIME:")
			p.buf.WriteString(seg.ProgramDateTime.Format(DATETIME))
			p.buf.WriteRune('\n')
		}
		if seg.Limit > 0 {
			p.buf.WriteString("#EXT-X-BYTERANGE:")
			p.buf.WriteString(strconv.FormatInt(seg.Limit, 10))
			p.buf.WriteRune('@')
			p.buf.WriteString(strconv.FormatInt(seg.Offset, 10))
			p.buf.WriteRune('\n')
		}

		// Add Custom Segment Tags here
		if seg.Custom != nil {
			for _, v := range seg.Custom {
				if customBuf := v.Encode(); customBuf != nil {
					p.buf.WriteString(customBuf.String())
					p.buf.WriteRune('\n')
				}
			}
		}

		p.buf.WriteString("#EXTINF:")
		if str, ok := durationCache[seg.Duration]; ok {
			p.buf.WriteString(str)
		} else {
			durationCache[seg.Duration] = strconv.FormatFloat(seg.Duration, 'f', 3, 32)
			p.buf.WriteString(durationCache[seg.Duration])
		}
		p.buf.WriteRune(',')
		p.buf.WriteString(seg.Title)
		p.buf.WriteRune('\n')
		p.buf.WriteString(seg.URI)
		if p.Args != "" {
			p.buf.WriteRune('?')
			p.buf.WriteString(p.Args)
		}
		p.buf.WriteRune('\n')
	}
	if p.Closed {
		p.buf.WriteString("#EXT-X-ENDLIST\n")
	}
	for _, dr := range p.DateRanges {
		writeDateRange(&p.buf, dr)
	}
	return &p.buf
}

// String provides the playlist fulfilling the Stringer interface.
func (p *MediaPlaylist) String() string {
	return p.Encode().String()
}

// Count tells us the number of items that are currently in the media playlist.
func (p *MediaPlaylist) Count() uint {
	return p.count
}

// Close sliding playlist and by setting the EXT-X-ENDLIST tag and setting the Closed flag.
func (p *MediaPlaylist) Close() {
	if p.buf.Len() > 0 {
		p.buf.WriteString("#EXT-X-ENDLIST\n")
	}
	p.Closed = true
}

// CalculateTargetDuration calculates the target duration for the playlist.
// For HLS v5 and earlier, it is the maximum segment duration as rounded up.
// For HLS v6 and later, it is the maximum segment duration as rounded to the nearest integer.
// It is not allowed to change when the playlist is updated.
func (p *MediaPlaylist) CalculateTargetDuration(hlsVer uint8) uint {
	if p.count == 0 {
		return 0
	}
	var max float64
	if p.tail >= p.head {
		for i := p.head; i < p.tail; i++ {
			if p.Segments[i].Duration > max {
				max = p.Segments[i].Duration
			}
		}
	} else {
		for i := p.head; i < p.capacity; i++ {
			if p.Segments[i].Duration > max {
				max = p.Segments[i].Duration
			}
		}
		for i := uint(0); i < p.tail; i++ {
			if p.Segments[i].Duration > max {
				max = p.Segments[i].Duration
			}
		}
	}
	return calcNewTargetDuration(max, hlsVer, 0)
}

// calcNewTargetDuration calculates a new target duration based on a segment duration.
func calcNewTargetDuration(segDur float64, hlsVer uint8, oldTargetDuration uint) uint {
	var new uint
	if hlsVer < 6 {
		new = uint(math.Ceil(segDur))
	} else {
		new = uint(math.Round(segDur))
	}
	if new > oldTargetDuration {
		return new
	}
	return oldTargetDuration
}

// SetTargetDuration sets the target duration for the playlist and stops automatic calculation.
// Since the target duration is not allowed to change, it is locked after the first call.
func (p *MediaPlaylist) SetTargetDuration(duration uint) {
	p.TargetDuration = duration
	p.targetDurLocked = true
}

// SetDefaultKey sets encryption key to appear before segments in the media playlist.
func (p *MediaPlaylist) SetDefaultKey(method, uri, iv, keyformat, keyformatversions string) error {
	if keyformat != "" || keyformatversions != "" {
		updateVersion(&p.ver, 5) // [Protocol Version Compatibility]
	}
	p.Key = &Key{method, uri, iv, keyformat, keyformatversions}
	return nil
}

// SetDefaultMap sets default Media Initialization Section (EXT-X-MAP)
// at start of playlist. May be overridden by individual segments.
func (p *MediaPlaylist) SetDefaultMap(uri string, limit, offset int64) {
	updateVersion(&p.ver, 5) // [Protocol Version Compatibility]
	p.Map = &Map{uri, limit, offset}
}

// SetIframeOnly marks medialist of only I-frames (Intra frames).
func (p *MediaPlaylist) SetIframeOnly() {
	updateVersion(&p.ver, 4) // [Protocol Version Compatibility]
	p.Iframe = true
}

// SetKey sets encryption key for the current (and following) segment of media playlist
func (p *MediaPlaylist) SetKey(method, uri, iv, keyformat, keyformatversions string) error {
	if p.count == 0 {
		return ErrPlaylistEmpty
	}

	if keyformat != "" || keyformatversions != "" {
		updateVersion(&p.ver, 5) // [Protocol Version Compatibility]
	}

	p.Segments[p.last()].Key = &Key{method, uri, iv, keyformat, keyformatversions}
	return nil
}

// SetMap sets map for the currently last segment of media playlist.
func (p *MediaPlaylist) SetMap(uri string, limit, offset int64) error {
	if p.count == 0 {
		return ErrPlaylistEmpty
	}
	updateVersion(&p.ver, 5) // [Protocol Version Compatibility]
	p.Segments[p.last()].Map = &Map{uri, limit, offset}
	return nil
}

// SetRange sets byte range limit and offset for the currently last media segment.
func (p *MediaPlaylist) SetRange(limit, offset int64) error {
	if p.count == 0 {
		return ErrPlaylistEmpty
	}
	updateVersion(&p.ver, 4) // [Protocol Version Compatibility]
	p.Segments[p.last()].Limit = limit
	p.Segments[p.last()].Offset = offset
	return nil
}

// SetSCTE sets the SCTE cue format for the currently last media segment.
//
// Deprecated: Use SetSCTE35 instead.
func (p *MediaPlaylist) SetSCTE(cue string, id string, time float64) error {
	return p.SetSCTE35(&SCTE{Syntax: SCTE35_67_2014, Cue: cue, ID: id, Time: time})
}

// SetSCTE35 sets the SCTE cue format for the currently last media segment.
func (p *MediaPlaylist) SetSCTE35(scte35 *SCTE) error {
	if p.count == 0 {
		return ErrPlaylistEmpty
	}
	p.Segments[p.last()].SCTE = scte35
	return nil
}

// SetDiscontinuity sets discontinuity flag for the currently last media segment.
// EXT-X-DISCONTINUITY indicates an encoding discontinuity
// between the media segment that follows it and the one that preceded it.
func (p *MediaPlaylist) SetDiscontinuity() error {
	if p.count == 0 {
		return ErrPlaylistEmpty
	}
	p.Segments[p.last()].Discontinuity = true
	return nil
}

// SetProgramDateTime sets program date and time for the currently last media segment.
// EXT-X-PROGRAM-DATE-TIME tag associates the first sample of
// a media segment with an absolute date and/or time. It applies only
// to the current media segment.  Date/time format is
// YYYY-MM-DDThh:mm:ssZ (ISO8601) and includes time zone.
func (p *MediaPlaylist) SetProgramDateTime(value time.Time) error {
	if p.count == 0 {
		return ErrPlaylistEmpty
	}
	p.Segments[p.last()].ProgramDateTime = value
	return nil
}

// SetCustomTag sets the provided tag on the media playlist for its TagName.
func (p *MediaPlaylist) SetCustomTag(tag CustomTag) {
	if p.Custom == nil {
		p.Custom = make(CustomMap)
	}

	p.Custom[tag.TagName()] = tag
}

// SetCustomSegmentTag sets the provided tag on the current media segment for its TagName.
func (p *MediaPlaylist) SetCustomSegmentTag(tag CustomTag) error {
	if p.count == 0 {
		return ErrPlaylistEmpty
	}

	last := p.Segments[p.last()]

	if last.Custom == nil {
		last.Custom = make(CustomMap)
	}

	last.Custom[tag.TagName()] = tag

	return nil
}

// Version returns playlist's version number.
func (p *MediaPlaylist) Version() uint8 {
	return p.ver
}

// SetVersion sets the playlist version number, note the version
// have increased automatically by other Set methods.
func (p *MediaPlaylist) SetVersion(ver uint8) {
	p.ver = ver
}

// WinSize returns the playlist's window size.
func (p *MediaPlaylist) WinSize() uint {
	return p.winsize
}

// SetWinSize overwrites the playlist's window size.
func (p *MediaPlaylist) SetWinSize(winsize uint) error {
	if winsize > p.capacity {
		return fmt.Errorf("capacity=%d < winsize=%d: %w", p.capacity, winsize, ErrWinSizeTooSmall)
	}
	p.winsize = winsize
	return nil
}

// GetAllSegments could get all segments currently added to playlist.
// Winsize is ignored.
func (p *MediaPlaylist) GetAllSegments() []*MediaSegment {
	if p.count == 0 {
		return nil
	}
	buf := make([]*MediaSegment, 0, p.count)
	if p.head < p.tail {
		for i := p.head; i < p.tail; i++ {
			buf = append(buf, p.Segments[i])
		}
		return buf
	}
	for i := uint(0); i < p.tail; i++ {
		buf = append(buf, p.Segments[i])
	}
	for i := p.head; i < p.capacity; i++ {
		buf = append(buf, p.Segments[i])
	}
	return buf
}

/*
[Protocol Version Compatibility]: https://datatracker.ietf.org/doc/html/draft-pantos-hls-rfc8216bis-16#section-8
*/

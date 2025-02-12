package m3u8

import "fmt"

type VersionMismatch struct {
	ActualVersion   uint8
	ExpectedVersion uint8
	Description     string
}

// Error implements error.
func (m VersionMismatch) Error() string {
	return fmt.Sprintf("Playlist version mismatch %s. Actual version is %d, expected version is %d", m.Description, m.ActualVersion, m.ExpectedVersion)
}

type VersionMatchingRule interface {
	Validate() (bool, VersionMismatch)
}

type DefaultMatchingRule struct{}

func (d *DefaultMatchingRule) Validate() (bool, VersionMismatch) {
	return true, VersionMismatch{}
}

type ValidIVInEXTXKey struct {
	ActualVersion uint8
	IV            string
}

func (v ValidIVInEXTXKey) Validate() (bool, VersionMismatch) {
	if v.IV == "" || v.ActualVersion >= 2 {
		return true, VersionMismatch{}
	}

	return false, VersionMismatch{
		ActualVersion:   v.ActualVersion,
		ExpectedVersion: 2,
		Description:     "Protocol version needs to be at least 2 if you have IV in EXT-X-KEY.",
	}
}

type FloatPointDuration struct {
	ActualVersion uint8
	duration      string
}

func (f FloatPointDuration) Validate() (bool, VersionMismatch) {
	if isStringInteger(f.duration) || f.ActualVersion >= 3 {
		return true, VersionMismatch{}
	}

	return false, VersionMismatch{
		ActualVersion:   f.ActualVersion,
		ExpectedVersion: 3,
		Description:     "Protocol version needs to be at least 3 if you have floating point duration.",
	}
}

type ContainsByteRangeOrIFrameOnly struct {
	ActualVersion uint8
}

func (c ContainsByteRangeOrIFrameOnly) Validate() (bool, VersionMismatch) {
	if c.ActualVersion >= 4 {
		return true, VersionMismatch{}
	}

	return false, VersionMismatch{
		ActualVersion:   c.ActualVersion,
		ExpectedVersion: 4,
		Description:     "Protocol version needs to be at least 4 if you have EXT-X-BYTERANGE or EXT-X-I-FRAMES-ONLY tag",
	}
}

package m3u8

import "fmt"

type VersionMismatchError struct {
	ActualVersion   uint8
	ExpectedVersion uint8
	Description     string
}

// Error implements error.
func (m VersionMismatchError) Error() string {
	return fmt.Sprintf("Playlist version mismatch %s. Actual version is %d, expected version is %d", m.Description, m.ActualVersion, m.ExpectedVersion)
}

type VersionMatchingRule interface {
	Validate() (bool, VersionMismatchError)
}

type DefaultMatchingRule struct{}

func (d *DefaultMatchingRule) Validate() (bool, VersionMismatchError) {
	return true, VersionMismatchError{}
}

type ValidIVInEXTXKEY struct {
	ActualVersion uint8
	IV            string
}

func (v ValidIVInEXTXKEY) Validate() (bool, VersionMismatchError) {
	if v.IV == "" || v.ActualVersion >= 2 {
		return true, VersionMismatchError{}
	}

	return false, VersionMismatchError{
		ActualVersion:   v.ActualVersion,
		ExpectedVersion: 2,
		Description:     "Protocol version needs to be at least 2 if you have IV in EXT-X-KEY.",
	}
}

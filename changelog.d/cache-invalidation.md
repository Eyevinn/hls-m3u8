### Fixed

- `Encode` and `String` no longer return stale output after `Set*` and `Append*` methods
  modify an already encoded playlist, and repeated `Close` calls write one `EXT-X-ENDLIST` (PR #94)

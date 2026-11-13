### Fixed

- `Decode`, `DecodeFrom` and `DecodeWith` now release the unused master or media playlist to its pool,
  so decoding a master playlist uses about a third of the memory (issue #97)

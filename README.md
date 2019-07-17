# unique

This package implements generation and serialization of [ULID](https://github.com/ulid/spec)-based
unique identifiers.

## FAQ

### Why ULID?

1. The [ULID specification](https://github.com/ulid/spec) clearly outlines ULID's
   advantages over traditional UUID/GUID.

1. We generate IDs in distributed environments, some of which don't have network
   access. This rules out centralized IDs and we must rely on low probability of
   collision.

1. ULIDs are k-sortable in both binary and string form, case-sensitive,
   URL-friendly, and reasonably compact.

### Why not just use [oklog/ulid](https://github.com/oklog/ulid) directly?

We use this under the hood. The following considerations lead us to maintaining
our own wrapper or fork.

1. We prefer controlling a stable interface for something as fundamental as IDs.

1. We want to tweak the interface to better suit our code bases. Specifically,
   - We prefer text marshalling to default to strict validation. We're willing
     to spend a few extra cycles on conversion.
   - `ulid.ULID` requires conversion from the 48-bit timestamp to `time.Time`.
     The latter is far more common in our code.

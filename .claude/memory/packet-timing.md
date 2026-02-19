# Packet Timing Design Notes

## Fields

| Field | Type | Meaning |
|---|---|---|
| `DTS` | `time.Duration` | Decode timestamp — when the decoder should process this packet |
| `PTSOffset` | `time.Duration` | PTS − DTS offset; zero for streams without B-frames |
| `Duration` | `time.Duration` | Packet duration |
| `WallClockTime` | `time.Time` | Real wall-clock time; zero value = unset |

## Methods

- `PTS()` — returns `DTS + PTSOffset`; the only correct way to get PTS; never compute inline
- `HasWallClockTime()` — returns `!WallClockTime.IsZero()`

## Naming rationale

- `DTS` (was `Time`): `Time` was ambiguous — could be read as PTS or DTS. `DTS` is explicit.
- `PTSOffset` (was `CompositionTime`): `CompositionTime` sounded like an absolute timestamp.
  `PTSOffset` is self-documenting: it is the offset added to DTS to obtain PTS.
  (ISO 14496-12 calls this `composition_time_offset` in the `ctts` box; `PTSOffset` is equivalent
  and more readable for Go developers unfamiliar with container spec terminology.)
- `DTS()` method removed: redundant once the field itself is named `DTS`.

## Why WallClockTime was added

- Live RTSP/ONVIF streams carry NTP timestamps (RTCP Sender Reports map RTP timestamps to wall clock).
- Needed to synchronise multiple streams (e.g. audio + video) with different relative start times.
- Needed to correlate captured packets to real-world events (e.g. motion detection timestamp).
- Zero value (Go default) = "not set", so existing code that doesn't fill it is unaffected.

## String() behaviour

- When `PTSOffset == 0`: logs `DTS=` only (common case — no B-frames)
- When `PTSOffset != 0`: logs both `DTS=` and `PTS=` (B-frame streams)

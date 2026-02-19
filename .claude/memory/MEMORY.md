# avsdk Project Memory

## Packet timing model (av/packet.go)

`Packet.DTS time.Duration` — decode timestamp (when decoder processes this packet).
`Packet.PTSOffset time.Duration` — PTS−DTS offset; non-zero only for B-frames (H.264/H.265).
`Packet.PTS()` — returns `DTS + PTSOffset`; always use this method, never compute manually.
`Packet.WallClockTime time.Time` — real wall-clock capture/arrival time (e.g. NTP from RTSP/ONVIF).
Zero value means unset; check with `HasWallClockTime()`.

See: [packet-timing.md](packet-timing.md) for design rationale.

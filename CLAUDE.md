# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

Install dev tools (run once):
```bash
make prerequisite
```

Build and test (full pipeline):
```bash
make all          # prerequisite + prepare + test + testsum + coverage + check
make test         # go test ./...
make testsum      # gotestsum (enhanced output)
make coverage     # generates coverage.html and coverage.txt
```

Run a single test:
```bash
go test ./codec/h264parser -run TestParseSPS
```

Format and lint:
```bash
make prepare      # gofumpt -l -w .
make check        # golangci-lint run --fix
```

Update dependencies:
```bash
make update       # go get -u ./... && go mod tidy
```

## Architecture

**avsdk** is a Go SDK for audio/video codec parsing and streaming. Module path: `github.com/vtpl1/avsdk`.

### Package layout

- **`/av`** — Core types shared across the SDK:
  - `av.go` — `CodecType` enum, `Stream` struct, `CodecData` / `VideoCodecData` / `AudioCodecData` interfaces
  - `packet.go` — `Packet` struct (see below)
  - `audtioframe.go` — `AudioFrame` with `SampleFormat` and `ChannelLayout`
  - `nalutype.go` — NALU type constants for H.264 and H.265
  - `demuxer.go` — `Demuxer`, `DemuxCloser`, `Pauser`, `TimeSeeker` interfaces
  - `muxer.go` — `Muxer`, `MuxCloser`, `PacketWriter`, `CodecChanger` interfaces

- **`/av/avutil`** — `CopyPackets`, `CopyFile`, `Equal`, handler registry (with tests)

- **`/codec`** — Codec-specific parsers, each in its own sub-package:
  - `/h264parser` — SPS/PPS parsing, Annex-B ↔ AVCC conversion
  - `/h265parser` — VPS/SPS/PPS parsing for HEVC
  - `/aacparser` — AAC audio decoder config parsing
  - `/pcm` — PCM variants: PCMU, PCMA, FLAC, Speex
  - `/mjpeg` — MJPEG parser
  - `/parser` — Generic NALU parser (auto-detects Annex-B vs AVCC format)
  - `sdp.go` — SDP parsing to extract codec info from streams
  - `opus_codec.go` — Opus codec support

- **`/utils/bits`** — Low-level bit manipulation: Golomb reader, bit reader, PIO (packet I/O)

### Packet

`Packet` is the central data-passing type:

```go
type Packet struct {
    KeyFrame        bool          // sync/random-access point
    IsDiscontinuity bool          // DTS gap — receivers must reinitialise timing
    IsParamSetNALU  bool          // contains SPS/PPS/VPS (no display output)
    Idx             uint16        // stream index; matches Stream.Idx
    DTS             time.Duration // decode timestamp
    PTSOffset       time.Duration // PTS = DTS + PTSOffset; non-zero only for B-frames
    Duration        time.Duration // packet duration
    WallClockTime   time.Time     // NTP/wall-clock anchor; zero means unset
    Data            []byte        // compressed payload; empty for codec-change notifications
    CodecType       CodecType
    FrameID         int64
    Extra           any
    NewCodecs       []Stream      // non-nil: mid-stream codec change for listed streams only
}
```

Key methods: `PTS()`, `HasWallClockTime()`, `IsKeyFrame()`, `IsAudio()`, `IsVideo()`.

### Stream

`Stream` pairs a stream index with its codec. Used by `GetCodecs`, `WriteHeader`, and `Packet.NewCodecs`:

```go
type Stream struct {
    Idx   uint16    // authoritative identifier — never infer from slice position
    Codec CodecData
}
```

### Demuxer / Muxer pipeline

Minimal remux loop:
```go
streams, _ := demux.GetCodecs(ctx)
mux.WriteHeader(ctx, streams)
for {
    pkt, err := demux.ReadPacket(ctx)
    if errors.Is(err, io.EOF) { break }
    if pkt.NewCodecs != nil {
        if cc, ok := mux.(av.CodecChanger); ok {
            cc.WriteCodecChange(ctx, pkt.NewCodecs)
        }
    }
    mux.WritePacket(ctx, pkt)
}
mux.WriteTrailer(ctx)
```

`avutil.CopyFile` and `avutil.CopyPackets` implement this pattern.

Optional demuxer capabilities are accessed by type assertion — not embedded composites:
```go
if p, ok := dmx.(av.Pauser);     ok { p.Pause(ctx) / p.Resume(ctx) }
if s, ok := dmx.(av.TimeSeeker); ok { s.SeekToTime(ctx, 30*time.Second) }
```

### Codec interfaces

- `VideoCodecData` — `Width()`, `Height()`, `TimeScale() uint32` (90000 for H.264/H.265, used for RTP clock and fMP4 `mdhd`)
- `AudioCodecData` — `SampleRate()`, `SampleFormat()`, `ChannelLayout()`, `PacketDuration()`
- Codec parsers are stateless functions where possible; codec data structs implement these interfaces.
- NALU parsing supports both Annex-B (`00 00 01` start codes) and AVCC (length-prefixed) — `/codec/parser` detects format automatically.

## Linting

`.golangci.yml` enables 70+ linters. Key limits:
- Line length: 160 characters (rulers in VS Code at 120 and 160)
- Function length: 110 lines / 70 statements
- Cyclomatic complexity: 30 (cyclop), cognitive complexity: 50 (gocognit)

The project uses `gofumpt` (stricter than `gofmt`) for formatting.

## Version management

Version is tracked in `.bumpversion.cfg` (currently `0.21.0`). Bump with `bump2version` (available in the devcontainer).

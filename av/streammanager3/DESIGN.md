# streammanager3 — Design Document

## Purpose

`streammanager3` implements `av.StreamManager`: a concurrent, lifecycle-managed
fanout hub that connects one AV source (a *producer*) to one or more downstream
sinks (*consumers*). It is the primary entry point for applications that need to
multiplex a live stream — RTSP, a camera feed, a file, anything wrapped in
`av.DemuxCloser` — to multiple simultaneous recipients.

---

## Three-level object hierarchy

```
StreamManager
│  map[producerID]*Producer          (one entry per live source)
│
└── Producer
    │  map[consumerID]*consumer       (one entry per attached sink)
    │  av.DemuxCloser                 (created lazily; torn down on last consumer)
    │  []av.Stream                    (kept current on NewCodecs)
    │  runLoop goroutine
    │
    └── consumer  (unexported)
           av.MuxCloser
           chan av.Packet  (buffered, size 64)
           write goroutine
```

### StreamManager

Owns the `producerID → *Producer` registry. All registry mutations are
protected by `mu sync.RWMutex`.

Responsibilities:
- Create a `Producer` lazily when the first consumer for a `producerID` is added.
- Destroy a `Producer` (via its `removeMe` callback) when the last consumer is removed.
- Forward `Pause`/`Resume` to the correct producer.
- Cancel all producer contexts on `Stop()` and drain all `runLoop` goroutines in `WaitStop()`.

### Producer

Owns one demuxer and N consumers. Protected by `mu sync.RWMutex`.

Responsibilities:
- Open the demuxer and call `GetCodecs` when the first consumer arrives.
- Track the current codec list (`streams`), updating it whenever a
  `pkt.NewCodecs != nil` packet is received so late-joining consumers get
  correct stream info.
- Run `runLoop` — a single goroutine that reads from the demuxer and fans each
  packet out to all live consumers.
- Tear down all resources (cancel context, wait for `runLoop`, close demuxer)
  when the last consumer is removed.

### consumer (unexported)

Owns one muxer and one write goroutine. Each consumer has an independent
`pktChan chan av.Packet` (buffered, capacity 64) that decouples the producer's
read loop from the muxer's write speed.

Responsibilities:
- Call `WriteHeader` with the current stream list on start.
- Drain `pktChan` in a dedicated goroutine, calling `WriteCodecChange` (if the
  muxer implements `av.CodecChanger`) and `WritePacket` for every packet.
- On a **write error**: forward the error to `errChan`, then launch an `onDead`
  goroutine that calls `Producer.RemoveConsumer` to trigger automatic removal
  and teardown. The producer stops fanning packets to the dead sink immediately
  (the consumer is deleted from `p.consumers` under `p.mu` before `pktChan` is
  closed, so `runLoop` can never send to a closed channel).
- On **normal stop** (`pktChan` closed by `stop()`): drain remaining buffered
  packets, then call `WriteTrailer`, `Close`, and `muxerRemover`. `onDead` is
  **not** called on the normal path.

---

## Data flow

```
demuxer.ReadPacket()
        │
        ▼
    runLoop (one goroutine per Producer)
        │
        │  p.mu.RLock
        ├──► consumer A: non-blocking send → pktChan (cap 64)
        ├──► consumer B: non-blocking send → pktChan (cap 64)
        └──► consumer C: non-blocking send → pktChan (cap 64)
        │  p.mu.RUnlock
        │
        ▼  (per consumer, independent goroutine)
    consumer.run()
        │
        ├── WriteCodecChange  (if pkt.NewCodecs != nil and mux implements CodecChanger)
        └── WritePacket
```

The non-blocking fanout (`select { case c.pktChan <- pkt: default: }`) means a
slow or stalled consumer never blocks the read loop or starves faster consumers.
Dropped packets are silent; the caller is expected to notice via the error
channel and remove the dead consumer.

---

## Lifecycle

### AddConsumer

```
StreamManager.AddConsumer
  ├── [loop] m.mu.Lock
  │     if no producer for producerID: create Producer, register makeProducerRemover
  │   m.mu.Unlock
  │
  ├── Producer.AddConsumer  (p.mu.Lock)
  │     if closing → return errProducerClosing  ──► loop back to top
  │     if first consumer:
  │       demuxerFactory() → GetCodecs() → store demuxer + streams
  │     muxerFactory()
  │     newConsumer()
  │       set c.onDead = func() { p.RemoveConsumer(Background, producerID, consumerID) }
  │     consumer.start() → WriteHeader + launch write goroutine
  │     store consumer in map
  │     if first consumer: p.wg.Add(1); go runLoop()
  │   p.mu.Unlock
  │
  └── return nil
```

If any step after the demuxer is opened fails on the first consumer, the
demuxer is closed and `p.demuxer` is set back to nil before returning the error.
If the producer was freshly created and `AddConsumer` ultimately fails, the
producer is deleted from the StreamManager map (identity-checked to avoid
evicting a replacement).

### RemoveConsumer

```
Producer.RemoveConsumer
  ├── p.mu.Lock
  │     delete consumer from map
  │     if now empty: p.closing = true
  │   p.mu.Unlock
  │
  ├── consumer.stop(ctx)
  │     close(pktChan)   ← write goroutine drains remaining packets, then exits
  │     wg.Wait()
  │     cancel()
  │     WriteTrailer + Close + muxerRemover
  │
  └── if was last consumer:
        p.cancel()        ← cancels runLoop's context
        p.wg.Wait()       ← waits for runLoop to exit
        demuxer.Close()
        removeMe()        ← identity-checked map deletion + demuxerRemover
```

### Stop

```
StreamManager.Stop
  ├── SignalStop: m.cancel()  ← all p.ctx (derived from m.ctx) are cancelled
  └── WaitStop:
        snapshot producers under m.mu.RLock
        for each Producer: p.wg.Wait()   ← blocks until runLoop exits
```

`Stop` does **not** remove consumers. It only cancels contexts and waits for
goroutines. The read loops exit naturally when their next `ReadPacket` call
returns `context.Canceled`. Applications that need a clean flush should call
`RemoveConsumer` for every consumer before calling `Stop`.

---

## Concurrency model

### Locks

| Lock | Scope | Protects |
|---|---|---|
| `StreamManager.mu` | RWMutex | `producers` map |
| `Producer.mu` | RWMutex | `consumers` map, `closing`, `demuxer`, `streams` |

There is no nested locking between the two levels. `StreamManager.mu` is never
held while calling into `Producer`, and `Producer.mu` is never held while
calling `StreamManager.mu`. This total ordering eliminates deadlock.

### Context hierarchy

```
parent ctx (caller-supplied)
  └── m.ctx  (StreamManager; cancelled by SignalStop)
        └── p.ctx  (Producer; cancelled when last consumer is removed, or by m.ctx)
              └── c.ctx  (consumer; cancelled in consumer.stop, or by p.ctx)
```

Each level can be independently cancelled. Consumer contexts are cancelled in
`stop()` after the write goroutine exits, primarily to propagate cancellation to
any `WritePacket` call already in flight.

### The add/remove race and the `closing` flag

Between `p.mu.Unlock()` (after marking the consumer map empty) and
`removeMe()` (which deletes the producer from the StreamManager map), a
concurrent `AddConsumer` could find the same `*Producer` in the map, see
`len(consumers)==0`, and try to re-initialise it — while `RemoveConsumer` is
still tearing it down.

**Fix:** `closing bool` (set under `p.mu` at the same time the map entry is
marked empty) causes `AddConsumer` to return the internal sentinel
`errProducerClosing`. `StreamManager.AddConsumer` catches that sentinel, removes
the stale map entry, and retries the whole operation, which creates a fresh
`Producer`.

### Identity-safe `removeMe`

`makeProducerRemover` closes over the specific `*Producer` pointer `p`. When
the remover runs, it checks `cur == p` before deleting from the map. This
prevents a closing producer from evicting a replacement producer that was
installed during the teardown window.

Without this check, the sequence below would silently orphan P2:

```
P1 closing → errProducerClosing returned to AddConsumer
AddConsumer installs P2 in map
P1.removeMe fires: delete(map, id)  ← would delete P2
```

With the check, P1's remover sees `cur == P2 ≠ P1` and skips the deletion.

### Packet fanout under RLock

`runLoop` holds `p.mu.RLock` only during the fanout loop. All sends are
non-blocking (`select { default: }`). Consumers that cannot keep up silently
drop packets; they are never removed from the map automatically. The caller
learns of problems through `errChan` and is responsible for calling
`RemoveConsumer`.

`p.streams` is the only field written by `runLoop` (under `p.mu.Lock`, not
RLock), and only when `pkt.NewCodecs != nil` — which is rare. There is no
lock upgrade; the write lock is acquired and released as a separate critical
section before the read lock for fanout.

---

## Error handling contract

| Error | Meaning | Expected caller action |
|---|---|---|
| `ErrProducerNotFound` | `RemoveConsumer` / `Pause` / `Resume` called with unknown producerID | Bug in caller; do not retry |
| `ErrConsumerNotFound` | `RemoveConsumer` called with unknown consumerID | Bug in caller; do not retry |
| Error from `demuxerFactory` | Source could not be opened | Surfaced synchronously from `AddConsumer`; caller may retry |
| Error sent to `errChan` (muxer write failure) | Consumer's write goroutine failed | Auto-removal runs; if the caller also calls `RemoveConsumer`, it will receive `ErrConsumerNotFound` — this is safe to ignore |
| Error sent to `errChan` (demuxer read failure, including `io.EOF`) | The producer's demuxer is done | All consumers received the error; call `RemoveConsumer` for each, or wait — if the producer has no remaining consumers it removes itself automatically |

Errors sent to `errChan` use a non-blocking send. If `errChan` is full, the
error is dropped and the write goroutine (for muxer errors) exits silently.
Size `errChan` generously (at least one slot per consumer) or drain it promptly.

---

## Invariants

1. `p.demuxer != nil` iff `len(p.consumers) > 0` and `!p.closing`.
2. `runLoop` is active iff `len(p.consumers) > 0` and `!p.closing`.
3. A `*Producer` in the StreamManager map always has `closing == false`.
4. A consumer's `pktChan` is closed exactly once, by `consumer.stop`.
5. `consumer.stop` is called exactly once, from `Producer.RemoveConsumer`, after
   the consumer has been removed from `p.consumers`.
6. `p.wg.Wait()` in `RemoveConsumer` only returns after `runLoop` has exited,
   ensuring `p.demuxer.Close()` is never called while `ReadPacket` is in flight.

---

## Known limitations

- **Stop does not flush consumers.** `Stop()` cancels contexts but does not
  call `RemoveConsumer` for active consumers. Muxers that require `WriteTrailer`
  for a valid output (e.g. MP4) will be closed without a trailer unless the
  caller removes all consumers first.

- **Packet drop is silent.** There is no drop counter or back-pressure signal.
  Applications that need lossless delivery should either size `pktChan`
  appropriately (currently a compile-time constant of 64) or implement their
  own flow control at the muxer layer.

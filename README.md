# net-protocol

A small **peer-to-peer style file server** in Go. Each node listens on TCP, keeps a **local store** under `{listenAddr}_network` (paths derived from keys; CAS-style hashing is optional), and exchanges files with connected peers.

## Flow in short

1. **Bootstrap** — On start, a node may `Dial` other nodes (see `Server_nodes`). Accepted and dialed connections are registered as **peers**.

2. **Control vs stream** — On the wire, the first byte of a unit is either:
   - `0x01` **IncomingMessage** — followed by a **gob**-encoded `Message` (metadata only).
   - `0x02` **IncomingStream** — followed by **raw bytes** (encrypted payload or file data; no gob in that segment).

3. **Store** (`Server.Store`)
   - Write the file locally, then **broadcast** a `Message` whose payload is `MessageStoreFile` (server ID, key, size).
   - Peers decode the gob, then **read exactly `Size` bytes** from the same TCP connection (encrypted stream) and write under the sender’s ID on disk.

4. **Get** (`Server.Get`)
   - If the object exists locally, read from disk.
   - Otherwise **broadcast** `MessageGetFile` (ID, key). A peer that has the file answers with `0x02`, then **little-endian int64 file size**, then plaintext file bytes; the requester decrypts/writes locally (same stream contract as store, but direction and format differ per handler).

5. **Main loop** — `Transport` pushes `RPC` values on a channel; `loop()` decodes the gob `Message` and dispatches to `handleMessage` (`MessageStoreFile` / `MessageGetFile`).

## Run

```bash
make run
```

This builds the binary (`make build` → `tmp/main`) and runs it. The sample `main` starts two nodes (`:3030` and `:7000`) and exercises store / delete / get.

## Stack

- **p2p** — TCP transport, framed reads via `DefaultDecoder`, peer `Send`.
- **Store** — Local filesystem layout + optional path transform.
- **crypto** — AES-GCM helpers used when storing/streaming encrypted blobs.

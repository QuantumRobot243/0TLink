# 0TLink

<p align="center">
  <img src="0TLink/assets/logo.jpg" alt="0TLink Logo" width="500">
</p>

## Overview

**0TLink** is more than just a simple proxy. It is a **multiplexed mTLS tunnel** designed for high-concurrency, identity-focused networking.  
The system is based on a single core principle:

> **Identity, not network location, is the main firewall.**

Every design choice in 0TLink supports this idea, from key generation to stream routing and failure recovery.

---

## 1. Zero-Trust Identity Lifecycle

The foundation of 0TLink’s security model is the **complete separation of identity from IP addresses**.

### Cryptographic Sovereignty
- Each Agent generates its own **RSA-2048 private key locally**.
- The private key stays on the machine.
- No keys are sent, stored, or derived remotely.
- This removes the risk of interception during provisioning.

### CSR Handshake
- The Agent submits a **Certificate Signing Request (CSR)** to the Control Plane.
- The CSR includes identity metadata (Common Name).
- The Control Plane checks a **bootstrap token** before signing.
- Only explicitly authorized nodes can join the mesh.

### Short-Lived Trust
- Certificates are signed with a **24-hour expiration**:
  - `NotAfter: now.Add(24 * time.Hour)`.
- Certificates are limited to:
  - `ExtKeyUsageClientAuth`.
- This creates a **time-limited trust window**.
- Even if a node is compromised, the impact is strictly limited.

---

## 2. Stream Multiplexing with Yamux

Traditional proxies create one TCP/TLS connection for each request. This adds handshake delay and connection overhead.

0TLink uses **Yamux (Yet Another Multiplexer)** to run hundreds of logical streams over a **single persistent mTLS connection**.

### Virtual Channels
- `SetupSession` upgrades a raw TLS 1.3 connection to a Yamux session.
- This removes repeated handshakes.
- It significantly reduces latency under load.

### Flow Control
- Configurable `YamuxMaxWindow` (default: 256 KB).
- This prevents high-bandwidth streams from starving others.
- It ensures fairness across multiplexed workloads.

### Persistence & Liveness
- `EnableKeepAlive` detects:
  - Silent firewall timeouts.
  - Network blackholes.
- It triggers automatic reconnection when failures occur.

---

## 3. Protocol Sniffing & Smart Routing

The Relay acts as an **intelligent traffic router** without performing full application-layer inspections.

### Non-Destructive Peeking
- `tunnel.Join` uses a `bufio.Reader`.
- It peeks at the first **512 bytes**.
- It identifies the protocol without consuming data.
- The destination application receives the untouched handshake.

### Matcher Pipeline
Traffic is classified using lightweight signature matching:

- **TLS**
  - Signature: `0x16 0x03`.
  - Routed as opaque, pre-encrypted traffic.
- **HTTP**
  - Matches standard verbs (`GET`, `POST`, etc.).
- **Databases**
  - Detects protocol-specific startup packets (e.g. PostgreSQL).

If no match is found, traffic falls back to **TCP/Opaque mode**.

### Dynamic Port Mapping
- A thread-safe `portMutex` assigns unique public ports.
- Ports are mapped to each unique **Common Name**.
- Example:
  - `:8081`, `:8082`, `:8083`.
- This lets multiple developers share a single Relay without conflicts.

---

## 4. Resiliency & Atomic Operations

0TLink is designed to be **crash-safe by default**, especially regarding identity and configuration state.

### Atomic Certificate Updates
- New credentials are first written to `.tmp` files.
- The final commit uses `os.Rename`.
- Guarantees:
  - No partial writes.
  - No corrupted certificates.
  - Safe recovery after crashes or power loss.

### Graceful Shutdown
- Relay listens for `SIGTERM` and `Interrupt` using `signal.NotifyContext`.
- On shutdown:
  - Listener ports close cleanly.
  - Connected Agents are notified.
  - No orphaned sockets or leaked state.

---

## Design Summary

| Aspect     | Guarantee                             |
|-----------|---------------------------------------|
| Identity  | Cryptographic, not network-based      |
| Transport | mTLS 1.3 everywhere                   |
| Concurrency | Hundreds of streams per tunnel       |
| Visibility | Protocol-aware, zero-knowledge       |
| Failure Mode | Crash-safe and self-healing        |

---

## Message
0TLink is an open-source project.feel free to open an issue or submit a pull request.


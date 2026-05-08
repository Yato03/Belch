# Belch

**A fast, terminal-native alternative to Burp Suite Community's Intruder.**

Burp Suite Community is an excellent free tool, but its Intruder module is deliberately throttled — concurrent threads are locked behind the Pro license, making brute-force and fuzzing attacks painfully slow for free users. This tool solves exactly that problem.

It replicates the Intruder workflow from the terminal: parse a raw `.req` file exported from Burp, mark injection points with `§`, pick an attack mode, point it at a wordlist — and run it with as many threads as your machine can handle. Results stream into a full-screen interactive TUI that mirrors the Intruder table (request number, status code, response length, response time, payload), with live filtering and export to CSV, JSON, or plain text.

## Requirements

- Go 1.22+

## Build

```bash
go build -o belch ./cmd/belch
```

On Windows:

```bash
go build -o belch.exe ./cmd/belch
```

## Usage

```
belch -req <file.req> -wordlist <list.txt> [options]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `-req` | (required) | Path to `.req` file |
| `-wordlist` | (required) | Wordlist file. Comma-separated for pitchfork: `users.txt,passwords.txt` |
| `-mode` | `sniper` | Attack mode: `sniper`, `battering-ram`, `pitchfork` |
| `-url` | (from Host header) | Base URL override, e.g. `https://target.example.com` |
| `-threads` | `1` | Number of concurrent requests |
| `-timeout` | `30s` | Per-request timeout |
| `-skip-verify` | `false` | Skip TLS certificate verification |

## Marking fuzz points

Export a raw request from Burp Suite and wrap injection points with `§`:

```
POST /login HTTP/2
Host: target.example.com
Content-Type: application/x-www-form-urlencoded

csrf=abc123&username=§admin§&password=§secret§
```

Save it as `request.req`.

## Examples

### Sniper — enumerate one field at a time

```bash
belch -req request.req -wordlist passwords.txt -mode sniper -url https://target.example.com
```

Output:
```
target: target.example.com
fuzz points: 2
mode: sniper

#      status   length   time         payload
────────────────────────────────────────────────────────────
1      200      1234     45ms         admin
2      302      0        12ms         root
3      200      1234     44ms         carlos
...
```

### Battering Ram — same payload in all fields

```bash
belch -req request.req -wordlist wordlist.txt -mode battering-ram
```

### Pitchfork — separate wordlist per field

```bash
belch -req request.req -wordlist users.txt,passwords.txt -mode pitchfork
```

Pairs `users.txt[i]` with `passwords.txt[i]` in lock-step. Stops when the shortest list is exhausted.

### Concurrent requests

```bash
belch -req request.req -wordlist wordlist.txt -threads 10
```

### Skip TLS verification (self-signed certs)

```bash
belch -req request.req -wordlist wordlist.txt -url https://target.local -skip-verify
```

## Attack mode comparison

| Mode | Points active | Results | Use case |
|---|---|---|---|
| Sniper | One at a time | N × M | Enumerate one field |
| Battering Ram | All simultaneously | M | Same value in multiple places |
| Pitchfork | All, separate lists | min(M₀…Mₙ) | Credential stuffing |

## Running tests

```bash
go test ./...
```

## Project structure

```
belch/
├── cmd/belch/main.go    # entry point
├── parser/              # parse .req files into Request structs
├── detector/            # detect and inject § fuzz points
├── wordlist/            # load payload lists from files
├── modes/               # sniper, battering-ram, pitchfork
├── executor/            # send HTTP requests, capture responses
└── ui/                  # Bubble Tea TUI (streaming results, filter, export)
```

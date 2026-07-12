# ssh-audits — Code Review

Reviewed: 2026-07-11. Build, `go vet`, and `gofmt` are all clean — the issues below are logic, design, and hygiene, not compile problems. Ordered by severity.

---

## 1. Bugs (fix these first)

### 1.1 The `-g` flag does nothing — it runs commands on EVERY host
`pkg/sshutil/sshutil.go:63-89` — `filterHosts()` only checks `*Limit` and `*LimitGroup`. `*Group` is never read there. So `Run()`'s `case *Group != ""` (line 107) routes to `ConnectToGroup()`, which falls through `filterHosts()` to `return inv.Hosts` — **all hosts in the inventory**, not the group you asked for.

For a tool that executes commands (often with `sudo`) on remote MySQL servers, running against every host when you asked for one group is the most dangerous bug in the repo.

**Fix:** add a `*Group` branch in `filterHosts()` that looks up `inv.Groups[*Group].Hosts`, and error out (don't fall through to all hosts) when a requested group/host doesn't exist.

### 1.2 Empty error check silently swallows SSH agent failure
`pkg/sshutil/sshutil.go:218-220`:
```go
auth, err := goph.UseAgent()
if err != nil {
}
```
The error is checked and then discarded. If the agent isn't running, `auth` is garbage and the user gets a confusing downstream connect error. Should be `return fmt.Errorf("error using SSH agent: %w", err)` like the other call sites.

### 1.3 The `-timeout` flag doesn't actually time anything out
`pkg/sshutil/sshutil.go:242-262` — `ExecuteCommands()` creates a context but calls blocking `client.Run(cmd)`. The `select { case <-ctx.Done(): ... default: ... }` only checks the deadline **between** commands; a single hung command (network stall, `sudo` password prompt) hangs forever. Same fake-timeout pattern in `pkg/sftp/sftp.go:57-64`, where the `select` runs exactly once before the upload starts, so it can never fire.

**Fix:** goph already has `client.RunContext(ctx, cmd)` — use it with a per-command context. Delete the `select`/`default` scaffolding.

### 1.4 Command failures don't fail the program (always exits 0)
`ExecuteCommands()` returns nothing; errors are printed and forgotten (`sshutil.go:255-258`). `ConnectToGroup()` also returns `nil` even if every host failed to connect. Scripting around this tool is impossible — you can't tell success from failure.

**Fix:** make `ExecuteCommands` return an error, aggregate per-host failures (e.g. `errors.Join`), and let `main()`'s `log.Fatal` do its job.

### 1.5 Empty/unmatched filters succeed silently
- `-l badhost` → `filterHosts` returns an empty map → loop over zero hosts → exit 0, no output. `sshutil.go:66-74`.
- Missing/unreadable commands file → `readCommands()` logs and returns `nil` (`sshutil.go:283-288`) → zero commands run → exit 0.

Both should be errors. Also: `readCommands()` ignores `scanner.Err()`, and runs blank lines as empty remote commands — skip blanks and `#` comments.

### 1.6 `ConnectToGroup` doesn't fall back to hostname; `ExecuteCommandOnHost` ignores `-p`
- `ConnectToHost()` falls back to `h.Name` when `ansible_host` is empty (`sshutil.go:191-194`); `ConnectToGroup()` (line 154) doesn't — empty `Addr` on hosts without `ansible_host`.
- `pkg/sftp/sftp.go:153` uses `goph.New(...)` — hardcoded port 22 (ignores `-p`), no hostname fallback, and a *different* host-key policy than every other path (no `AddNewHost` callback). One host, three behaviors depending on which flag you used.

### 1.7 `-h` collides with the conventional help flag
`sshutil.go:27` defines `-h` as the host flag, so `ssh-audits -h` errors with "flag needs an argument" instead of printing usage. Rename to `-host` (keep `-h` for help), or at minimum document it. Also the `-l` help text has an unclosed paren (line 33).

---

## 2. Security considerations

This is a personal ops tool, so judge these against your threat model — but worth knowing:

### 2.1 Trust-on-first-use with no prompt
Every connection path auto-adds unknown host keys via the `AddNewHost` callback — no fingerprint shown, no confirmation. A MITM on first contact is silently trusted. Consider printing the fingerprint and prompting (or a `-yes` flag for automation).

### 2.2 known_hosts check matches key material, not host+key
`sshutil.go:124-129` (and the sftp copy) do `strings.Contains(hostData, marshaledKey)` — if any *other* host in known_hosts has the same key, the new hostname is treated as already known and never gets added. `goph.AddKnownHost` already handles dedup correctly; the pre-check can just be dropped.

### 2.3 Unquoted remote paths / shell injection surface
`pkg/sftp/sftp.go:69,72,79,91` — `"chmod +x " + remoteFilePath` etc. The remote path comes from `filepath.Base()` of your local file: a filename with a space, `;`, or `$()` breaks or injects. Same in `pkg/sshkey/sshkey.go:77` (`echo '%s' >> authorized_keys` — a single quote in the key escapes it). Quote arguments (or use `shlex`-style quoting helper).

### 2.4 Fixed, predictable `/tmp` upload path
`sftp.go:122,165` — uploading to `/tmp/<name>` on a multi-user box is a classic symlink/squatting risk, and two concurrent runs clobber each other. Use remote `mktemp -d` and clean it up.

### 2.5 MD5 fingerprints in sshkey
`pkg/sshkey/sshkey.go:45-48` — MD5 over the key string is fine for local dedupe, but if you keep this package, match OpenSSH and use SHA256 so fingerprints are comparable with `ssh-keygen -lf`.

---

## 3. Structure & design

### 3.1 Flags as exported package globals
All flags live as exported globals in `pkg/sshutil` (`sshutil.go:25-38`), `main` reaches into them (`*sshutil.Host`), and the `-sftp` flag lives in `main`. This inverts the dependency (library owns CLI config), makes the packages untestable, and means importing `sshutil` registers flags as a side effect.

**Fix:** define flags in `cmd/ssh-audits/main.go`, collect them into a `Config`/`Options` struct, and pass it (or explicit params) into the packages. This one change unlocks testing.

### 3.2 `AddNewHost` is copy-pasted verbatim
`sshutil.go:115-139` and `sftp.go:21-45` are the same function (one has colors, one doesn't). Extract to one shared location — e.g. `pkg/hostkeys` or an `internal/` package both import.

### 3.3 The `"Vars: "` TrimPrefix hack, six times
`strings.TrimPrefix(h.Vars["ansible_user"], "Vars: ")` appears in 6 places. Current `aini` doesn't prefix values with `"Vars: "` — this looks like a workaround for a long-fixed bug. Verify with `-vars` output, then delete the trims; either way, centralize into one `hostVar(h *aini.Host, key string) string` helper that also handles the empty-value fallback (fixes 1.6 for free).

### 3.4 `pkg/sftp` re-implements inventory logic
`ExecuteCommandOnHost()` (`sftp.go:127-174`) parses the inventory and resolves user/addr itself, duplicating `sshutil`. Better shape: `sshutil` resolves inventory → hands a connected `*goph.Client` to the sftp package, which only does upload/exec. Also delete the commented-out flag-parsing scaffolding at `sftp.go:128-135`.

### 3.5 Hosts run sequentially
For an audit tool, running 50 hosts one at a time is the difference between 2 seconds and 2 minutes. `ConnectToGroup` is an ideal candidate for `golang.org/x/sync/errgroup` with a concurrency limit and per-host output prefixes (buffer each host's output so it doesn't interleave).

---

## 4. Remove

| What | Where | Why |
|---|---|---|
| `pkg/sshkey` (entire package) | `pkg/sshkey/` | Dead code — nothing imports it. Either wire it up as an `-copy-id` feature or delete it (git remembers). |
| Blank imports | `sshutil.go:20-21`, `sftp.go:12,16-17` | `_ "golang.org/x/crypto/ssh/agent"`, `_ ".../knownhosts"`, `_ "github.com/fatih/color"` do nothing. |
| `list.sh` | repo root | Personal Vagrant setup: hardcoded `/Users/klarsen/...` key paths, private IPs, your username. Doesn't belong in a public repo. |
| `my-thing.sh` commented block | `my-thing.sh:9-17` | Commented-out password-mangling and KILL-query SQL. If `my-thing.sh` stays as the `-sftp` demo script, trim it to the two live commands. |
| Site-specific `commands.txt` | repo root | It's the default `-f` target, so a fresh user runs *your* MySQL audit queries. Ship a generic `commands.txt.example` (like you did for `inventory/hosts`) and add `commands.txt` to `.gitignore`. Fold `commands-KEEP.txt` into it or drop it. |
| Stale comments | `sshutil.go:114` ("Export functions…"), `sshkey.go:90` ("...existing code...") | Copilot-session leftovers. |

---

## 5. Add

1. **Tests.** There are none. Start with the pure functions — table-driven tests for `filterHosts` (would have caught bug 1.1), `readCommands`, `GetCommands`, and the `hostVar` helper from 3.3. The 3.1 refactor makes this possible.
2. **CI.** A minimal GitHub Actions workflow: `go build ./... && go vet ./... && go test ./...` (+ `golangci-lint` if you like). You already take Dependabot PRs — CI is what makes merging them safe.
3. **`-version` flag** and a tagged release / `goreleaser` config, so remote boxes can report what they're running.
4. **README updates.** Document `-l`, `-lg`, `-p`, `-timeout`, `-f`, `-vars`; current examples only show `-i/-h/-g/-c/-sftp`. Note the `-h`/help collision until renamed.
5. **A `dry-run` flag** that prints host list + commands without connecting. Cheap to build, huge safety win given 1.1.

---

## 6. Housekeeping

- `github.com/fatih/color v1.13.0` is ~4 years old (current is v1.18+); `goph v1.4.0` and `aini v1.6.0` are current-ish. `go get -u ./... && go mod tidy` after the fixes.
- `go.sum` lists `pkg/sftp` as indirect even though the sftp feature depends on it via goph — fine, just don't be surprised.
- Consider `internal/` instead of `pkg/` for packages you don't intend others to import.

---

## Suggested order of attack

1. Bug 1.1 (`-g` filter) + error on unmatched host/group — the safety fix.
2. Bug 1.2 (empty `if err`) — one line.
3. 3.1 flag refactor (flags → main, config struct) — unlocks everything else.
4. 3.3 `hostVar` helper (kills the `TrimPrefix` hack, fixes 1.6).
5. 1.3/1.4 (`RunContext` + error propagation).
6. Deduplicate `AddNewHost` (3.2), remove dead code & personal files (§4).
7. Tests + CI (§5).

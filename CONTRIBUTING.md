# Contributing to gone

Thanks for considering a contribution. `gone` is a small, focused tool — we want to keep it that way. Before writing code, please **open an issue** so we can align on scope.

## Ground rules

- **macOS only.** `gone` targets macOS exclusively. Cross-platform PRs will be closed.
- **No databases of known apps.** `gone` scans the actual filesystem. We don't want a curated list of tools.
- **Trash, never `rm`.** File removal must go through macOS Trash via `osascript`. Put Back must always work.
- **Small surface area.** Prefer fixing bugs and polishing existing features over adding new ones.

## Development setup

```bash
git clone https://github.com/yasku/gone.git
cd gone
go build ./cmd/gone/
go test ./...
```

Requires Go 1.26+ and macOS. No other dependencies beyond what's in `go.mod`.

## Running locally

```bash
go run ./cmd/gone
```

Use `Tab` to switch between Uninstall and Monitor. Press `?` for help.

## Pull request flow

1. **Open an issue first.** Describe the bug or feature. Wait for a thumbs-up before writing code.
2. **Fork and branch.** Name your branch something descriptive: `fix/rc-scanner-zshenv`, `feat/preview-images`.
3. **Keep commits atomic.** One logical change per commit. Follow the repo's commit prefix style:
   - `feat:` — new functionality
   - `fix:` — bug fix
   - `docs:` — README, CHANGELOG, comments
   - `refactor:` — code restructure without behavior change
   - `test:` — test-only changes
   - `chore:` — tooling, deps, housekeeping
4. **Run the full verification before pushing.**

   ```bash
   go fmt ./...
   go vet ./...
   go build ./...
   go test ./...
   ```

   All must pass. PRs with failing builds or tests will not be reviewed.
5. **Update `CHANGELOG.md`.** Add an entry under the current date describing your change.
6. **Open the PR.** Link the issue. Describe what changed and why. Include screenshots or GIFs for TUI changes.

## Code style

- `go fmt` is the only formatter. Don't argue with it.
- Exported identifiers get doc comments. Internal ones don't need them unless the "why" is non-obvious.
- Prefer small files and small functions. The `internal/` tree is organized by domain (`scanner`, `remover`, `sysinfo`, `tui`).
- TUI styles live in `internal/tui/styles.go`. Don't scatter lipgloss calls across models.

## Testing

- Unit tests live next to the code (`foo.go` → `foo_test.go`).
- Test names follow `TestFooDoesX` — readable, not cryptic.
- Avoid fragile filesystem tests. Use `t.TempDir()` and clean up after yourself.
- TUI rendering is not unit-tested — verify manually and include a GIF in the PR.

## Reporting bugs

Include:

- macOS version (`sw_vers -productVersion`)
- Go version (`go version`)
- Full command that triggered the issue
- Expected vs actual behavior
- Relevant output from `~/Library/Application Support/gone/operations.jsonl` if applicable

## Reporting security issues

Do **not** open a public issue. Email the maintainer directly via the address on [agustiny-dev.ar](http://agustiny-dev.ar).

## Code of conduct

Be direct. Be respectful. Don't waste people's time. We ship code, not drama.

---

*WITNESS ME.*

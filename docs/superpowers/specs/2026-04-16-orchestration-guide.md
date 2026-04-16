# Orchestrating Claude Code Workers — A Practical Guide

**Date:** 2026-04-16  
**Context:** Built while implementing `gone`, a Go TUI with 9 sequential build steps.

---

## 1. The Problem

Claude Code is powerful in a single session but breaks down on large projects in three specific ways.

**Context limits.** Every session has a finite context window. A project with 9 incremental build steps, a Go codebase, a research doc, and a plan file will overflow the context before step 6. The agent starts hallucinating, forgetting what it already built, or re-implementing things that exist.

**Agents that forget.** Even within the window, a long-running session accumulates cruft: failed attempts, backtracked decisions, retried commands. The agent's working memory is polluted. Each new task starts with the full residue of every previous task.

**Plans that don't get executed.** You write a careful step-by-step plan. The agent reads it, nods, and then does something different — or implements all 9 steps at once instead of one at a time, skipping verify-and-commit checkpoints. The plan was aspirational. Nobody enforced it.

The fix: **one fresh worker per task, coordinated by a supervisor that can't forget.**

---

## 2. The Solution: Plan Files as Bridges + Supervisor as Dispatcher

The architecture has three components:

```
┌─────────────────────────────────────┐
│          Supervisor (TypeScript)    │
│  Reads task list, dispatches        │
│  workers one at a time, logs        │
│  session IDs, stops on failure      │
└──────────────┬──────────────────────┘
               │ spawns
               ▼
      ┌─────────────────┐
      │  Worker (claude) │  ← fresh session every task
      │  Reads plan file │
      │  Reads KB file   │
      │  Does ONE task   │
      │  Commits + stops │
      └─────────────────┘
               │ writes to
               ▼
      ┌─────────────────┐
      │  Shared codebase │
      │  + plan file     │
      │  + knowledge base│
      └─────────────────┘
```

**Plan files are the bridge.** The plan lives on disk. Each worker reads it fresh. The supervisor does not need to carry the context of what was built — the plan and the codebase carry it. Workers communicate through artifacts (files, commits), not through shared memory.

**Key invariants:**
- One worker per task. No worker sees the previous worker's session.
- Workers are stateless. All state is in the filesystem.
- The supervisor is dumb. It reads a task list, spawns workers, reads exit codes, logs session IDs.
- Workers self-terminate. The system prompt says "STOP after Task N."

---

## 3. How to Spawn a Nested Claude Session

When Claude Code runs, it sets two environment variables that prevent nested invocations:

```
CLAUDECODE=1
CLAUDE_CODE_ENTRYPOINT=cli
```

If you try to spawn `claude` from inside a Claude Code session without clearing these, it refuses or behaves unpredictably. The fix is to strip them before spawning:

```bash
env -u CLAUDECODE -u CLAUDE_CODE_ENTRYPOINT claude [flags] "prompt"
```

In TypeScript/Node.js (as used in `supervisor.ts`):

```typescript
const env = { ...process.env };
delete env.CLAUDECODE;
delete env.CLAUDE_CODE_ENTRYPOINT;

const proc = spawn("claude", args, {
  cwd: workingDir,
  env,
  stdio: ["ignore", "pipe", "pipe"],
});
```

> **Critical:** Pass `stdio: ["ignore", ...]` to avoid a stdin warning. When `claude` is invoked non-interactively with `-p`, it expects no stdin. If you pass `stdin: "inherit"` from a terminal session, you'll get a warning. Passing `"ignore"` suppresses it cleanly.

---

## 4. Key CLI Flags

These are the flags you actually need. All others are optional or situational.

### Required for non-interactive operation

| Flag | Purpose |
|------|---------|
| `-p` / `--print` | Non-interactive mode. Takes a prompt, runs, exits. No REPL. |
| `--output-format json` | Returns a JSON object with `result`, `session_id`, `cost_usd`, `duration_ms`. Without this you get raw text and no session ID. |

### Identity and control

| Flag | Purpose |
|------|---------|
| `--system-prompt "..."` | Replaces the default system prompt entirely. Use this to give the worker its role, its constraints, its stop condition. |
| `--name gone-task-3` | Labels the session. Shows up in logs. Makes debugging easier. |
| `--model sonnet` | Picks the model. `sonnet` (claude-sonnet-4-5 or latest) for implementation tasks. `haiku` for cheap triage. `opus` when you need maximum reasoning. |
| `--effort high` | Sets thinking depth. `high` = extended thinking enabled. `low` = faster and cheaper. |

### Permissions

| Flag | Purpose |
|------|---------|
| `--allowedTools Edit,Write,Bash,Read,Glob,Grep` | Whitelist exactly the tools the worker needs. Reject everything else. Comma-separated. |
| `--dangerously-skip-permissions` | Skip the interactive permission prompts for bash commands. Required for automated workers that run `go build`, `git commit`, etc. Use only in controlled environments. |

### Budget

| Flag | Purpose |
|------|---------|
| `--max-budget-usd 2` | Hard cap per worker invocation. The session aborts if it would exceed this. Protects against runaway tasks. |

### Resumption

| Flag | Purpose |
|------|---------|
| `--resume <session_id>` | Resume a previous session by ID. Use this when a worker fails and you want to debug interactively. |

### Input/output formats

| Flag | Purpose |
|------|---------|
| `--input-format stream-json` | Accept a stream of JSON messages on stdin instead of a plain prompt. Required for bidirectional pipe architectures (see §7). |
| `--bare` | Strip all wrapping — no system prompt injection, no default tools, no auth context. **Kills authentication.** Do not use unless you know exactly what you're doing (see §8). |

### Full example command (what `supervisor.ts` builds):

```bash
env -u CLAUDECODE -u CLAUDE_CODE_ENTRYPOINT \
  claude \
  --system-prompt "You are a Go developer implementing gone. Do Task 3 only. Stop when done." \
  --name "gone-task-3" \
  --allowedTools "Edit,Write,Bash,Read,Glob,Grep" \
  --dangerously-skip-permissions \
  --model sonnet \
  --effort high \
  --max-budget-usd 2 \
  -p \
  --output-format json \
  "Implement Task 3 from the plan. Read docs/superpowers/plans/2026-04-15-gone.md first."
```

---

## 5. The supervisor.ts Pattern

`gone/orchestrator/supervisor.ts` is a Bun script (~325 lines) that implements the full sequential dispatch pattern. Here's how it works.

### How it works

**1. Parse args.** Accepts `--from N`, `--to N`, `--only 0,1,2`, `--dry-run`, `--interactive N`. This gives you surgical control over which tasks run.

**2. Build the command for each task.** The `buildCommand(taskNum, interactive)` function assembles the full `claude` CLI invocation. For non-interactive mode it adds `-p`, `--output-format json`, and `--max-budget-usd`. For interactive mode it omits those (giving you a live terminal session).

**3. Per-task system prompt.** Each worker gets a strict system prompt via `--system-prompt`. The prompt tells the worker:
   - What role it is ("you are a Go developer")
   - What file to read first (the plan)
   - What task number to implement
   - The fallback procedure if imports fail
   - Explicit stop condition: "STOP after Task N is complete"

**4. Spawn the worker.** Uses Node.js `child_process.spawn` with the cleaned env (CLAUDECODE stripped). Non-interactive workers pipe stdout/stderr; interactive workers use `stdio: "inherit"`.

**5. Parse JSON output.** When the worker exits, the supervisor tries to `JSON.parse` the output and extracts `session_id` and `result`.

**6. Log to JSONL.** Every task run appends a line to `orchestrator/logs/sessions.jsonl`:
   ```json
   {"task": 3, "name": "Uninstall TUI", "success": true, "sessionId": "abc123", "duration": 187, "ts": "2026-04-15T..."}
   ```

**7. Stop on failure.** If any task exits with non-zero, the supervisor prints the failure, shows the session ID for resumption, and halts. Does not continue to the next task.

**8. Print summary.** After all tasks complete, prints a table of ✓/✗ per task with duration and session IDs.

### How to run it

```bash
# Run all tasks 0-8
bun run orchestrator/supervisor.ts

# Start from task 3 (tasks 0-2 already done)
bun run orchestrator/supervisor.ts --from 3

# Run only tasks 0, 1, 2
bun run orchestrator/supervisor.ts --only 0,1,2

# Preview commands without executing
bun run orchestrator/supervisor.ts --dry-run

# Debug task 5 interactively (you see the live session)
bun run orchestrator/supervisor.ts --interactive 5
```

### How to resume a failed session

When a task fails, the supervisor prints:

```
*** Task 3 FAILED. Stopping. ***
Resume with: bun run orchestrator/supervisor.ts --from 3
Debug session: env -u CLAUDECODE -u CLAUDE_CODE_ENTRYPOINT claude --resume "abc123xyz"
```

Use `--resume` with the session ID to re-enter the exact conversation state and diagnose the failure interactively.

### How to customize it for your project

1. **Change `PLAN_FILE` and `KNOWLEDGE_BASE`** to point to your plan and research docs.
2. **Update `TASK_NAMES`** with your task names.
3. **Update `TOTAL_TASKS`**.
4. **Rewrite `systemPrompt(taskNum)`** to describe your project domain (Go, Python, Rust — whatever).
5. **Adjust `MODEL`, `EFFORT`, `MAX_BUDGET_PER_TASK`** based on task complexity.
6. **Adjust `ALLOWED_TOOLS`** to the exact tool set your workers need.
7. **Change `cwd`** in the `spawn` call to point at your project root.

The system prompt is the most important thing to get right. It must tell the worker:
- What project it's working on
- Where to find the plan (full relative path from cwd)
- What task number to implement
- The explicit stop condition

---

## 6. Mini-Agent as an Alternative Orchestrator

[Mini-Agent](https://github.com/MiniMax-AI/Mini-Agent) is a Python ReAct-loop agent framework originally built by MiniMax. It runs one agent per session but its `BashTool` with background process support makes it usable as a supervisor that dispatches `claude` CLI workers.

### When to use it instead of supervisor.ts

- You want a **reasoning orchestrator** — one that decides which task to run next, not just iterates a list
- You need to **adapt the plan dynamically** based on worker output
- You want an **interactive REPL** where you can steer the orchestration live
- You're already running Python and don't want a Bun dependency

### Architecture with Mini-Agent

The cleanest setup uses a custom `ClaudeWorkerTool` that the orchestrator LLM can call:

```python
class ClaudeWorkerTool(Tool):
    name = "claude_worker"
    description = "Dispatch a task to a claude CLI worker and return results"
    parameters = {
        "type": "object",
        "properties": {
            "task": {"type": "string"},
            "working_dir": {"type": "string"},
        },
        "required": ["task"]
    }

    async def execute(self, task: str, working_dir: str = ".") -> ToolResult:
        env = {k: v for k, v in os.environ.items()
               if k not in ("CLAUDECODE", "CLAUDE_CODE_ENTRYPOINT")}
        proc = await asyncio.create_subprocess_exec(
            "claude", "--print", "--output-format", "json",
            "--allowedTools", "Edit,Write,Bash,Read,Glob,Grep",
            "--dangerously-skip-permissions",
            "-p", task,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
            cwd=working_dir,
            env=env,
        )
        stdout, stderr = await proc.communicate()
        return ToolResult(
            success=proc.returncode == 0,
            content=stdout.decode(),
            error=stderr.decode() if proc.returncode != 0 else None,
        )
```

Wire it into a Mini-Agent `Agent` alongside `BashTool` and `ReadTool`. The orchestrator LLM calls `claude_worker` for each implementation task, reads results, and decides whether to continue or retry.

### Approach B: Background workers via BashTool

For parallel dispatch, use `BashTool` with `run_in_background=True`:

```
bash(command="env -u CLAUDECODE -u CLAUDE_CODE_ENTRYPOINT claude --print -p 'Implement X' > /tmp/task1.json", run_in_background=True)
bash(command="env -u CLAUDECODE -u CLAUDE_CODE_ENTRYPOINT claude --print -p 'Implement Y' > /tmp/task2.json", run_in_background=True)
bash_output(bash_id="abc12345")  # poll until done
```

### Limitations of Mini-Agent for orchestration

| Issue | Impact |
|-------|--------|
| Tools run sequentially (no native parallel dispatch) | Must use background bash as the workaround for parallelism |
| No built-in multi-agent coordination primitives | You build task routing yourself |
| Orchestrator context grows with each worker result | The summarization pass helps but adds ~2-4s latency per summary |
| Python runtime required | Additional dependency vs. pure TypeScript supervisor |
| Default config points at MiniMax endpoints | Must override `api_base: https://api.anthropic.com` and set `provider: anthropic` to use real Claude |

### When supervisor.ts is simpler

If your task list is fixed and sequential (like the `gone` build steps), `supervisor.ts` is ~100 lines less code and requires no reasoning from the orchestrator. Use Mini-Agent when the orchestration logic itself needs to be smart.

---

## 7. Future Improvements

### 7.1 Stream-JSON bidirectional pipes

The `--input-format stream-json` flag enables a different communication model: instead of passing a static prompt string, you pipe a stream of JSON messages to the worker's stdin. This unlocks mid-session injection — the orchestrator can push new instructions after seeing partial output.

Architecture:
```
orchestrator stdout → worker stdin  (stream-json messages)
worker stdout       → orchestrator stdin (stream-json events)
```

This is more complex to implement (requires managing two async streams) but allows the orchestrator to react to worker output in real time rather than waiting for the session to exit.

### 7.2 Persistent sessions (resume instead of restart)

Currently each worker is a fresh session. A smarter approach: capture the `session_id` from each worker's JSON output and resume it if the task partially succeeds. This avoids re-reading large context files on retry.

```bash
# First attempt
claude --output-format json -p "Implement Task 3..." → {"session_id": "abc123", ...}

# Retry (picks up where it left off)
claude --resume abc123 --output-format json -p "The tests failed. Fix the scanner_test.go"
```

The `sessions.jsonl` log already captures session IDs. Wire in a retry loop that resumes before falling back to a fresh session.

### 7.3 Multi-model: cheap orchestrator + expensive workers

Current setup: every worker runs on `sonnet` at `effort: high`. Better: use `haiku` or `sonnet` at `effort: low` for triage and planning tasks, `opus` or `sonnet` at `effort: high` only for implementation.

Example tiered strategy:
```
Task 0 (scaffold): haiku, effort=low    → very cheap, trivial task
Task 1 (scanner):  sonnet, effort=low   → moderate complexity
Task 3 (TUI):      sonnet, effort=high  → complex, needs extended thinking
Task 5 (trash):    sonnet, effort=low   → straightforward
```

Implement as a `TASK_CONFIG` map in `supervisor.ts`:

```typescript
const TASK_CONFIG: Record<number, {model: string; effort: string; budget: string}> = {
  0: { model: "haiku", effort: "low", budget: "0.25" },
  3: { model: "sonnet", effort: "high", budget: "5" },
  // ...
};
```

### 7.4 Parallel workers for independent tasks

Some tasks have no dependencies on each other. Steps 1 (file scanner) and 2 (RC scanner) both write to `internal/scanner/` but touch different files. They could run in parallel with worktree isolation:

```bash
git worktree add /tmp/gone-task-1 main
git worktree add /tmp/gone-task-2 main
# spawn two workers concurrently, each in its own worktree
# merge results when both complete
```

`supervisor.ts` would need a `Promise.all` path with worktree setup/teardown.

---

## 8. Gotchas and Lessons Learned

### --bare kills authentication

Do not pass `--bare` to worker sessions unless you have explicitly set up API key authentication in the environment. `--bare` strips all wrapping, including the auth context that Claude Code normally injects. Workers launched with `--bare` will fail with auth errors.

**When you'd use it:** Building a truly minimal subprocess that speaks raw JSON and manages its own tools. For typical worker dispatch, leave `--bare` out entirely.

### The stdin warning

If you spawn `claude -p "..."` with `stdio: "inherit"` (the default when spawning from a terminal), you'll get:

```
Warning: stdin is a TTY but --print was specified. Ignoring stdin.
```

This is harmless but noisy. Fix it by passing `"ignore"` for stdin:

```typescript
stdio: ["ignore", "pipe", "pipe"]  // stdin ignored, stdout/stderr piped
```

### Capture session IDs, always

Every `--output-format json` response includes a `session_id`. Log it immediately. If a worker fails silently (exit 0 but bad output), you can re-enter the session with `--resume` and see exactly what happened. Without the session ID you have to start over.

```typescript
const json = JSON.parse(output);
const sessionId = json.session_id;  // capture this
logSession({ task: taskNum, sessionId, ... });
```

### The CLAUDECODE env var guard

Claude Code sets `CLAUDECODE=1` and `CLAUDE_CODE_ENTRYPOINT=cli` to detect nesting. Forgetting to strip these before spawning a child `claude` process is the most common failure mode. The child either refuses to run or behaves strangely.

Always strip both:
```typescript
const env = { ...process.env };
delete env.CLAUDECODE;
delete env.CLAUDE_CODE_ENTRYPOINT;
```

### System prompts must include an explicit stop condition

If you don't tell the worker when to stop, it won't. A worker given access to a 9-task plan will try to implement all 9 tasks. The system prompt must say, explicitly:

```
10. STOP after Task N is complete. Do NOT start the next task.
```

Workers will honor this if it's unambiguous. Without it, they drift.

### Workers need working directory set correctly

Relative paths in prompts (like `docs/superpowers/plans/2026-04-15-gone.md`) only resolve correctly if `cwd` in the spawn call matches the root those paths are relative to. In `supervisor.ts`, the cwd is set to `scripts/` (the project root), not `gone/orchestrator/`. Double-check this if workers can't find their plan files.

```typescript
cwd: join(PROJECT_ROOT, ".."),  // scripts/ root — plan paths are relative here
```

### $2/task budget is about right for implementation tasks

For a task like "implement a Bubble Tea TUI with multi-select" on `sonnet` at `effort: high`, budget ~$1.50-$2.00 per task. Simpler tasks (scaffold, log module) run well under $0.50. Set per-task budgets if you want fine control. A flat $2 cap works as a reasonable safety net.

### Non-zero exit ≠ bad code, zero exit ≠ good code

`claude` exits with 0 even if the implementation is wrong. Exit code only tells you whether the process itself crashed. Always inspect the `result` field in the JSON output. The supervisor logs `output.substring(0, 500)` for quick triage — look there first when a task "succeeds" but the code is broken.

---

## Quick Reference

```bash
# Spawn a single worker manually
env -u CLAUDECODE -u CLAUDE_CODE_ENTRYPOINT \
  claude \
  --system-prompt "You are a worker. Do Task 3 only. Stop after." \
  --name "task-3" \
  --allowedTools "Edit,Write,Bash,Read,Glob,Grep" \
  --dangerously-skip-permissions \
  --model sonnet \
  --effort high \
  --max-budget-usd 2 \
  -p \
  --output-format json \
  "Implement Task 3 from docs/superpowers/plans/your-plan.md"

# Run supervisor
bun run orchestrator/supervisor.ts --from 0

# Resume a failed session
env -u CLAUDECODE -u CLAUDE_CODE_ENTRYPOINT claude --resume <session_id>

# Dry-run to see commands
bun run orchestrator/supervisor.ts --dry-run

# Interactive debug for one task
bun run orchestrator/supervisor.ts --interactive 3
```

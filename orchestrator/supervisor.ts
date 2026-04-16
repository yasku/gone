#!/usr/bin/env bun
/**
 * gone — Task Supervisor
 *
 * Dispatches Claude Code workers to implement plan tasks sequentially.
 * Each worker gets a fresh session with a strict system prompt.
 *
 * Usage:
 *   bun run orchestrator/supervisor.ts                  # run all tasks (0-8)
 *   bun run orchestrator/supervisor.ts --from 3         # start from task 3
 *   bun run orchestrator/supervisor.ts --only 0         # run only task 0
 *   bun run orchestrator/supervisor.ts --only 0,1,2,3   # run tasks 0-3
 *   bun run orchestrator/supervisor.ts --dry-run        # show commands without executing
 *   bun run orchestrator/supervisor.ts --interactive 3   # run task 3 interactively (no -p)
 */

import { spawn, execSync } from "child_process";
import { readFileSync, appendFileSync, mkdirSync, existsSync } from "fs";
import { join } from "path";

// --- Config ---
const PROJECT_ROOT = join(import.meta.dir, "..");
const DOCS_ROOT = join(PROJECT_ROOT, "..", "docs", "superpowers");
const PLAN_FILE = "docs/superpowers/plans/2026-04-15-gone.md";
const KNOWLEDGE_BASE = "docs/superpowers/specs/2026-04-15-gone-research.md";
const LOG_DIR = join(PROJECT_ROOT, "orchestrator", "logs");
const SESSION_LOG = join(LOG_DIR, "sessions.jsonl");
const MODEL = "sonnet";
const EFFORT = "high";
const MAX_BUDGET_PER_TASK = "2"; // USD
const ALLOWED_TOOLS = "Edit,Write,Bash,Read,Glob,Grep";
const TOTAL_TASKS = 9; // Tasks 0-8

// --- Task metadata ---
const TASK_NAMES: Record<number, string> = {
  0: "Scaffold + Hello World",
  1: "File Scanner",
  2: "Shell RC Scanner",
  3: "Uninstall TUI — Search, Scan, List, Multi-Select",
  4: "Preview Pane",
  5: "Trash + Operation Log",
  6: "System Monitor",
  7: "Root Model + Tab Switching",
  8: "Polish",
};

function systemPrompt(taskNum: number): string {
  return `You are a Go developer implementing a TUI application called "gone".
You are a WORKER agent controlled by a supervisor. Do exactly what is asked.

RULES:
1. Read the plan file FIRST: ${PLAN_FILE}
2. Read the knowledge base: ${KNOWLEDGE_BASE} — use the patterns exactly as written, do not improvise.
3. Implement ONLY Task ${taskNum}: "${TASK_NAMES[taskNum]}".
4. Follow every step in the plan IN ORDER. Do not skip steps.
5. If "go mod tidy" fails with charm.land imports, apply the v1 fallback described in the plan header:
   - charm.land/bubbletea/v2 -> github.com/charmbracelet/bubbletea
   - charm.land/bubbles/v2/* -> github.com/charmbracelet/bubbles/*
   - charm.land/lipgloss/v2 -> github.com/charmbracelet/lipgloss
   - tea.KeyPressMsg -> tea.KeyMsg
   - tea.View / tea.NewView(s) -> string
6. Run the verify command specified in the plan after implementation.
7. If tests fail, fix them before committing.
8. Commit when done with the message from the plan.
9. Update CHANGELOG.md at the project root (gone/CHANGELOG.md) with: date, task number, what was done.
10. STOP after Task ${taskNum} is complete. Do NOT start the next task.
11. When you finish, output a summary: what files were created/modified, what worked, any issues.`;
}

function userPrompt(taskNum: number): string {
  return `Implement Task ${taskNum} ("${TASK_NAMES[taskNum]}") from the plan.
Start by reading the plan file: ${PLAN_FILE}
Then read the knowledge base: ${KNOWLEDGE_BASE}
Work in the gone/ directory. Create it if it doesn't exist (Task 0 only).`;
}

// --- Helpers ---

function ensureLogDir() {
  if (!existsSync(LOG_DIR)) {
    mkdirSync(LOG_DIR, { recursive: true });
  }
}

function logSession(entry: Record<string, unknown>) {
  ensureLogDir();
  appendFileSync(SESSION_LOG, JSON.stringify(entry) + "\n");
}

function timestamp(): string {
  return new Date().toISOString();
}

function parseArgs() {
  const args = process.argv.slice(2);
  const config = {
    from: 0,
    to: TOTAL_TASKS - 1,
    only: null as number[] | null,
    dryRun: false,
    interactive: null as number | null,
  };

  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case "--from":
        config.from = parseInt(args[++i]);
        break;
      case "--to":
        config.to = parseInt(args[++i]);
        break;
      case "--only":
        config.only = args[++i].split(",").map(Number);
        break;
      case "--dry-run":
        config.dryRun = true;
        break;
      case "--interactive":
        config.interactive = parseInt(args[++i]);
        break;
    }
  }
  return config;
}

function buildCommand(taskNum: number, interactive: boolean): string[] {
  const args = [
    "claude",
    "--system-prompt",
    systemPrompt(taskNum),
    "--name",
    `gone-task-${taskNum}`,
    "--allowedTools",
    ALLOWED_TOOLS,
    "--dangerously-skip-permissions",
    "--model",
    MODEL,
    "--effort",
    EFFORT,
  ];

  if (!interactive) {
    args.push("-p");
    args.push("--output-format", "json");
    args.push("--max-budget-usd", MAX_BUDGET_PER_TASK);
  }

  // User prompt goes last (as the prompt argument)
  args.push(userPrompt(taskNum));

  return args;
}

function getLastCommit(): string {
  try {
    return execSync("git log --oneline -1", {
      cwd: join(PROJECT_ROOT),
      encoding: "utf-8",
    }).trim();
  } catch {
    return "(no commits yet)";
  }
}

async function runTask(taskNum: number, interactive: boolean): Promise<{
  success: boolean;
  sessionId?: string;
  output: string;
  duration: number;
}> {
  const args = buildCommand(taskNum, interactive);
  const startTime = Date.now();

  console.log(`\n${"=".repeat(60)}`);
  console.log(`  TASK ${taskNum}: ${TASK_NAMES[taskNum]}`);
  console.log(`  Started: ${timestamp()}`);
  console.log(`  Model: ${MODEL} | Effort: ${EFFORT} | Budget: $${MAX_BUDGET_PER_TASK}`);
  console.log(`${"=".repeat(60)}\n`);

  // Remove Claude nesting guards
  const env = { ...process.env };
  delete env.CLAUDECODE;
  delete env.CLAUDE_CODE_ENTRYPOINT;

  return new Promise((resolve) => {
    let output = "";
    const proc = spawn(args[0], args.slice(1), {
      cwd: join(PROJECT_ROOT, ".."), // scripts/ root so relative paths to docs/ work
      env,
      stdio: interactive ? "inherit" : ["ignore", "pipe", "pipe"], // ignore stdin to avoid warning
    });

    if (!interactive) {
      proc.stdout?.on("data", (data: Buffer) => {
        output += data.toString();
        // Print progress dots
        process.stdout.write(".");
      });
      proc.stderr?.on("data", (data: Buffer) => {
        // Stderr often has progress info
        const line = data.toString().trim();
        if (line) {
          console.log(`  [worker] ${line}`);
        }
      });
    }

    proc.on("close", (code) => {
      const duration = Math.round((Date.now() - startTime) / 1000);
      console.log(`\n  Completed in ${duration}s (exit code: ${code})`);

      let sessionId: string | undefined;
      if (!interactive && output) {
        try {
          const json = JSON.parse(output);
          sessionId = json.session_id;
          // Print the actual text result
          if (json.result) {
            console.log(`\n  Result:\n${json.result.substring(0, 500)}...`);
          }
        } catch {
          // Output wasn't JSON, print raw
          console.log(`\n  Output:\n${output.substring(0, 500)}...`);
        }
      }

      const lastCommit = getLastCommit();
      console.log(`  Last commit: ${lastCommit}`);

      resolve({
        success: code === 0,
        sessionId,
        output,
        duration,
      });
    });
  });
}

// --- Main ---
async function main() {
  const config = parseArgs();
  const tasks = config.only ?? Array.from(
    { length: config.to - config.from + 1 },
    (_, i) => config.from + i
  );

  console.log(`
╔══════════════════════════════════════════════════════╗
║           gone — Task Supervisor                     ║
║                                                      ║
║  Tasks: ${tasks.join(", ").padEnd(44)}║
║  Model: ${MODEL.padEnd(44)}║
║  Mode:  ${config.dryRun ? "DRY RUN" : config.interactive !== null ? "INTERACTIVE" : "AUTOMATED".padEnd(44)}║
╚══════════════════════════════════════════════════════╝
`);

  if (config.dryRun) {
    for (const taskNum of tasks) {
      const args = buildCommand(taskNum, false);
      console.log(`\nTask ${taskNum}: ${TASK_NAMES[taskNum]}`);
      console.log(`  Command: ${args.join(" ").substring(0, 200)}...`);
    }
    return;
  }

  // Interactive mode for a single task
  if (config.interactive !== null) {
    const result = await runTask(config.interactive, true);
    logSession({
      task: config.interactive,
      name: TASK_NAMES[config.interactive],
      ...result,
      ts: timestamp(),
    });
    return;
  }

  // Sequential automated execution
  const results: Record<number, { success: boolean; sessionId?: string; duration: number }> = {};

  for (const taskNum of tasks) {
    const result = await runTask(taskNum, false);

    logSession({
      task: taskNum,
      name: TASK_NAMES[taskNum],
      success: result.success,
      sessionId: result.sessionId,
      duration: result.duration,
      ts: timestamp(),
    });

    results[taskNum] = {
      success: result.success,
      sessionId: result.sessionId,
      duration: result.duration,
    };

    if (!result.success) {
      console.log(`\n  *** Task ${taskNum} FAILED. Stopping. ***`);
      console.log(`  Resume with: bun run orchestrator/supervisor.ts --from ${taskNum}`);
      if (result.sessionId) {
        console.log(`  Debug session: env -u CLAUDECODE -u CLAUDE_CODE_ENTRYPOINT claude --resume "${result.sessionId}"`);
      }
      break;
    }

    console.log(`  ✓ Task ${taskNum} complete\n`);
  }

  // Summary
  console.log(`\n${"=".repeat(60)}`);
  console.log("  SUMMARY");
  console.log(`${"=".repeat(60)}`);
  for (const [task, r] of Object.entries(results)) {
    const status = r.success ? "✓" : "✗";
    const session = r.sessionId ? ` (session: ${r.sessionId})` : "";
    console.log(`  ${status} Task ${task}: ${r.duration}s${session}`);
  }
  console.log(`\nSession log: ${SESSION_LOG}`);
}

main().catch(console.error);

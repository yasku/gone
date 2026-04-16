# Mini-Agent Research Report

**Repo:** https://github.com/MiniMax-AI/Mini-Agent  
**Date:** 2026-04-16  
**Version:** 0.1.0 (no tags; latest commit: `d76a4f6`)

---

## 1. What Is It?

Mini-Agent is a **minimal, production-quality single-agent framework** built by MiniMax to showcase best practices for their MiniMax M2.5 model. It is explicitly a demo/reference implementation, not a full orchestration platform.

Key capabilities:
- Multi-turn interactive CLI (`mini-agent` command) with ESC-to-cancel
- Non-interactive `--task` flag for scripted execution
- Full tool-calling agent loop (ReAct-style)
- Context management: auto-summarizes history when tokens exceed a configurable limit
- MCP (Model Context Protocol) tool integration
- ACP (Agent Client Protocol) server mode for IDE integration (e.g., Zed)
- 15 bundled "Claude Skills" (document generation, canvas design, webapp testing, etc.)

It is **not** a multi-agent orchestrator out of the box — it runs one `Agent` instance per session. However the architecture is clean enough to extend.

---

## 2. Architecture & Agent Loop

### Package Layout

```
mini_agent/
├── agent.py          # Core Agent class — the ReAct loop
├── cli.py            # CLI entry point + interactive REPL
├── config.py         # Pydantic config (YAML-backed)
├── schema/           # Pydantic data models (Message, ToolCall, LLMResponse)
├── llm/
│   ├── llm_wrapper.py    # LLMClient: selects Anthropic or OpenAI backend
│   ├── anthropic_client.py
│   ├── openai_client.py
│   └── base.py
├── tools/
│   ├── base.py           # Tool + ToolResult base classes
│   ├── bash_tool.py      # BashTool, BashOutputTool, BashKillTool
│   ├── file_tools.py     # ReadTool, WriteTool, EditTool
│   ├── mcp_loader.py     # MCP client integration
│   ├── note_tool.py      # SessionNoteTool (persistent memory JSON)
│   └── skill_tool.py     # Skill loader / get_skill tool
├── acp/
│   └── __init__.py       # ACP bridge (Agent Client Protocol / stdio server)
├── skills/           # 15 bundled Claude Code skills (git submodule)
└── config/
    ├── config-example.yaml
    ├── mcp-example.json
    └── system_prompt.md
```

### The Agent Loop (`agent.py`)

```
while step < max_steps:
    1. Check cancellation (asyncio.Event)
    2. Summarize history if tokens > limit  (calls LLM for summary)
    3. Call LLM with full message history + tool schemas → LLMResponse
    4. Append assistant message (with optional thinking block)
    5. If no tool_calls → DONE, return response.content
    6. For each tool_call:
       a. Look up tool by name
       b. await tool.execute(**arguments) → ToolResult
       c. Append tool result message (role="tool")
       d. Check cancellation
    7. step++
```

This is a standard **ReAct** (reason + act) loop. No parallel tool execution — tools run sequentially. The loop is `async` throughout.

### Context Management

- Token counting via `tiktoken` (`cl100k_base` encoder)
- Triggered when either local estimate **or** API-reported `total_tokens` exceeds `token_limit` (default 80k)
- Summarization strategy: keep all user messages, summarize agent+tool execution between each user turn, call LLM once per turn to generate the summary

---

## 3. Language & Dependencies

**Language:** Python ≥ 3.10 (uses `X | Y` union syntax)  
**Package manager:** uv (recommended), pip works too

**Runtime dependencies** (from `pyproject.toml`):

| Package | Purpose |
|---------|---------|
| `anthropic >= 0.39.0` | Anthropic API SDK |
| `openai >= 1.57.4` | OpenAI-compatible API SDK |
| `pydantic >= 2.0.0` | Data models |
| `pyyaml >= 6.0.0` | Config loading |
| `httpx >= 0.27.0` | HTTP client |
| `mcp >= 1.0.0` | Model Context Protocol client |
| `agent-client-protocol >= 0.6.0` | ACP server |
| `tiktoken >= 0.5.0` | Token counting |
| `prompt-toolkit >= 3.0.0` | Interactive CLI |
| `requests >= 2.31.0` | HTTP |

**Dev deps:** pytest, pytest-asyncio, pytest-cov, pytest-xdist

---

## 4. Tool Calling

### Tool Definition

Every tool extends `Tool` (abstract base):

```python
class Tool:
    @property
    def name(self) -> str: ...
    @property
    def description(self) -> str: ...
    @property
    def parameters(self) -> dict[str, Any]: ...  # JSON Schema
    async def execute(self, **kwargs) -> ToolResult: ...
    def to_schema(self) -> dict:          # Anthropic format
    def to_openai_schema(self) -> dict:   # OpenAI format
```

### Built-in Tools

| Tool | Class | Description |
|------|-------|-------------|
| `bash` | `BashTool` | Foreground/background shell execution |
| `bash_output` | `BashOutputTool` | Poll output from background bash process |
| `bash_kill` | `BashKillTool` | Terminate background bash process |
| `read_file` | `ReadTool` | Read file contents |
| `write_file` | `WriteTool` | Write/create files |
| `edit_file` | `EditTool` | Edit existing files |
| `session_note` | `SessionNoteTool` | Persistent JSON memory store |
| `get_skill` | skill tool | Load one of 15 bundled Claude Skills |
| MCP tools | `MCPTool` | Any tool exposed by an MCP server |

### Tool Call Flow

1. LLM returns `ToolCall` objects (parsed from either Anthropic `tool_use` blocks or OpenAI `tool_calls` array)
2. Agent looks up tool by `function.name`, calls `await tool.execute(**arguments)`
3. Result (`ToolResult.content` or `ToolResult.error`) is appended as a `role="tool"` message
4. Loop continues until LLM produces a response with no tool calls

---

## 5. API Compatibility

The framework is **dual-protocol** — it speaks both Anthropic and OpenAI APIs, with automatic routing based on a `provider` config key.

### Anthropic Protocol (`provider: "anthropic"`)

- Uses the official `anthropic.AsyncAnthropic` SDK
- Default endpoint: `https://api.minimax.io/anthropic` (MiniMax's Anthropic-compatible endpoint)
- Supports **extended thinking** (Anthropic-style `thinking` content blocks)
- Tool call format: `tool_use` / `tool_result` content blocks
- Can point at **real Anthropic API** (`https://api.anthropic.com`) by setting `api_base`

### OpenAI Protocol (`provider: "openai"`)

- Uses the official `openai.AsyncOpenAI` SDK
- Default endpoint: `https://api.minimax.io/v1`
- Supports reasoning via `reasoning_split: True` extra body parameter and `reasoning_details` in message history
- Can point at **any OpenAI-compatible API** (e.g., `https://api.siliconflow.cn/v1`)

### ACP Protocol

- Exposes the agent as an ACP-compatible stdio server (`mini-agent-acp`)
- Implements: `initialize`, `newSession`, `prompt`, `cancel`
- Used for Zed editor integration and other ACP-aware clients

### Third-Party API Support

The `LLMClient` constructor auto-detects whether `api_base` is a MiniMax domain:
- **MiniMax domains** (`api.minimax.io`, `api.minimaxi.com`): auto-appends `/anthropic` or `/v1`
- **Everything else**: uses `api_base` as-is (enabling any OpenAI-compatible endpoint)

---

## 6. Key Files

| File | Purpose |
|------|---------|
| `mini_agent/agent.py` | Core `Agent` class, ReAct loop, token management, cancellation |
| `mini_agent/llm/llm_wrapper.py` | `LLMClient` — provider dispatch |
| `mini_agent/llm/anthropic_client.py` | Anthropic SDK integration, message format conversion, thinking |
| `mini_agent/llm/openai_client.py` | OpenAI SDK integration, reasoning_details |
| `mini_agent/tools/base.py` | `Tool` + `ToolResult` abstractions |
| `mini_agent/tools/bash_tool.py` | Shell execution (foreground + background with monitoring) |
| `mini_agent/tools/file_tools.py` | File read/write/edit |
| `mini_agent/tools/mcp_loader.py` | MCP client, stdio/SSE/HTTP transports |
| `mini_agent/acp/__init__.py` | ACP bridge (`MiniMaxACPAgent`) |
| `mini_agent/cli.py` | CLI entrypoint, tool wiring, interactive REPL |
| `mini_agent/config.py` | Pydantic config models, YAML loading, config path resolution |
| `mini_agent/schema/schema.py` | `Message`, `ToolCall`, `LLMResponse`, `LLMProvider` |
| `mini_agent/config/config-example.yaml` | Reference config with all options documented |
| `mini_agent/config/system_prompt.md` | Default system prompt |

---

## 7. Using Mini-Agent as an Orchestrator for a Go TUI Project

### The Core Idea

Mini-Agent's `Agent` class + `BashTool` already knows how to:
- Call an LLM to plan and reason
- Execute shell commands (foreground and background)
- Monitor long-running processes via `bash_output`
- Cancel execution cleanly

This makes it straightforward to use Mini-Agent as a **Python orchestrator** that dispatches `claude` CLI workers (subprocesses) for Go TUI development tasks.

### Approach A: Python Orchestrator Wrapping `claude` CLI

The cleanest approach: write a custom tool that spawns `claude --print --no-color -p "<task>"` as a subprocess and returns the output.

```python
# Custom tool: ClaudeWorkerTool
class ClaudeWorkerTool(Tool):
    name = "claude_worker"
    description = "Dispatch a task to a claude CLI worker and return results"
    parameters = {
        "type": "object",
        "properties": {
            "task": {"type": "string", "description": "Task for the Claude worker"},
            "working_dir": {"type": "string", "description": "Working directory for the worker"},
        },
        "required": ["task"]
    }

    async def execute(self, task: str, working_dir: str = ".") -> ToolResult:
        proc = await asyncio.create_subprocess_shell(
            f'claude --print --no-color -p "{task}"',
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
            cwd=working_dir,
        )
        stdout, stderr = await proc.communicate()
        return ToolResult(
            success=proc.returncode == 0,
            content=stdout.decode(),
            error=stderr.decode() if proc.returncode != 0 else None
        )
```

Then create an `Agent` with `[ClaudeWorkerTool(), BashTool(), ReadTool(), WriteTool()]` and give it a system prompt that describes the Go TUI project structure. The orchestrator LLM (MiniMax M2.5 or even real Claude via `api_base: https://api.anthropic.com`) plans the work and delegates implementation tasks to `claude` CLI workers.

### Approach B: Background Workers with `BashTool`

Use the existing `BashTool` with `run_in_background=True` to dispatch multiple concurrent `claude` CLI invocations:

```
bash(command="claude --print -p 'Implement X' > /tmp/task1.txt", run_in_background=True)
bash(command="claude --print -p 'Implement Y' > /tmp/task2.txt", run_in_background=True)
bash_output(bash_id="abc12345")   # poll for completion
```

This gives the orchestrator visibility into multiple parallel workers.

### Approach C: ACP Server Mode

Run `mini-agent-acp` as an stdio ACP server. An ACP-compatible client (or a Go TUI that speaks ACP over stdio) can issue `PromptRequest` messages and receive streaming `session_notification` updates with tool call visibility. This is the cleanest architecture for a TUI that wants to display live agent progress.

### Practical Setup for Go TUI Orchestration

1. **Config**: Point `api_base` at Anthropic (`https://api.anthropic.com`) with `provider: anthropic`, use a real Claude model as the orchestrator brain
2. **System prompt**: Describe the Go TUI project — module layout, bubbletea/lipgloss conventions, target features
3. **Tools**: Add `ClaudeWorkerTool` (or just use `BashTool` to run `claude --print -p "..."`)
4. **Task delegation pattern**: Orchestrator breaks work into chunks (design component X, implement Y, write tests for Z), dispatches each to a `claude` worker in its own directory context
5. **Result integration**: Workers write files directly; orchestrator reads results and synthesizes

### Limitations / Caveats

| Issue | Detail |
|-------|--------|
| No native parallel tool dispatch | Tools execute sequentially in the loop; background bash is the workaround |
| Single-model, single-session | No built-in multi-agent coordination primitives |
| Python runtime required | Orchestrator is Python; workers are `claude` CLI (different runtime) |
| No shared state between workers | Each `claude` invocation is stateless; orchestrator must manage state |
| Token context of orchestrator | Context grows with each worker result injected; summarization helps but adds latency |
| MiniMax-centric defaults | Default model and endpoints are MiniMax; must override `api_base` + `model` to use Anthropic/Claude directly |

### Recommended Architecture for Go TUI Project

```
┌─────────────────────────────────────┐
│  Mini-Agent Orchestrator (Python)   │
│  - LLM: Claude 3.7 Sonnet           │
│  - api_base: api.anthropic.com      │
│  - Tools: BashTool + ClaudeWorker   │
│  - System prompt: Go TUI project    │
└─────────────┬───────────────────────┘
              │ spawns subprocesses
    ┌─────────┼──────────┐
    ▼         ▼          ▼
claude CLI  claude CLI  claude CLI
worker 1    worker 2    worker 3
(component) (tests)     (docs)
    │         │          │
    └────files written to shared workspace────┘
              │
    ┌─────────▼──────────┐
    │   Go TUI Project   │
    └────────────────────┘
```

The orchestrator runs `mini-agent --task "Build Go TUI project: ..."` non-interactively, or interactively via the REPL for iterative development.

---

## Summary

Mini-Agent is a clean, well-structured Python agent framework with dual Anthropic/OpenAI API support, MCP integration, and an ACP server mode. Its `BashTool` with background process support makes it trivially usable as a supervisor that dispatches `claude` CLI workers. The main gap is native parallel agent coordination — this must be handled manually via background bash processes or custom tools. The codebase is small enough (~500 LOC in core) to fork and extend without difficulty.

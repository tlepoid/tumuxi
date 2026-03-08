<p align="center">
  <img width="339" height="105" alt="Screenshot 2026-01-20 at 1 00 23 AM" src="https://github.com/user-attachments/assets/fdbefab9-9f7c-4e08-a423-a436dda3c496" />
</p>

<p align="center">TUI for easily running parallel coding agents</p>

<p align="center">
  <a href="https://github.com/tlepoid/tumuxi/releases">
    <img src="https://img.shields.io/github/v/release/tlepoid/tumuxi?style=flat-square" alt="Latest release" />
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/tlepoid/tumuxi?style=flat-square" alt="License" />
  </a>
  <img src="https://img.shields.io/badge/Go-1.24.12-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go version" />
</p>

<p align="center">
  <a href="#quick-start">Quick start</a> ·
  <a href="#features">Features</a> ·
  <a href="#configuration">Configuration</a>
</p>

<img width="3840" height="2160" alt="image" src="https://github.com/user-attachments/assets/63aa6c74-0a71-4475-a493-404be6408f5b" />

## What is tumuxi?

tumuxi is a TUI for running multiple coding agents in parallel. Each agent works in isolation on its own git worktree branch, so you can merge changes back when done.

## Prerequisites

tumuxi requires [tmux](https://github.com/tmux/tmux) (minimum 3.2). Each agent runs in its own tmux session for terminal isolation and persistence.

## Quick start

Via the install script:

```bash
curl -fsSL https://raw.githubusercontent.com/tlepoid/tumuxi/main/install.sh | sh
```

Or with Go:

```bash
go install github.com/tlepoid/tumuxi/cmd/tumuxi@latest
```

## Features

- **Parallel agents**: Launch multiple agents within main repo and within workspaces
- **No wrappers**: Works with Claude Code, Codex, Gemini, Amp, OpenCode, and Droid
- **Keyboard + mouse**: Can be operated with just the keyboard or with a mouse
- **All-in-one tool**: Run agents, view diffs via lazygit, and access terminal

## Architecture quick tour

Start with `internal/app/ARCHITECTURE.md` for lifecycle, PTY flow, tmux tagging, and persistence invariants. Message boundaries and command discipline are documented in `internal/app/MESSAGE_FLOW.md`.

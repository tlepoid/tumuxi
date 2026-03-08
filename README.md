<p align="center">
  <img width="339" height="105" alt="Screenshot 2026-01-20 at 1 00 23 AM" src="https://github.com/user-attachments/assets/fdbefab9-9f7c-4e08-a423-a436dda3c496" />  
</p>

<p align="center">TUI for easily running parallel coding agents</p>

<p align="center">
  <a href="https://github.com/andyrewlee/amux/releases">
    <img src="https://img.shields.io/github/v/release/andyrewlee/amux?style=flat-square" alt="Latest release" />
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/andyrewlee/amux?style=flat-square" alt="License" />
  </a>
  <img src="https://img.shields.io/badge/Go-1.24.12-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go version" />
  <a href="https://discord.gg/Dswc7KFPxs">
    <img src="https://img.shields.io/badge/Discord-5865F2?style=flat-square&logo=discord&logoColor=white" alt="Discord" />
  </a>
</p>

<p align="center">
  <a href="#quick-start">Quick start</a> ·
  <a href="#how-it-works">How it works</a> ·
  <a href="#features">Features</a> ·
  <a href="#configuration">Configuration</a>
</p>

<img width="3840" height="2160" alt="image" src="https://github.com/user-attachments/assets/63aa6c74-0a71-4475-a493-404be6408f5b" />

## What is amux?

amux is a terminal UI for running multiple coding agents in parallel with a workspace-first model that can import git worktrees.

> This repository is a personal fork of [andyrewlee/amux](https://github.com/andyrewlee/amux).

## Prerequisites

amux requires [tmux](https://github.com/tmux/tmux) (minimum 3.2). Each agent runs in its own tmux session for terminal isolation and persistence.

## Quick start

Via the install script:

```bash
curl -fsSL https://raw.githubusercontent.com/tlepoid/agent-mux/main/install.sh | sh
```

Or with Go:

```bash
go install github.com/tlepoid/agent-mux/cmd/amux@latest
```


## How it works

Each workspace tracks a repo checkout and its metadata. For local workflows, workspaces are typically backed by git worktrees on their own branches so agents work in isolation and you can merge changes back when done.

## Architecture quick tour

Start with `internal/app/ARCHITECTURE.md` for lifecycle, PTY flow, tmux tagging, and persistence invariants. Message boundaries and command discipline are documented in `internal/app/MESSAGE_FLOW.md`.

## Features

- **Parallel agents**: Launch multiple agents within main repo and within workspaces
- **No wrappers**: Works with Claude Code, Codex, Gemini, Amp, OpenCode, and Droid
- **Keyboard + mouse**: Can be operated with just the keyboard or with a mouse
- **All-in-one tool**: Run agents, view diffs, and access terminal

<p align="center">
  <img width="400" height="135" alt="image" src="https://github.com/user-attachments/assets/0650f319-7b4c-4448-af4c-f929f8b55de0" />
</p>

<p align="center">TUI for running parallel coding agents</p>

<img width="3831" height="2155" alt="image" src="https://github.com/user-attachments/assets/cd1e75a7-6568-4863-bb47-7b0230c1eb6f" />

---


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
- **Integrate with GitHub**: Syncs with github to autopopulate agents with context from issues

## Architecture

```mermaid
graph TB
    subgraph TUI["Bubble Tea TUI"]
        direction TB
        Input["Input Handler<br/>(keyboard + mouse)"]
        App["App State Manager"]
        Compositor["Compositor & Chrome Cache"]

        subgraph Layout["Three-Pane Layout"]
            Dashboard["Dashboard<br/>(projects & workspaces)"]
            Center["Center Pane<br/>(agent tabs)"]
            Sidebar["Sidebar<br/>(git status & terminal)"]
        end
    end

    subgraph Services["Background Services"]
        Supervisor["Supervisor<br/>(worker pool)"]
        FileWatcher["File Watcher<br/>(fsnotify)"]
        StateWatcher["State Watcher"]
        ActivityTracker["Activity Tracker"]
        UpdateChecker["Update Checker"]
    end

    subgraph Core["Core Systems"]
        Git["Git / Worktree Manager"]
        Tmux["tmux Session Manager"]
        PTY["PTY / Agent Manager"]
        VTerm["Virtual Terminal<br/>(VT100 emulator)"]
        Config["Config & Registry"]
        Data["Data Layer<br/>(Project, Workspace, Tab)"]
    end

    subgraph External["External"]
        GitRepo["Git Repository"]
        TmuxServer["tmux Server"]
        Agents["AI Agents<br/>(Claude, Codex, Gemini, ...)"]
    end

    Input --> App
    App --> Layout
    Layout --> Compositor
    Compositor --> VTerm

    App <--> Data
    App <--> Config
    App --> Supervisor

    Supervisor --> FileWatcher
    Supervisor --> StateWatcher
    Supervisor --> ActivityTracker
    Supervisor --> UpdateChecker

    FileWatcher --> Git
    PTY --> VTerm
    Center --> PTY

    Git --> GitRepo
    Tmux --> TmuxServer
    PTY --> Tmux
    Tmux --> Agents
```

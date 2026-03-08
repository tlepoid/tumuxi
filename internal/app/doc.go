// Package app is the central orchestration layer of tumuxi. It owns the
// top-level Bubbletea model (App), coordinates workspace and project
// lifecycle, routes messages between UI panels (dashboard, center agent pane,
// sidebar, terminal), and enforces the single-writer invariant for workspace
// state mutations described in ARCHITECTURE.md.
package app

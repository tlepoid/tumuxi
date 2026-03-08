// Package cli defines the tumuxi command-line interface using Cobra. It exposes
// subcommands for managing workspaces, projects, agents, sessions, and
// terminals, as well as the root TUI entry point. Each subcommand wires
// together service dependencies and delegates to the appropriate business
// logic in internal/app.
package cli

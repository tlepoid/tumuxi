package process

import (
	"fmt"
	"os"
	"strconv"

	"github.com/tlepoid/tumux/internal/data"
)

// EnvBuilder builds environment variables for script execution
type EnvBuilder struct {
	portAllocator *PortAllocator
}

// NewEnvBuilder creates a new environment builder
func NewEnvBuilder(ports *PortAllocator) *EnvBuilder {
	return &EnvBuilder{
		portAllocator: ports,
	}
}

// BuildEnv creates environment variables for a workspace
func (b *EnvBuilder) BuildEnv(ws *data.Workspace) []string {
	env := os.Environ()

	// Add workspace-specific variables
	env = append(env,
		"TUMUX_WORKSPACE_NAME="+ws.Name,
		"TUMUX_WORKSPACE_ROOT="+ws.Root,
		"TUMUX_WORKSPACE_BRANCH="+ws.Branch,
		"ROOT_WORKSPACE_PATH="+ws.Repo,
	)

	// Add port allocation
	if b.portAllocator != nil {
		port, rangeEnd := b.portAllocator.PortRange(ws.Root)
		env = append(env,
			fmt.Sprintf("TUMUX_PORT=%d", port),
			fmt.Sprintf("TUMUX_PORT_RANGE=%d-%d", port, rangeEnd),
		)
	}

	// Add custom environment from workspace
	for k, v := range ws.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}

// BuildEnvMap creates a map of environment variables
func (b *EnvBuilder) BuildEnvMap(ws *data.Workspace) map[string]string {
	envMap := make(map[string]string)

	envMap["TUMUX_WORKSPACE_NAME"] = ws.Name
	envMap["TUMUX_WORKSPACE_ROOT"] = ws.Root
	envMap["TUMUX_WORKSPACE_BRANCH"] = ws.Branch
	envMap["ROOT_WORKSPACE_PATH"] = ws.Repo

	if b.portAllocator != nil {
		port, rangeEnd := b.portAllocator.PortRange(ws.Root)
		envMap["TUMUX_PORT"] = strconv.Itoa(port)
		envMap["TUMUX_PORT_RANGE"] = fmt.Sprintf("%d-%d", port, rangeEnd)
	}

	for k, v := range ws.Env {
		envMap[k] = v
	}

	return envMap
}

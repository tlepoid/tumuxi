package common

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
)

// omarchyThemeDir returns the path to the omarchy current theme directory.
func omarchyThemeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "omarchy", "current", "theme")
}

// omarchyThemeName reads the current omarchy theme name.
func omarchyThemeName() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "omarchy", "current", "theme.name"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// parseColorsToml parses a simple key = "value" TOML file into a map.
func parseColorsToml(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	colors := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"")
		colors[key] = val
	}
	return colors, scanner.Err()
}

// blendColor blends two hex colors by the given ratio (0.0 = a, 1.0 = b).
func blendColor(a, b string, ratio float64) string {
	ca, err1 := colorful.Hex(a)
	cb, err2 := colorful.Hex(b)
	if err1 != nil || err2 != nil {
		return a
	}
	return ca.BlendLab(cb, ratio).Hex()
}

// SystemTheme builds a theme from the current omarchy system colors.
// Falls back to Gruvbox if omarchy config is not available.
func SystemTheme() Theme {
	dir := omarchyThemeDir()
	if dir == "" {
		return fallbackSystemTheme()
	}

	colors, err := parseColorsToml(filepath.Join(dir, "colors.toml"))
	if err != nil {
		return fallbackSystemTheme()
	}

	get := func(key, fallback string) string {
		if v, ok := colors[key]; ok && v != "" {
			return v
		}
		return fallback
	}

	bg := get("background", "#282828")
	fg := get("foreground", "#ebdbb2")
	accent := get("accent", "#d79921")
	color0 := get("color0", bg)
	color1 := get("color1", "#cc241d")
	color2 := get("color2", "#98971a")
	color3 := get("color3", "#d79921")
	color4 := get("color4", "#458588")
	color6 := get("color6", "#689d6a")
	color8 := get("color8", "#928374")
	selBg := get("selection_background", blendColor(bg, fg, 0.2))
	activTab := get("active_tab_background", accent)

	name := omarchyThemeName()
	if name == "" {
		name = "System"
	} else {
		name = "System (" + name + ")"
	}

	return Theme{
		ID:   ThemeSystem,
		Name: name,
		Colors: ThemeColors{
			Background:    lipgloss.Color(bg),
			Foreground:    lipgloss.Color(fg),
			Muted:         lipgloss.Color(color8),
			Border:        lipgloss.Color(blendColor(bg, fg, 0.15)),
			BorderFocused: lipgloss.Color(accent),

			Primary:   lipgloss.Color(accent),
			Secondary: lipgloss.Color(color6),
			Success:   lipgloss.Color(color2),
			Warning:   lipgloss.Color(color3),
			Error:     lipgloss.Color(color1),
			Info:      lipgloss.Color(color4),

			Surface0: lipgloss.Color(color0),
			Surface1: lipgloss.Color(blendColor(bg, fg, 0.05)),
			Surface2: lipgloss.Color(blendColor(bg, fg, 0.10)),
			Surface3: lipgloss.Color(blendColor(bg, fg, 0.15)),

			Selection: lipgloss.Color(selBg),
			Highlight: lipgloss.Color(activTab),
		},
	}
}

// fallbackSystemTheme returns a Gruvbox theme labeled as System when omarchy is unavailable.
func fallbackSystemTheme() Theme {
	t := GruvboxTheme()
	t.ID = ThemeSystem
	t.Name = "System"
	return t
}

package common

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// renderIssuePickerLines renders the combined name-input + filtered issue list.
func (d *Dialog) renderIssuePickerLines() []string {
	var lines []string
	contentWidth := d.dialogContentWidth()

	// Input field.
	lines = append(lines, d.input.View())
	lines = append(lines, "")

	// Issue list section.
	if len(d.options) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorMuted()).Render("No open issues found"))
		return lines
	}

	headerStyle := lipgloss.NewStyle().Foreground(ColorMuted())
	if len(d.filteredIndices) == 0 {
		lines = append(lines, headerStyle.Render("No matching issues"))
		return lines
	}

	lines = append(lines, headerStyle.Render("Issues:"))

	selectedStyle := lipgloss.NewStyle().Foreground(ColorForeground()).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(ColorMuted())
	cursorOn := Icons.Cursor + " "
	cursorOff := Icons.CursorEmpty + " "

	baseLine := d.renderedLineCount(lines)

	for cursorIdx, originalIdx := range d.filteredIndices {
		label := d.options[originalIdx]
		// Truncate long labels to fit the dialog.
		maxLabelWidth := contentWidth - len(cursorOn) - 1
		if lipgloss.Width(label) > maxLabelWidth && maxLabelWidth > 3 {
			label = label[:maxLabelWidth-1] + "…"
		}

		var line string
		if cursorIdx == d.cursor {
			line = cursorOn + selectedStyle.Render(label)
		} else {
			line = cursorOff + normalStyle.Render(label)
		}

		d.addOptionHit(cursorIdx, originalIdx, baseLine+cursorIdx, 0, contentWidth)
		lines = append(lines, line)
	}

	return lines
}

// issuePickerCursorOffset returns the number of rendered lines before the input
// field inside a DialogIssuePicker, for correct terminal cursor placement.
func (d *Dialog) issuePickerCursorOffset() int {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary()).MarginBottom(1)
	title := titleStyle.Render(d.title)
	// title renders with MarginBottom(1) which adds one blank line.
	return strings.Count(title, "\n") + 2 // title lines + the blank line after
}

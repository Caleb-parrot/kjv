package tui

import "strings"

func wrapWords(s string, width int) []string {
	if width < 8 {
		width = 8
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	var cur strings.Builder
	col := 0
	for _, w := range words {
		need := len(w)
		if col > 0 {
			need++
		}
		if col > 0 && col+need > width {
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteString(w)
			col = len(w)
			continue
		}
		if col > 0 {
			cur.WriteByte(' ')
			col++
		}
		cur.WriteString(w)
		col += len(w)
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/caleb-parrot/kjv/internal/bible"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type pane int

const (
	paneIndex pane = iota
	paneText
)

type itemKind int

const (
	kindSection itemKind = iota
	kindBook
	kindChapter
)

type indexItem struct {
	kind    itemKind
	book    int
	chapter int
	label   string
}

type model struct {
	b      *bible.Bible
	width  int
	height int

	items   []indexItem
	cursor  int
	offset  int
	open    map[int]bool
	query   string
	typing  bool
	focus   pane
	book    int
	chapter int

	vp     viewport.Model
	status string
}

func Run(b *bible.Bible) error {
	// Last-horizon remaps the 16-color yellow/blue slots; keep the
	// cream-and-blue reader look with 24-bit color.
	lipgloss.SetColorProfile(termenv.TrueColor)
	m := newModel(b)
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func newModel(b *bible.Bible) model {
	m := model{
		b:       b,
		open:    map[int]bool{0: true},
		book:    0,
		chapter: 1,
		vp:      viewport.New(40, 20),
	}
	if r, err := b.ParseRef(loadLast()); err == nil {
		m.book = r.BookIdx
		m.chapter = r.Chapter
		m.open = map[int]bool{r.BookIdx: true}
	}
	m.rebuild()
	m.cursor = m.findItem(m.book, m.chapter)
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) rebuild() {
	m.items = buildIndex(m.b, m.open, m.query)
	if len(m.items) == 0 {
		m.cursor = 0
		return
	}
	m.cursor = clamp(m.cursor, 0, len(m.items)-1)
	m.skipNonSelectable(0)
}

func buildIndex(b *bible.Bible, open map[int]bool, query string) []indexItem {
	q := strings.ToLower(strings.TrimSpace(query))
	var qBook string
	qChap := 0
	if q != "" {
		qBook, qChap = splitQuery(q)
	}

	var items []indexItem
	var last bible.Section = -1
	for i, bk := range b.Books {
		sec := bible.SectionOf(bk.Number)
		if sec != last {
			items = append(items, indexItem{
				kind:  kindSection,
				label: sec.Label(),
			})
			last = sec
		}

		bookHit := q == "" || bookMatches(bk, qBook)
		if !bookHit {
			continue
		}

		items = append(items, indexItem{
			kind:  kindBook,
			book:  i,
			label: bk.Name,
		})

		showCh := open[i] || q != ""
		if !showCh {
			continue
		}
		for _, ch := range bk.Chapters {
			if qChap > 0 && ch.Num != qChap {
				continue
			}
			items = append(items, indexItem{
				kind:    kindChapter,
				book:    i,
				chapter: ch.Num,
				label:   strconv.Itoa(ch.Num),
			})
		}
	}
	return items
}

func splitQuery(q string) (book string, chapter int) {
	q = strings.ReplaceAll(q, ":", " ")
	parts := strings.Fields(q)
	if len(parts) == 0 {
		return "", 0
	}
	last := parts[len(parts)-1]
	if n, err := strconv.Atoi(last); err == nil && n > 0 && len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], " "), n
	}
	return q, 0
}

func bookMatches(bk bible.Book, q string) bool {
	if q == "" {
		return true
	}
	name := strings.ToLower(bk.Name)
	abbr := strings.ToLower(bk.Abbr)
	q = strings.ToLower(q)
	if strings.Contains(name, q) || strings.Contains(abbr, q) {
		return true
	}
	// numbered books: "1 jo" → "1 john"
	compact := func(s string) string {
		return strings.ReplaceAll(s, " ", "")
	}
	return strings.Contains(compact(name), compact(q)) || strings.HasPrefix(abbr, q)
}

func (m model) findItem(book, chapter int) int {
	bestBook := -1
	for i, it := range m.items {
		if it.kind == kindChapter && it.book == book && it.chapter == chapter {
			return i
		}
		if it.kind == kindBook && it.book == book {
			bestBook = i
		}
	}
	if bestBook >= 0 {
		return bestBook
	}
	for i, it := range m.items {
		if it.kind != kindSection {
			return i
		}
	}
	return 0
}

func (m *model) skipNonSelectable(dir int) {
	if len(m.items) == 0 {
		return
	}
	if dir == 0 {
		dir = 1
	}
	start := m.cursor
	for m.items[m.cursor].kind == kindSection {
		m.cursor += dir
		if m.cursor < 0 || m.cursor >= len(m.items) {
			m.cursor = start
			// try the other way
			m.cursor = clamp(m.cursor-dir, 0, len(m.items)-1)
			if m.items[m.cursor].kind == kindSection {
				for i, it := range m.items {
					if it.kind != kindSection {
						m.cursor = i
						return
					}
				}
			}
			return
		}
		if m.cursor == start {
			return
		}
	}
}

func (m *model) applyCursor() {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return
	}
	it := m.items[m.cursor]
	switch it.kind {
	case kindBook:
		m.book = it.book
		m.chapter = 1
	case kindChapter:
		m.book = it.book
		m.chapter = it.chapter
	}
}

func (m *model) refreshText() {
	bk, ch, _, ok := m.b.Get(m.book, m.chapter, 1)
	if !ok {
		m.vp.SetContent("")
		return
	}
	inner := max(20, m.vp.Width)
	m.vp.SetContent(renderChapter(ch, inner))
	_ = bk
}

func renderChapter(ch bible.Chapter, width int) string {
	numW := 1
	if len(ch.Verses) >= 10 {
		numW = 2
	}
	if len(ch.Verses) >= 100 {
		numW = 3
	}
	textW := max(12, width-numW-2)
	var b strings.Builder
	for i, v := range ch.Verses {
		if i > 0 {
			b.WriteByte('\n')
		}
		num := verseNumStyle.Render(fmt.Sprintf("%*d", numW, v.Num))
		lines := wrapWords(v.Text, textW)
		pad := strings.Repeat(" ", numW)
		for li, line := range lines {
			if li == 0 {
				b.WriteString(num)
				b.WriteString(" ")
				b.WriteString(verseTextStyle.Render(line))
			} else {
				b.WriteByte('\n')
				b.WriteString(verseNumStyle.Render(pad))
				b.WriteString(" ")
				b.WriteString(verseTextStyle.Render(line))
			}
		}
	}
	return b.String()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout()
		m.syncWindow()
		m.refreshText()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	if m.focus == paneText {
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) layout() {
	_, _, textH, textW := m.paneSizes()
	m.vp.Width = max(12, textW-2)
	m.vp.Height = max(3, textH-3)
}

func (m model) paneSizes() (leftW, bodyH, textH, textW int) {
	w, h := m.width, m.height
	help := 1
	bodyH = max(8, h-help)
	leftW = w * 32 / 100
	if leftW < 22 {
		leftW = 22
	}
	if leftW > 36 {
		leftW = 36
	}
	if w-leftW < 28 {
		leftW = max(16, w/3)
	}
	textW = max(20, w-leftW)
	textH = bodyH
	return
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.typing {
		return m.handleSearchKey(msg)
	}

	switch key {
	case "ctrl+c", "q":
		saveLast(m.b.Books[m.book].Name, m.chapter)
		return m, tea.Quit
	case "/":
		m.typing = true
		return m, nil
	case "esc":
		if m.query != "" {
			m.query = ""
			m.rebuild()
			m.cursor = m.findItem(m.book, m.chapter)
			m.syncWindow()
			m.refreshText()
		}
		return m, nil
	case "tab":
		if m.focus == paneIndex {
			m.focus = paneText
		} else {
			m.focus = paneIndex
		}
		return m, nil
	case "?":
		if m.status == helpText {
			m.status = ""
		} else {
			m.status = helpText
		}
		return m, nil
	}

	if m.focus == paneText {
		return m.handleTextKey(key)
	}
	return m.handleIndexKey(key)
}

const helpText = "j/k move  enter open  h/l chapter  tab pane  y copy  / filter  q quit"

func (m model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.typing = false
		if m.query != "" {
			m.query = ""
			m.rebuild()
			m.cursor = m.findItem(m.book, m.chapter)
			m.syncWindow()
			m.refreshText()
		}
		return m, nil
	case "enter":
		m.typing = false
		m.applyCursor()
		m.open[m.book] = true
		m.refreshText()
		m.vp.GotoTop()
		return m, nil
	case "backspace":
		if m.query == "" {
			m.typing = false
			return m, nil
		}
		r := []rune(m.query)
		m.query = string(r[:len(r)-1])
		m.rebuild()
		m.cursor = firstChapterOrBook(m.items)
		m.applyCursor()
		m.syncWindow()
		m.refreshText()
		m.vp.GotoTop()
		return m, nil
	case "up", "ctrl+k":
		m.move(-1)
		m.applyCursor()
		m.refreshText()
		m.vp.GotoTop()
		return m, nil
	case "down", "ctrl+j":
		m.move(1)
		m.applyCursor()
		m.refreshText()
		m.vp.GotoTop()
		return m, nil
	}
	if msg.Type == tea.KeyRunes {
		for _, r := range msg.Runes {
			if unicode.IsControl(r) {
				continue
			}
			m.query += string(r)
		}
		m.rebuild()
		m.cursor = firstChapterOrBook(m.items)
		m.applyCursor()
		m.syncWindow()
		m.refreshText()
		m.vp.GotoTop()
	}
	return m, nil
}

func firstSelectable(items []indexItem) int {
	for i, it := range items {
		if it.kind != kindSection {
			return i
		}
	}
	return 0
}

func firstChapterOrBook(items []indexItem) int {
	for i, it := range items {
		if it.kind == kindChapter {
			return i
		}
	}
	return firstSelectable(items)
}

func (m model) handleIndexKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		m.move(1)
	case "k", "up":
		m.move(-1)
	case "g", "home":
		m.cursor = firstSelectable(m.items)
		m.skipNonSelectable(1)
	case "G", "end":
		m.cursor = len(m.items) - 1
		m.skipNonSelectable(-1)
	case "ctrl+d", "pgdown":
		m.move(max(1, m.indexHeight()/2))
	case "ctrl+u", "pgup":
		m.move(-max(1, m.indexHeight()/2))
	case "enter", "l", "right":
		m.toggleOrSelect()
	case "h", "left":
		m.collapseOrPrev()
	case "n":
		m.shiftChapter(1)
		m.cursor = m.findItem(m.book, m.chapter)
	case "p":
		m.shiftChapter(-1)
		m.cursor = m.findItem(m.book, m.chapter)
	case "y":
		m.copyChapter()
		return m, nil
	default:
		return m, nil
	}
	m.applyCursor()
	m.syncWindow()
	m.refreshText()
	if key == "enter" || key == "l" || key == "right" || key == "n" || key == "p" || key == "h" || key == "left" {
		m.vp.GotoTop()
	}
	return m, nil
}

func (m model) handleTextKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "j", "down":
		m.vp.LineDown(1)
	case "k", "up":
		m.vp.LineUp(1)
	case "g", "home":
		m.vp.GotoTop()
	case "G", "end":
		m.vp.GotoBottom()
	case "ctrl+d", "pgdown":
		m.vp.HalfViewDown()
	case "ctrl+u", "pgup":
		m.vp.HalfViewUp()
	case "h", "left", "p":
		m.shiftChapter(-1)
		m.cursor = m.findItem(m.book, m.chapter)
		m.syncWindow()
		m.refreshText()
		m.vp.GotoTop()
	case "l", "right", "n":
		m.shiftChapter(1)
		m.cursor = m.findItem(m.book, m.chapter)
		m.syncWindow()
		m.refreshText()
		m.vp.GotoTop()
	case "y":
		m.copyChapter()
	case "esc":
		m.focus = paneIndex
	}
	return m, nil
}

func (m *model) move(delta int) {
	if len(m.items) == 0 {
		return
	}
	dir := 1
	if delta < 0 {
		dir = -1
	}
	for i := 0; i < abs(delta); i++ {
		m.cursor += dir
		if m.cursor < 0 {
			m.cursor = 0
			break
		}
		if m.cursor >= len(m.items) {
			m.cursor = len(m.items) - 1
			break
		}
		m.skipNonSelectable(dir)
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func (m *model) toggleOrSelect() {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return
	}
	it := m.items[m.cursor]
	if it.kind == kindBook {
		m.open[it.book] = !m.open[it.book]
		keep := it.book
		m.rebuild()
		m.cursor = m.findItem(keep, 1)
		if m.open[keep] {
			// land on chapter 1
			m.chapter = 1
			m.cursor = m.findItem(keep, 1)
		}
		return
	}
	if it.kind == kindChapter {
		m.focus = paneText
	}
}

func (m *model) collapseOrPrev() {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return
	}
	it := m.items[m.cursor]
	if it.kind == kindChapter {
		m.open[it.book] = false
		keep := it.book
		m.rebuild()
		m.cursor = m.findItem(keep, 1)
		return
	}
	if it.kind == kindBook && m.open[it.book] {
		m.open[it.book] = false
		keep := it.book
		m.rebuild()
		m.cursor = m.findItem(keep, 1)
	}
}

func (m *model) shiftChapter(delta int) {
	bk := m.b.Books[m.book]
	ch := m.chapter + delta
	if ch < 1 {
		if m.book == 0 {
			m.chapter = 1
			return
		}
		m.book--
		m.open[m.book] = true
		m.chapter = len(m.b.Books[m.book].Chapters)
		m.rebuild()
		return
	}
	if ch > len(bk.Chapters) {
		if m.book+1 >= len(m.b.Books) {
			m.chapter = len(bk.Chapters)
			return
		}
		m.book++
		m.open[m.book] = true
		m.chapter = 1
		m.rebuild()
		return
	}
	m.chapter = ch
	m.open[m.book] = true
	m.rebuild()
}

func (m *model) copyChapter() {
	bk, ch, _, ok := m.b.Get(m.book, m.chapter, 1)
	if !ok {
		return
	}
	text := bible.FormatChapter(bk, ch, 0)
	if err := copyText(text); err != nil {
		m.status = "copy failed: " + err.Error()
		return
	}
	m.status = "copied " + fmt.Sprintf("%s %d", bk.Name, ch.Num)
}

func (m model) indexHeight() int {
	_, bodyH, _, _ := m.paneSizes()
	h := bodyH - 3 // border + title
	if m.typing || m.query != "" {
		h--
	}
	return max(3, h)
}

func (m *model) syncWindow() {
	h := m.indexHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+h {
		m.offset = m.cursor - h + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m model) View() string {
	if m.width < 40 || m.height < 10 {
		return "widen the terminal"
	}
	leftW, bodyH, _, _ := m.paneSizes()
	left := m.renderIndex(leftW, bodyH)
	right := m.renderText(m.width-leftW, bodyH)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	help := m.renderHelp()
	return lipgloss.JoinVertical(lipgloss.Left, body, help)
}

func (m model) renderIndex(width, height int) string {
	innerW := max(8, width-2)
	innerH := max(3, height-2)
	title := badgeStyle.Render("KJV") + titleStyle.Width(max(0, innerW-5)).Render(" INDEX")

	h := innerH - 1
	filterLine := ""
	if m.typing || m.query != "" {
		prompt := "/"
		if m.typing {
			prompt = "/" + m.query + "█"
		} else {
			prompt = "/" + m.query
		}
		filterLine = searchStyle.Width(innerW).Render(prompt)
		h--
	}

	listH := max(1, h)
	var lines []string
	end := min(len(m.items), m.offset+listH)
	for i := m.offset; i < end; i++ {
		lines = append(lines, m.renderItem(m.items[i], i == m.cursor, innerW))
	}
	for len(lines) < listH {
		lines = append(lines, indexItemStyle.Width(innerW).Render(""))
	}
	content := title + "\n" + strings.Join(lines, "\n")
	if filterLine != "" {
		content += "\n" + filterLine
	}

	st := frameStyle
	if m.focus == paneIndex {
		st = activeFrameStyle
	}
	return st.Width(innerW).Height(innerH).Render(content)
}

func (m model) renderItem(it indexItem, selected bool, width int) string {
	var s string
	switch it.kind {
	case kindSection:
		s = "─ " + strings.ToUpper(it.label) + " "
		return indexMutedStyle.Width(width).Render(truncate(s, width))
	case kindBook:
		marker := "> "
		if m.open[it.book] {
			marker = "v "
		}
		if selected {
			marker = "* "
		}
		n := len(m.b.Books[it.book].Chapters)
		s = marker + it.label + fmt.Sprintf(" (%d)", n)
	case kindChapter:
		mark := "  "
		if selected {
			mark = "* "
		}
		s = mark + "  " + it.label
	}
	st := indexItemStyle
	if selected {
		st = indexSelStyle
	}
	return st.Width(width).Render(truncate(s, width))
}

func (m model) renderText(width, height int) string {
	innerW := max(8, width-2)
	innerH := max(3, height-2)

	bk, ch, _, ok := m.b.Get(m.book, m.chapter, 1)
	ref := ""
	if ok {
		ref = fmt.Sprintf("%s %d", bk.Name, ch.Num)
	}
	title := titleStyle.Width(innerW).Render(truncate(ref, innerW))

	body := m.vp.View()
	content := title + "\n" + body

	st := frameStyle
	if m.focus == paneText {
		st = activeFrameStyle
	}
	return st.Width(innerW).Height(innerH).Render(content)
}

func (m model) renderHelp() string {
	if m.status != "" {
		return statusStyle.Width(m.width).Render(truncate(m.status, m.width))
	}
	left := "j/k  enter  h/l ch  tab  y copy  /  q"
	if m.typing {
		left = "filter books/chapters  esc  enter"
	}
	return helpStyle.Width(m.width).Render(truncate(left, m.width))
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 1 {
		return string(r[:width])
	}
	return string(r[:width-1]) + "…"
}

func statePath() string {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(dir, "kjv-tui", "last")
}

func loadLast() string {
	if b, err := os.ReadFile(statePath()); err == nil {
		return strings.TrimSpace(string(b))
	}
	// Previous binary name, before the Arch package.
	old := strings.Replace(statePath(), "/kjv-tui/", "/kjv-fzf/", 1)
	if b, err := os.ReadFile(old); err == nil {
		return strings.TrimSpace(string(b))
	}
	return ""
}

func saveLast(book string, chapter int) {
	path := statePath()
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, []byte(fmt.Sprintf("%s %d\n", book, chapter)), 0o644)
}

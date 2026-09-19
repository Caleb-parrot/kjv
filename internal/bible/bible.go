package bible

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	_ "embed"
)

//go:embed kjv.tsv.gz
var rawGZ []byte

// Verse is a single KJV verse.
type Verse struct {
	Book    string
	Abbr    string
	BookN   int
	Chapter int
	Num     int
	Text    string
}

// Chapter is a numbered chapter inside a book.
type Chapter struct {
	Num    int
	Verses []Verse
}

// Book is a Bible book in file order (OT, Apocrypha, NT).
type Book struct {
	Name     string
	Abbr     string
	Number   int
	Chapters []Chapter
}

// Section groups books for the index.
type Section int

const (
	SectionOT Section = iota
	SectionApo
	SectionNT
)

func (s Section) Label() string {
	switch s {
	case SectionApo:
		return "Apocrypha"
	case SectionNT:
		return "New Testament"
	default:
		return "Old Testament"
	}
}

func SectionOf(bookN int) Section {
	switch {
	case bookN >= 40 && bookN <= 66:
		return SectionNT
	case bookN >= 67:
		return SectionApo
	default:
		return SectionOT
	}
}

// Bible is the loaded King James text.
type Bible struct {
	Books  []Book
	Verses []Verse
	byNorm map[string]int
}

// Load reads the embedded KJV TSV (Luke Smith / public domain, with Apocrypha).
func Load() (*Bible, error) {
	zr, err := gzip.NewReader(bytes.NewReader(rawGZ))
	if err != nil {
		return nil, fmt.Errorf("kjv: gzip: %w", err)
	}
	defer zr.Close()

	b := &Bible{byNorm: make(map[string]int)}
	sc := bufio.NewScanner(zr)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var cur *Book
	var chap *Chapter
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		f := strings.SplitN(line, "\t", 6)
		if len(f) < 6 {
			// Headings in the Apocrypha (e.g. Sirach) are not verses.
			continue
		}
		bookN, err1 := strconv.Atoi(f[2])
		chN, err2 := strconv.Atoi(f[3])
		vsN, err3 := strconv.Atoi(f[4])
		if err1 != nil || err2 != nil || err3 != nil {
			return nil, fmt.Errorf("kjv: bad numbers in %q", line)
		}
		v := Verse{
			Book:    f[0],
			Abbr:    f[1],
			BookN:   bookN,
			Chapter: chN,
			Num:     vsN,
			Text:    f[5],
		}
		b.Verses = append(b.Verses, v)

		if cur == nil || cur.Name != v.Book {
			b.Books = append(b.Books, Book{Name: v.Book, Abbr: v.Abbr, Number: v.BookN})
			cur = &b.Books[len(b.Books)-1]
			chap = nil
		}
		if chap == nil || chap.Num != v.Chapter {
			cur.Chapters = append(cur.Chapters, Chapter{Num: v.Chapter})
			chap = &cur.Chapters[len(cur.Chapters)-1]
		}
		chap.Verses = append(chap.Verses, v)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("kjv: read: %w", err)
	}
	if len(b.Books) == 0 {
		return nil, fmt.Errorf("kjv: empty text")
	}
	for i, bk := range b.Books {
		b.byNorm[norm(bk.Name)] = i
		b.byNorm[norm(bk.Abbr)] = i
	}
	// Common aliases that are not unique prefixes of the TSV names.
	aliases := map[string]string{
		"psalm": "psalms",
		"ps":    "psalms",
		"sos":   "song of solomon",
		"song":  "song of solomon",
		"acts":  "acts",
		"ssol":  "song of solomon",
		"qoh":   "ecclesiastes",
		"ecc":   "ecclesiastes",
		"jn":    "john",
		"mt":    "matthew",
		"mk":    "mark",
		"lk":    "luke",
		"phil":  "philippians",
		"php":   "philippians",
		"jas":   "james",
		"rev":   "revelation",
		"dt":    "deuteronomy",
		"nm":    "numbers",
		"ex":    "exodus",
		"gn":    "genesis",
		"is":    "isaiah",
		"jer":   "jeremiah",
		"ez":    "ezekiel",
	}
	for alias, target := range aliases {
		if _, ok := b.byNorm[alias]; ok {
			continue
		}
		if i, ok := b.findUnique(target); ok {
			b.byNorm[alias] = i
		}
	}
	return b, nil
}

func norm(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ".", "")
	s = strings.Join(strings.Fields(s), " ")
	s = strings.TrimPrefix(s, "the ")
	return s
}

func (b *Bible) findUnique(n string) (int, bool) {
	n = norm(n)
	if i, ok := b.byNorm[n]; ok {
		return i, true
	}
	found := -1
	for i, bk := range b.Books {
		nn := norm(bk.Name)
		na := norm(bk.Abbr)
		if nn == n || na == n || strings.HasPrefix(nn, n) || strings.HasPrefix(na, n) {
			if found != -1 && found != i {
				return 0, false
			}
			found = i
		}
	}
	if found == -1 {
		return 0, false
	}
	return found, true
}

// FindBook returns a book by name, abbreviation, or unique prefix.
func (b *Bible) FindBook(name string) *Book {
	n := norm(name)
	if n == "" {
		return nil
	}
	if i, ok := b.byNorm[n]; ok {
		return &b.Books[i]
	}
	if len(n) < 2 {
		return nil
	}
	found := -1
	for i, bk := range b.Books {
		nn := norm(bk.Name)
		na := norm(bk.Abbr)
		nameHit := strings.HasPrefix(nn, n)
		abbrHit := len(n) >= 2 && strings.HasPrefix(na, n)
		if nameHit || abbrHit {
			if found != -1 && found != i {
				return nil
			}
			found = i
		}
	}
	if found == -1 {
		return nil
	}
	return &b.Books[found]
}

// BookIndex is the 0-based index of a book pointer, or -1.
func (b *Bible) BookIndex(bk *Book) int {
	if bk == nil {
		return -1
	}
	for i := range b.Books {
		if b.Books[i].Name == bk.Name {
			return i
		}
	}
	return -1
}

// Get returns a chapter. verse is 1-based; 0 means the whole chapter (first verse).
func (b *Bible) Get(bookIdx, chapter, verse int) (Book, Chapter, Verse, bool) {
	if bookIdx < 0 || bookIdx >= len(b.Books) {
		return Book{}, Chapter{}, Verse{}, false
	}
	bk := b.Books[bookIdx]
	if chapter < 1 || chapter > len(bk.Chapters) {
		return bk, Chapter{}, Verse{}, false
	}
	ch := bk.Chapters[chapter-1]
	if len(ch.Verses) == 0 {
		return bk, ch, Verse{}, false
	}
	if verse < 1 {
		verse = 1
	}
	if verse > len(ch.Verses) {
		verse = len(ch.Verses)
	}
	return bk, ch, ch.Verses[verse-1], true
}

// Ref is a parsed scripture reference.
type Ref struct {
	BookIdx int
	Chapter int
	Verse   int
	HasVs   bool
}

func (r Ref) String(b *Bible) string {
	if r.BookIdx < 0 || r.BookIdx >= len(b.Books) {
		return ""
	}
	name := b.Books[r.BookIdx].Name
	if r.Chapter < 1 {
		return name
	}
	if !r.HasVs || r.Verse < 1 {
		return fmt.Sprintf("%s %d", name, r.Chapter)
	}
	return fmt.Sprintf("%s %d:%d", name, r.Chapter, r.Verse)
}

// ParseRef understands "John", "John 3", "Jn 3:16", "1 John 1:9".
func (b *Bible) ParseRef(s string) (Ref, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Ref{}, fmt.Errorf("empty reference")
	}
	s = strings.ReplaceAll(s, ":", " ")
	s = strings.ReplaceAll(s, ".", " ")
	parts := strings.Fields(s)
	for n := len(parts); n >= 1; n-- {
		// Don't treat a trailing number as part of the book name unless
		// the book is numbered ("1 John"), which is handled by FindBook.
		name := strings.Join(parts[:n], " ")
		if n < len(parts) && isAllDigits(parts[n-1]) && n > 1 {
			// "John 3 16" — last of prefix is a chapter, not "John 3" the book.
			continue
		}
		bk := b.FindBook(name)
		if bk == nil {
			continue
		}
		rest := parts[n:]
		r := Ref{BookIdx: b.BookIndex(bk), Chapter: 1, Verse: 1}
		if len(rest) >= 1 {
			ch, err := strconv.Atoi(rest[0])
			if err != nil || ch < 1 {
				continue
			}
			r.Chapter = ch
		}
		if len(rest) >= 2 {
			vs, err := strconv.Atoi(rest[1])
			if err != nil || vs < 1 {
				continue
			}
			r.Verse = vs
			r.HasVs = true
		}
		if _, _, _, ok := b.Get(r.BookIdx, r.Chapter, r.Verse); !ok {
			return Ref{}, fmt.Errorf("no such passage: %s", s)
		}
		return r, nil
	}
	return Ref{}, fmt.Errorf("unknown reference: %s", s)
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// FormatChapter prints a chapter as plain text (CLI).
func FormatChapter(bk Book, ch Chapter, wrap int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %d\n\n", bk.Name, ch.Num)
	for _, v := range ch.Verses {
		line := fmt.Sprintf("%d %s", v.Num, v.Text)
		if wrap > 8 {
			line = wrapIndent(line, wrap, digits(v.Num)+1)
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func digits(n int) int {
	if n < 10 {
		return 1
	}
	if n < 100 {
		return 2
	}
	return 3
}

func wrapIndent(s string, width, indent int) string {
	if width < indent+8 {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}
	pad := strings.Repeat(" ", indent)
	var b strings.Builder
	col := 0
	first := true
	for _, w := range words {
		need := len(w)
		if !first {
			need++
		}
		if !first && col+need > width {
			b.WriteByte('\n')
			b.WriteString(pad)
			b.WriteString(w)
			col = indent + len(w)
			continue
		}
		if !first {
			b.WriteByte(' ')
			col++
		}
		b.WriteString(w)
		col += len(w)
		first = false
	}
	return b.String()
}

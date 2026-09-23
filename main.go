package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/caleb-parrot/kjv/internal/bible"
	"github.com/caleb-parrot/kjv/internal/tui"
)

func main() {
	b, err := bible.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	args := os.Args[1:]
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(usage())
		return
	}
	if len(args) == 1 && args[0] == "-l" {
		for _, bk := range b.Books {
			fmt.Printf("%s (%s)  %d ch\n", bk.Name, bk.Abbr, len(bk.Chapters))
		}
		return
	}
	if len(args) > 0 {
		ref, err := b.ParseRef(strings.Join(args, " "))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		bk, ch, _, ok := b.Get(ref.BookIdx, ref.Chapter, 1)
		if !ok {
			fmt.Fprintln(os.Stderr, "no such chapter")
			os.Exit(1)
		}
		fmt.Print(bible.FormatChapter(bk, ch, 72))
		return
	}

	if err := tui.Run(b); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() string {
	return `kjv-tui — King James Bible, two-pane reader

  kjv-tui                 open the TUI (index left, chapter right)
  kjv-tui -l              list books
  kjv-tui John 3          print a chapter

TUI keys:
  j/k enter    move / open book
  h/l  n/p     previous / next chapter
  tab          switch index and text
  y            copy chapter
  /            filter books and chapters
  q            quit
`
}

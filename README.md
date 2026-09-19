# kjv-fzf

King James Bible reader for the terminal. Index on the left, full chapter on the right.

![kjv-fzf two-pane reader](screenshot.png)

Built for [Omarchy](https://omarchy.org/). It is a TUI app, not an Omarchy shell plugin, so it is not listed on the plugins marketplace.

The embedded text is the public-domain KJV (with Apocrypha) from [LukeSmithxyz/kjv](https://github.com/LukeSmithxyz/kjv).

## Install on Omarchy

Needs [Go](https://go.dev/). On Omarchy:

```bash
omarchy install dev-env go
```

Then build and add it to the app launcher:

```bash
git clone https://github.com/Caleb-parrot/kjv.git
cd kjv
go build -o ~/.local/bin/kjv-fzf .
omarchy tui install KJV ~/.local/bin/kjv-fzf tile accessories-dictionary
```

Open it from Super+Space as **KJV**, or run `kjv-fzf`.

Optional Super-menu row — add this to `~/.config/omarchy/extensions/omarchy-menu.jsonc`:

```jsonc
"kjv": {
  "icon": "",
  "label": "KJV",
  "description": "King James Bible",
  "action": "omarchy-launch-or-focus-tui --app-id=TUI.tile ~/.local/bin/kjv-fzf"
}
```

## Keys

| Key | Action |
| --- | --- |
| `j` / `k` | Move in the index, or scroll the chapter |
| `enter` | Open or close a book, or focus the chapter |
| `h` / `l` or `n` / `p` | Previous / next chapter |
| `tab` | Switch panes |
| `y` | Copy the chapter |
| `/` | Filter books and chapters |
| `q` | Quit (remembers your place) |

## CLI

```bash
kjv-fzf              # TUI
kjv-fzf -l           # list books
kjv-fzf John 3       # print a chapter
kjv-fzf Phi 1        # abbreviations work
```

## License

Code is Unlicense. The KJV text is public domain.

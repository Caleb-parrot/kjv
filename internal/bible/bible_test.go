package bible

import "testing"

func TestLoadGenesisAndJohn(t *testing.T) {
	b, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Books) < 66 {
		t.Fatalf("books=%d", len(b.Books))
	}
	_, _, v, ok := b.Get(0, 1, 1)
	if !ok {
		t.Fatal("missing Genesis 1:1")
	}
	if v.Text != "In the beginning God created the heaven and the earth." {
		t.Fatalf("Gen 1:1 = %q", v.Text)
	}

	ref, err := b.ParseRef("John 3:16")
	if err != nil {
		t.Fatal(err)
	}
	_, _, v, ok = b.Get(ref.BookIdx, ref.Chapter, ref.Verse)
	if !ok {
		t.Fatal("missing John 3:16")
	}
	want := "For God so loved the world, that he gave his only begotten Son, that whosoever believeth in him should not perish, but have everlasting life."
	if v.Text != want {
		t.Fatalf("John 3:16 = %q", v.Text)
	}

	ref, err = b.ParseRef("Phi 1:6")
	if err != nil {
		t.Fatal(err)
	}
	if ref.String(b) != "Philippians 1:6" {
		t.Fatalf("ref=%s", ref.String(b))
	}
	_, _, v, ok = b.Get(ref.BookIdx, ref.Chapter, ref.Verse)
	if !ok {
		t.Fatal("missing Phil 1:6")
	}
	if !contains(v.Text, "Being confident of this very thing") {
		t.Fatalf("Phil 1:6 = %q", v.Text)
	}

	if b.FindBook("acts") == nil || b.FindBook("Acts").Name != "The Acts" {
		t.Fatalf("acts book=%v", b.FindBook("acts"))
	}
	if b.FindBook("1 john") == nil {
		t.Fatal("1 john")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}

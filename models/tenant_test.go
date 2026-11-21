package models

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := []struct {
		name, want string
	}{
		{"Artisan Dupont", "artisan-dupont"},
		{"ARTISAN DUPONT", "artisan-dupont"},
		{"Crème & Co", "creme-co"},
		{"Étoile Noire", "etoile-noire"},
		{"naïve café!!", "naive-cafe"},
		{"  --  Padded  --  Name  ---  ", "padded-name"},
		{"multi    spaces", "multi-spaces"},
		{"123 Numbers 456", "123-numbers-456"},
		{"Ünïcödé Tëst", "unicode-test"},
		{"already-a-slug", "already-a-slug"},
		{"a", "a"},
		{"ß Sharp", "sharp"}, // ß has no ASCII decomposition: dropped
		{"Cœur de France", "c-ur-de-france"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Slugify(tc.name); got != tc.want {
				t.Fatalf("Slugify(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestValidSlug(t *testing.T) {
	cases := []struct {
		slug string
		want bool
	}{
		{"abc", true},
		{"artisan-dupont", true},
		{"a1", false},                   // too short: 2 chars
		{"abc-", false},                 // trailing dash
		{"-abc", false},                 // leading dash
		{"a-c", true},                   // 3 chars: minimum
		{"a" + filler(37) + "z", true},  // 39 chars
		{"a" + filler(38) + "z", true},  // 40 chars: maximum
		{"a" + filler(39) + "z", false}, // 41 chars
		{"has space", false},            // only letters, digits, dashes
		{"has.dot", false},
		{"has_underscore", false},
	}
	for _, tc := range cases {
		t.Run(tc.slug, func(t *testing.T) {
			if got := ValidSlug(tc.slug); got != tc.want {
				t.Fatalf("ValidSlug(%q) = %v, want %v", tc.slug, got, tc.want)
			}
		})
	}
}

func TestValidSlugReserved(t *testing.T) {
	for slug := range reservedSlugs {
		t.Run("reserved:"+slug, func(t *testing.T) {
			if ValidSlug(slug) {
				t.Fatalf("ValidSlug(%q) = true, want false for reserved slug", slug)
			}
		})
	}
	// A reserved prefix with a suffix remains valid.
	if !ValidSlug("admin-s") {
		t.Fatal("ValidSlug(\"admin-s\") = false, want true")
	}
}

// filler returns n dash-safe filler bytes to build fixed-length slugs.
func filler(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'x'
	}
	return string(b)
}

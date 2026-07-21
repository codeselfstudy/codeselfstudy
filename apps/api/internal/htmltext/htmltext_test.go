package htmltext

import "testing"

func TestConvert(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"anchor with text", `<a href="https://x.com/deal">Buy</a>`, "Buy (https://x.com/deal)"},
		{"anchor equals href", `<a href="https://x.com">https://x.com</a>`, "https://x.com"},
		{"anchor image only", `<a href="https://x.com/img"><img src="a.png"></a>`, "https://x.com/img"},
		{"anchor mailto keeps text only", `<a href="mailto:a@b.com">Email us</a>`, "Email us"},
		{"paragraphs break", `<p>One</p><p>Two</p>`, "One\nTwo"},
		{"br breaks", `a<br>b`, "a\nb"},
		{"heading then block", `<h1>Title</h1><div>Body</div>`, "Title\nBody"},
		{"whitespace collapses", "<p>  lots   of\n\n  space  </p>", "lots of space"},
		{"style and script dropped", `<style>.x{color:red}</style><script>bad()</script><p>Keep</p>`, "Keep"},
		{"inline elements flow", `<p>Hello <b>bold</b> world</p>`, "Hello bold world"},
		{"whitespace-only node separates", `<strong>50%</strong> <em>off</em>`, "50% off"},
		{"newline between inline tags", "<b>Big</b>\n<b>Sale</b>", "Big Sale"},
		{"uppercase scheme kept", `<a href="HTTPS://x.com/deal">Buy</a>`, "Buy (HTTPS://x.com/deal)"},
		{"list items break", `<ul><li>First</li><li>Second</li></ul>`, "First\nSecond"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert(tt.in)
			if err != nil {
				t.Fatalf("Convert: %v", err)
			}
			if got != tt.want {
				t.Errorf("Convert(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

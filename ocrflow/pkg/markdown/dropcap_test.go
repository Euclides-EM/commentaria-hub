package markdown

import "testing"

func TestParseDropcapPrefix(t *testing.T) {
	dropcap, remainder, ok := ParseDropcapPrefix(`{dropcap:P|lines=8|style=decorated|decoration="foliate ornamental initial in a square frame"}Vnctum`)
	if !ok {
		t.Fatal("ParseDropcapPrefix() did not recognize a valid drop cap")
	}
	if dropcap.Text != "P" || dropcap.Lines != "8" || dropcap.Style != "decorated" || dropcap.Decoration != "foliate ornamental initial in a square frame" {
		t.Fatalf("ParseDropcapPrefix() = %#v", dropcap)
	}
	if remainder != "Vnctum" {
		t.Fatalf("remainder = %q, want %q", remainder, "Vnctum")
	}
}

func TestExpandDropcaps(t *testing.T) {
	input := `1 {dropcap:P|lines=8|style=decorated|decoration="foliate"}Vnctum and {dropcap:A|lines=?|style=plain}Lterum`
	if got, want := ExpandDropcaps(input), "1 PVnctum and ALterum"; got != want {
		t.Fatalf("ExpandDropcaps() = %q, want %q", got, want)
	}
}

func TestExpandDropcapsPreservesMalformedSyntax(t *testing.T) {
	input := `{dropcap:P|style=plain}Ropositio`
	if got := ExpandDropcaps(input); got != input {
		t.Fatalf("ExpandDropcaps() = %q, want malformed input preserved", got)
	}
}

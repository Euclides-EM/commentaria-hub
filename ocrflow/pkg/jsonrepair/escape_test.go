package jsonrepair

import "testing"

func TestRepairInvalidStringEscapes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        string
		wantChanged bool
	}{
		{
			name:        "repairs invalid string escape",
			input:       `{"title":"D\. HENRION"}`,
			want:        `{"title":"D. HENRION"}`,
			wantChanged: true,
		},
		{
			name:        "repairs multiple invalid string escapes",
			input:       `{"path":"A\ B\ C","note":"keeps newline\n"}`,
			want:        `{"path":"A B C","note":"keeps newline\n"}`,
			wantChanged: true,
		},
		{
			name:        "preserves valid json escapes",
			input:       `{"quote":"He said \"hi\"","slash":"a\\b","unicode":"\u00e9"}`,
			want:        `{"quote":"He said \"hi\"","slash":"a\\b","unicode":"\u00e9"}`,
			wantChanged: false,
		},
		{
			name:        "ignores backslash outside strings",
			input:       `{"ok":true}\ trailing`,
			want:        `{"ok":true}\ trailing`,
			wantChanged: false,
		},
		{
			name:        "leaves trailing string backslash alone",
			input:       `{"title":"abc\`,
			want:        `{"title":"abc\`,
			wantChanged: false,
		},
		{
			name:        "no string escapes",
			input:       `{"title":"Elements"}`,
			want:        `{"title":"Elements"}`,
			wantChanged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed := RepairInvalidStringEscapes(tt.input)
			if got != tt.want {
				t.Fatalf("RepairInvalidStringEscapes() = %q, want %q", got, tt.want)
			}
			if changed != tt.wantChanged {
				t.Fatalf("RepairInvalidStringEscapes() changed = %v, want %v", changed, tt.wantChanged)
			}
		})
	}
}

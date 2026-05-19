package diagramcrops

import "testing"

func TestBaseKeyRemovesVolumeSuffix(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{name: "volume", key: "Paris_1615_vol2", want: "Paris_1615"},
		{name: "non volume", key: "Venice_1482", want: "Venice_1482"},
		{name: "missing volume number", key: "Paris_1615_vol", want: "Paris_1615_vol"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BaseKey(tt.key); got != tt.want {
				t.Fatalf("BaseKey(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestValidKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{key: "Paris_1615_vol2", want: true},
		{key: ".hidden", want: false},
		{key: "with/slash", want: false},
		{key: "", want: false},
	}

	for _, tt := range tests {
		if got := ValidKey(tt.key); got != tt.want {
			t.Fatalf("ValidKey(%q) = %t, want %t", tt.key, got, tt.want)
		}
	}
}

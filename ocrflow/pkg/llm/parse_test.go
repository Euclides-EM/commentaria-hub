package llm

import (
	"strings"
	"testing"
)

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr string
	}{
		{
			name:  "plain json object",
			input: `{"title":"Elements","author":"Euclid"}`,
			want: map[string]string{
				"title":  "Elements",
				"author": "Euclid",
			},
		},
		{
			name:  "json wrapped in prose",
			input: "Here is the result:\n\n{\"title\":\"Elements\"}\n\nThanks.",
			want: map[string]string{
				"title": "Elements",
			},
		},
		{
			name:  "braces inside string",
			input: `preface {"title":"{Elements}","author":"Euclid"} suffix`,
			want: map[string]string{
				"title":  "{Elements}",
				"author": "Euclid",
			},
		},
		{
			name:    "no json object",
			input:   "no structured output here",
			wantErr: "no json object found",
		},
		{
			name:    "incomplete json object",
			input:   `prefix {"title":"Elements"`,
			wantErr: "incomplete",
		},
		{
			name:    "json does not match target shape",
			input:   `{"title":["Elements"]}`,
			wantErr: "decode extracted json object",
		},
		{
			name:  "repairs invalid string escape",
			input: `{"title":"D\. HENRION","author":"Line\nBreak"}`,
			want: map[string]string{
				"title":  "D. HENRION",
				"author": "Line\nBreak",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJSON[map[string]string](tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("ParseJSON() error = nil, want substring %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("ParseJSON() error = %q, want substring %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseJSON() unexpected error = %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("ParseJSON() len = %d, want %d", len(got), len(tt.want))
			}
			for k, wantV := range tt.want {
				if got[k] != wantV {
					t.Fatalf("ParseJSON()[%q] = %q, want %q", k, got[k], wantV)
				}
			}
		})
	}
}

func TestExtractJSONObject(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{
			name:  "extracts first balanced object",
			input: `note {"a":{"b":"c"}} trailing {"ignored":true}`,
			want:  `{"a":{"b":"c"}}`,
		},
		{
			name:  "ignores braces in quoted strings",
			input: `x {"a":"value with } and { inside","b":"ok"} y`,
			want:  `{"a":"value with } and { inside","b":"ok"}`,
		},
		{
			name:    "missing object",
			input:   "[]",
			wantErr: "no json object found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractJSONObject(tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("extractJSONObject() error = nil, want substring %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("extractJSONObject() error = %q, want substring %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("extractJSONObject() unexpected error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("extractJSONObject() = %q, want %q", got, tt.want)
			}
		})
	}
}

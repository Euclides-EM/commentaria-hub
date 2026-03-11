package llm

import (
	"encoding/json"
	"fmt"
	"log"
)

func ParseJSON[T any](rawText string) (T, error) {
	var out T

	jsonText, err := extractJSONObject(rawText)
	if err != nil {
		return out, err
	}

	if err := json.Unmarshal([]byte(jsonText), &out); err != nil {
		return out, fmt.Errorf("decode extracted json object: %w", err)
	}

	return out, nil
}

func extractJSONObject(rawText string) (string, error) {
	start := -1
	depth := 0
	inString := false
	escaped := false

	for i, r := range rawText {
		if start == -1 {
			if r == '{' {
				start = i
				depth = 1
			}
			continue
		}

		if inString {
			if escaped {
				escaped = false
				continue
			}
			if r == '\\' {
				escaped = true
				continue
			}
			if r == '"' {
				inString = false
			}
			continue
		}

		switch r {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return rawText[start : i+1], nil
			}
		}
	}

	if start == -1 {
		log.Printf("llm parse json: no json object found in raw text: %q", rawText)
		return "", fmt.Errorf("no json object found in llm output")
	}

	log.Printf("llm parse json: incomplete json object in raw text: %q", rawText)
	return "", fmt.Errorf("json object in llm output is incomplete")
}

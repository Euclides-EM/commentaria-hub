package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/llm"
)

func parseIndexResponse(raw, volume string) ([]indexEntry, error) {
	entries, _, err := parseIndexResponseWithIssues(raw, volume)
	return entries, err
}

func parseIndexResponseWithIssues(raw, volume string) ([]indexEntry, []string, error) {
	response, err := parseStrictJSON[indexResponse](raw)
	if err != nil {
		return nil, nil, err
	}
	if response.Entries == nil {
		return nil, nil, errors.New("response requires an entries array")
	}
	entries := make([]indexEntry, 0, len(response.Entries))
	issues := make([]string, 0)
	for i := range response.Entries {
		compact := response.Entries[i]
		compact.Name = strings.TrimSpace(compact.Name)
		compact.Reference = strings.TrimSpace(compact.Reference)
		if isIndexSectionHeading(indexEntry{Name: compact.Name}) && len(compact.PageReferences) == 0 && compact.Reference == "" {
			continue
		}
		if compact.Name == "" || (len(compact.PageReferences) == 0) == (compact.Reference == "") {
			issues = append(issues, fmt.Sprintf(
				"entry %d requires name and exactly one of page_references or reference (name=%q, page_references=%d, reference=%q)",
				i+1, compact.Name, len(compact.PageReferences), compact.Reference,
			))
			continue
		}
		if compact.Reference != "" {
			entries = append(entries, indexEntry{Name: compact.Name, Reference: compact.Reference, Volume: volume})
			continue
		}
		for j := range compact.PageReferences {
			pageRef := compact.PageReferences[j]
			pageRef.PageNumber = strings.TrimSpace(pageRef.PageNumber)
			if pageRef.PageNumber == "" {
				issues = append(issues, fmt.Sprintf("entry %d page reference %d requires page_number (name=%q)", i+1, j+1, compact.Name))
				continue
			}
			entries = append(entries, indexEntry{Name: compact.Name, PageNumber: pageRef.PageNumber, IsBold: pageRef.IsBold, Volume: volume})
		}
	}
	return entries, issues, nil
}

func isIndexSectionHeading(entry indexEntry) bool {
	if entry.PageNumber != "" || entry.Reference != "" || utf8.RuneCountInString(entry.Name) != 1 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(entry.Name)
	return unicode.IsLetter(r)
}

func parseLettersResponse(raw, volume string) ([]letterEntry, error) {
	entries, _, err := parseLettersResponseWithIssues(raw, volume)
	return entries, err
}

func parseLettersResponseWithIssues(raw, volume string) ([]letterEntry, []string, error) {
	response, err := parseStrictJSON[lettersResponse](raw)
	if err != nil {
		return nil, nil, err
	}
	if response.Entries == nil {
		return nil, nil, errors.New("response requires an entries array")
	}
	entries := make([]letterEntry, 0, len(response.Entries))
	issues := make([]string, 0)
	for i := range response.Entries {
		entry := response.Entries[i]
		entry.LetterNumber = strings.TrimSpace(entry.LetterNumber)
		entry.LetterName = strings.TrimSpace(entry.LetterName)
		entry.PageNumber = strings.TrimSpace(entry.PageNumber)
		entry.Volume = volume
		if entry.LetterNumber == "" || entry.LetterName == "" || entry.PageNumber == "" {
			issues = append(issues, fmt.Sprintf(
				"entry %d requires letter_number, letter_name, and page_number (letter_number=%q, letter_name=%q, page_number=%q)",
				i+1, entry.LetterNumber, entry.LetterName, entry.PageNumber,
			))
			continue
		}
		entries = append(entries, entry)
	}
	return entries, issues, nil
}

func parseStrictJSON[T any](raw string) (T, error) {
	var result T
	object, err := llm.ParseJSON[json.RawMessage](raw)
	if err != nil {
		return result, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(object)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}

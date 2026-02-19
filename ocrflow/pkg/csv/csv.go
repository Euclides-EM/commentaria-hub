package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

func LoadCSVRecords(csvPath string) (header []string, rows []map[string]string, err error) {
	all, err := LoadCSV(csvPath)
	if err != nil {
		return nil, nil, err
	}
	if len(all) == 0 {
		return nil, nil, fmt.Errorf("CSV file empty: %s", csvPath)
	}
	header = all[0]
	for _, rec := range all[1:] {
		rows = append(rows, RowToMap(header, rec))
	}
	return header, rows, nil
}

func SaveCSVRecords(csvPath string, header []string, rows []map[string]string) error {
	// Ensure header contains all keys present in any row
	seen := make(map[string]struct{})
	for _, h := range header {
		seen[h] = struct{}{}
	}
	for _, m := range rows {
		for k := range m {
			if _, ok := seen[k]; !ok {
				header = append(header, k)
				seen[k] = struct{}{}
			}
		}
	}
	records := [][]string{header}
	for _, m := range rows {
		records = append(records, MapToRow(header, m))
	}
	return SaveCSV(csvPath, records)
}

// UpsertRow updates or appends a row by keyField value.
func UpsertRow(csvPath, keyField, key string, row map[string]string) error {
	header, rows, err := LoadCSVRecords(csvPath)
	if err != nil {
		return err
	}
	row[keyField] = key
	found := false
	for i, r := range rows {
		if r[keyField] == key {
			for k, v := range row {
				if slices.Contains(header, k) {
					rows[i][k] = v
				}
			}
			found = true
			break
		}
	}
	if !found {
		rows = append(rows, row)
	}
	return SaveCSVRecords(csvPath, header, rows)
}

// DeleteRows removes all rows where keyField == key.
func DeleteRows(csvPath, keyField, key string) error {
	header, rows, err := LoadCSVRecords(csvPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var kept []map[string]string
	for _, r := range rows {
		if r[keyField] != key {
			kept = append(kept, r)
		}
	}
	return SaveCSVRecords(csvPath, header, kept)
}

// ReplaceRowsForKey replaces all rows with keyField == key with newRows (same key in each).
func ReplaceRowsForKey(csvPath, keyField, key string, newRows []map[string]string) error {
	header, rows, err := LoadCSVRecords(csvPath)
	if err != nil {
		return err
	}
	var out []map[string]string
	done := false
	for _, r := range rows {
		if r[keyField] == key {
			if !done {
				for _, nr := range newRows {
					nr[keyField] = key
					out = append(out, nr)
				}
				done = true
			}
		} else {
			out = append(out, r)
		}
	}
	if !done {
		for _, nr := range newRows {
			nr[keyField] = key
			out = append(out, nr)
		}
	}
	return SaveCSVRecords(csvPath, header, out)
}

// BatchUpsertRows merges rows by key (and optionally secondary key like "field" for translations).
func BatchUpsertRows(csvPath, keyField string, secondaryKeyField string, newRows []map[string]string) error {
	if len(newRows) == 0 {
		return nil
	}
	header, rows, err := LoadCSVRecords(csvPath)
	if err != nil {
		return err
	}
	keyMap := make(map[string]map[string]string)
	for _, r := range newRows {
		k := r[keyField]
		if secondaryKeyField != "" && r[secondaryKeyField] != "" {
			k = k + ":" + r[secondaryKeyField]
		}
		keyMap[k] = r
	}
	var out []map[string]string
	seen := make(map[string]struct{})
	for _, r := range rows {
		k := r[keyField]
		if secondaryKeyField != "" && r[secondaryKeyField] != "" {
			k = k + ":" + r[secondaryKeyField]
		}
		if nr, ok := keyMap[k]; ok {
			out = append(out, nr)
			seen[k] = struct{}{}
			delete(keyMap, k)
		} else {
			out = append(out, r)
		}
	}
	for _, nr := range keyMap {
		out = append(out, nr)
	}
	return SaveCSVRecords(csvPath, header, out)
}

func LoadCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	return r.ReadAll()
}

func SaveCSV(path string, records [][]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err := w.WriteAll(records); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

// RowToMap converts a record to a map by header (first record).
func RowToMap(header, record []string) map[string]string {
	m := make(map[string]string, len(header))
	for i, h := range header {
		v := ""
		if i < len(record) {
			v = record[i]
		}
		m[h] = v
	}
	return m
}

// MapToRow writes a map to a record using header order; missing keys become empty.
func MapToRow(header []string, m map[string]string) []string {
	out := make([]string, len(header))
	for i, h := range header {
		out[i] = m[h]
	}
	return out
}

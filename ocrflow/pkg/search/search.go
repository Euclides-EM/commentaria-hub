package search

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Query struct {
	FieldsFilter     map[string][]string `json:"fields_filter"`
	FilterIncludes   map[string]bool     `json:"filter_includes"`
	TextSearch       string              `json:"text_search"`
	TextSearchFields []string            `json:"text_search_fields"`
	RangeFilter      map[string]Range    `json:"range_filter"`
	OrderBy          []OrderByOption     `json:"order_by"`
	Limit            int                 `json:"limit"`
	Offset           int                 `json:"offset"`
}

type Range struct {
	Min *float64 `json:"min"`
	Max *float64 `json:"max"`
	// Strict mode means that if the field exists but isn't numeric, the element fails the filter. Otherwise, non-numeric fields are ignored.
	Strict bool `json:"strict"`
}

type OrderByOption struct {
	Field      string `json:"field"`
	Descending bool   `json:"descending"`
}

func normalizeField(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "-", "")
	return s
}

func normalizeValue(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func fieldValue(e any, field string) (reflect.Value, bool) {
	v := reflect.ValueOf(e)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}, false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	t := v.Type()

	if f := v.FieldByName(field); f.IsValid() {
		return f, true
	}

	norm := normalizeField(field)

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)

		tag := sf.Tag.Get("json")
		tagName := strings.Split(tag, ",")[0]

		if tagName != "" && tagName != "-" {
			if tagName == field || normalizeField(tagName) == norm {
				return v.Field(i), true
			}
		}

		if normalizeField(sf.Name) == norm {
			return v.Field(i), true
		}
	}

	return reflect.Value{}, false
}

func valueToString(v reflect.Value) string {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	default:
		return fmt.Sprint(v.Interface())
	}
}

// valueToFloat64 tries hard to interpret a struct field as a number.
// Supports numeric kinds, pointers to numeric, and strings that parse.
func valueToFloat64(v reflect.Value) (float64, bool) {
	if !v.IsValid() {
		return 0, false
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0, false
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint()), true
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	case reflect.String:
		s := strings.TrimSpace(v.String())
		if s == "" {
			return 0, false
		}
		if idx := strings.Index(s, "/"); idx >= 0 {
			s = strings.TrimSpace(s[:idx])
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

func matchRange(num float64, r Range) bool {
	if r.Min != nil && num < *r.Min {
		return false
	}
	if r.Max != nil && num > *r.Max {
		return false
	}
	return true
}

func matchEditionYearRange(e any, r Range) (matched bool, handled bool) {
	isManuscriptField, ok := fieldValue(e, "isManuscript")
	if !ok || isManuscriptField.Kind() != reflect.Bool || !isManuscriptField.Bool() {
		return false, false
	}

	yearFromField, _ := fieldValue(e, "manuscriptYearFrom")
	yearToField, _ := fieldValue(e, "manuscriptYearTo")

	yearFrom, hasYearFrom := valueToFloat64(yearFromField)
	yearTo, hasYearTo := valueToFloat64(yearToField)

	if !hasYearFrom && !hasYearTo {
		return !r.Strict, true
	}
	if !hasYearFrom {
		yearFrom = yearTo
	}
	if !hasYearTo {
		yearTo = yearFrom
	}

	if r.Min != nil && yearTo < *r.Min {
		return false, true
	}
	if r.Max != nil && yearFrom > *r.Max {
		return false, true
	}

	return true, true
}

func editionYearSortValues(e any) (primary float64, secondary float64, ok bool) {
	if _, handled := matchEditionYearRange(e, Range{}); handled {
		yearFromField, _ := fieldValue(e, "manuscriptYearFrom")
		yearToField, _ := fieldValue(e, "manuscriptYearTo")

		yearFrom, hasYearFrom := valueToFloat64(yearFromField)
		yearTo, hasYearTo := valueToFloat64(yearToField)

		if !hasYearFrom && !hasYearTo {
			return 0, 0, false
		}
		if !hasYearFrom {
			yearFrom = yearTo
		}
		if !hasYearTo {
			yearTo = yearFrom
		}
		return yearFrom, yearTo, true
	}

	yearField, ok := fieldValue(e, "year")
	if !ok {
		return 0, 0, false
	}
	year, ok := valueToFloat64(yearField)
	if !ok {
		return 0, 0, false
	}
	return year, year, true
}

func (q Query) FilterFunc() func(e any) bool {
	return func(e any) bool {
		// Field filters (case-insensitive for string-like fields).
		// When filter_includes is not set for a field, default to true (allow-list: keep only if value matches).
		for field, allowed := range q.FieldsFilter {
			v, ok := fieldValue(e, field)
			if !ok {
				continue
			}

			// unwrap pointer once
			vv := v
			if vv.Kind() == reflect.Ptr {
				if vv.IsNil() {
					vv = reflect.Value{}
				} else {
					vv = vv.Elem()
				}
			}

			match := false
			if vv.IsValid() && vv.Kind() == reflect.Slice {
				// Slice (e.g. languages, corpus): match if any element is in allowed list
				for i := 0; i < vv.Len(); i++ {
					el := vv.Index(i)
					elStr := valueToString(el)
					for _, a := range allowed {
						if normalizeValue(elStr) == normalizeValue(a) {
							match = true
							break
						}
					}
					if match {
						break
					}
				}
			} else {
				// Single value
				val := valueToString(v)
				isStringLike := vv.IsValid() && vv.Kind() == reflect.String
				if isStringLike {
					nVal := normalizeValue(val)
					for _, a := range allowed {
						if nVal == normalizeValue(a) {
							match = true
							break
						}
					}
				} else {
					for _, a := range allowed {
						if val == a {
							match = true
							break
						}
					}
				}
			}

			// Default to include=true when not specified: fields_filter means "restrict to these values"
			include := true
			if got, set := q.FilterIncludes[field]; set {
				include = got
			}
			if include && !match {
				return false
			}
			if !include && match {
				return false
			}
		}

		// Range filters (inclusive bounds)
		for field, r := range q.RangeFilter {
			if normalizeField(field) == "year" {
				if matched, handled := matchEditionYearRange(e, r); handled {
					if !matched {
						return false
					}
					continue
				}
			}

			v, ok := fieldValue(e, field)
			if !ok {
				continue
			}

			num, ok := valueToFloat64(v)
			if !ok {
				// In strict mode, fail if field exists but isn't numeric.
				// Otherwise, ignore non-numeric fields.
				if r.Strict {
					return false
				}
				continue
			}

			if !matchRange(num, r) {
				return false
			}
		}

		// Text search
		if q.TextSearch != "" && len(q.TextSearchFields) > 0 {
			found := false
			search := strings.ToLower(q.TextSearch)

			for _, field := range q.TextSearchFields {
				v, ok := fieldValue(e, field)
				if !ok {
					continue
				}
				// allow *string too
				if v.Kind() == reflect.Ptr {
					if v.IsNil() {
						continue
					}
					v = v.Elem()
				}
				if v.Kind() != reflect.String {
					continue
				}

				if strings.Contains(strings.ToLower(v.String()), search) {
					found = true
					break
				}
			}

			if !found {
				return false
			}
		}

		return true
	}
}

func compareValues(v1, v2 reflect.Value) int {
	// unwrap pointers so sorting on *T works too
	if v1.Kind() == reflect.Ptr {
		if v1.IsNil() {
			return -1
		}
		v1 = v1.Elem()
	}
	if v2.Kind() == reflect.Ptr {
		if v2.IsNil() {
			return 1
		}
		v2 = v2.Elem()
	}

	switch v1.Kind() {
	case reflect.String:
		s1, s2 := v1.String(), v2.String()
		if s1 < s2 {
			return -1
		}
		if s1 > s2 {
			return 1
		}
	case reflect.Int, reflect.Int64, reflect.Int32:
		i1, i2 := v1.Int(), v2.Int()
		if i1 < i2 {
			return -1
		}
		if i1 > i2 {
			return 1
		}
	case reflect.Float32, reflect.Float64:
		f1, f2 := v1.Float(), v2.Float()
		if f1 < f2 {
			return -1
		}
		if f1 > f2 {
			return 1
		}
	default:
		s1, s2 := fmt.Sprint(v1.Interface()), fmt.Sprint(v2.Interface())
		if s1 < s2 {
			return -1
		}
		if s1 > s2 {
			return 1
		}
	}
	return 0
}

func (q Query) OrderByFunc() func(e1 any, e2 any) int {
	return func(e1 any, e2 any) int {
		for _, opt := range q.OrderBy {
			if normalizeField(opt.Field) == "year" {
				y1From, y1To, ok1 := editionYearSortValues(e1)
				y2From, y2To, ok2 := editionYearSortValues(e2)

				if ok1 || ok2 || (!ok1 && !ok2) {
					cmp := 0
					switch {
					case !ok1 && !ok2:
						cmp = 0
					case !ok1:
						cmp = 1
					case !ok2:
						cmp = -1
					case y1From < y2From:
						cmp = -1
					case y1From > y2From:
						cmp = 1
					case y1To < y2To:
						cmp = -1
					case y1To > y2To:
						cmp = 1
					}

					if opt.Descending && ok1 && ok2 {
						cmp = -cmp
					}
					if cmp != 0 {
						return cmp
					}
					continue
				}
			}

			v1, ok1 := fieldValue(e1, opt.Field)
			v2, ok2 := fieldValue(e2, opt.Field)
			if !ok1 || !ok2 {
				continue
			}

			cmp := compareValues(v1, v2)
			if opt.Descending {
				cmp = -cmp
			}

			if cmp != 0 {
				return cmp
			}
		}
		return 0
	}
}

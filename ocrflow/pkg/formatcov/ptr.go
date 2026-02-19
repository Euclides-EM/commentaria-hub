package formatcov

import (
	"strconv"
	"strings"

	"github.com/samber/lo"
)

func PtrToStr(x *string) string {
	return lo.IfF(x != nil, func() string { return *x }).Else("")
}

func StrToPtr(s string) *string {
	return lo.IfF(s != "", func() *string { return &s }).Else(nil)
}

func IntPtrToStr(x *int) string {
	return lo.IfF(x != nil, func() string { return strconv.Itoa(*x) }).Else("")
}

func IntOpt(s string) *int {
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return nil
	}
	return &n
}

// BoolPtrToStr returns "True", "False", or "" for nil.
func BoolPtrToStr(x *bool) string {
	if x == nil {
		return ""
	}
	return lo.If(*x, "True").Else("False")
}

// StrToBoolPtr returns a *bool from CSV-style "True"/"False", or nil if s is empty.
func StrToBoolPtr(s string) *bool {
	if s == "" {
		return nil
	}
	b := strings.TrimSpace(s) == "True"
	return &b
}

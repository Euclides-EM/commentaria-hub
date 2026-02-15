package formatcov

import (
	"strconv"

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

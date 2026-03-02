package xml

import (
	"fmt"
	"regexp"
)

func ModifyTag(origData []byte, tag string, modify func(v string) string) []byte {
	re := regexp.MustCompile(fmt.Sprintf(`<%s[^>]*>([^<]+)</%s>`, tag, tag))
	fixedData := re.ReplaceAllFunc(origData, func(match []byte) []byte {
		submatches := re.FindSubmatch(match)
		if len(submatches) <= 1 {
			return match
		}
		originalValue := submatches[1]
		modifiedValue := modify(string(originalValue))
		return []byte(fmt.Sprintf("<%s>%s</%s>", tag, modifiedValue, tag))
	})
	return fixedData
}

package xml

import (
	"fmt"
	"regexp"
)

func DeleteTag(origData []byte, tag string) []byte {
	return regexp.MustCompile(fmt.Sprintf(`\s*<%s>([^<]+)</%s>`, tag, tag)).ReplaceAll(origData, []byte(""))
}

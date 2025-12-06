package xml

import (
	"fmt"
	"regexp"
)

func ExtractTagValue(xmlContent, tag string) (string, error) {
	re := regexp.MustCompile(fmt.Sprintf(`<%s>([^<]+)</%s>`, tag, tag))
	matches := re.FindStringSubmatch(xmlContent)
	if len(matches) < 2 {
		return "", fmt.Errorf("tag <%s> not found in the provided XML content", tag)
	}
	return matches[1], nil
}

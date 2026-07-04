package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
)

type USTC struct{}

func NewUSTC() *USTC {
	return &USTC{}
}

// Lookup fetches edition data from the USTC website for the given id.
func (u *USTC) Lookup(ustcID int) (*model.USTC, error) {
	url := fmt.Sprintf("https://www.ustc.ac.uk/editions/%d", ustcID)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	body := make([]byte, 0, 512*1024)
	for {
		buf := make([]byte, 32*1024)
		n, _ := resp.Body.Read(buf)
		if n == 0 {
			break
		}
		body = append(body, buf[:n]...)
	}
	re := regexp.MustCompile(`data-page="([^"]+)"`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, nil
	}
	decoded := string(matches[1])
	decoded = strings.ReplaceAll(decoded, "&quot;", `"`)
	decoded = strings.ReplaceAll(decoded, "&amp;", "&")
	decoded = strings.ReplaceAll(decoded, "&lt;", "<")
	decoded = strings.ReplaceAll(decoded, "&gt;", ">")
	decoded = strings.ReplaceAll(decoded, "&#039;", "'")
	var pageData struct {
		Props struct {
			Edition       map[string]interface{} `json:"edition"`
			Digitisations []struct {
				URL string `json:"url"`
			} `json:"digitisations"`
		} `json:"props"`
	}
	if err := json.Unmarshal([]byte(decoded), &pageData); err != nil {
		return nil, err
	}
	ed := pageData.Props.Edition
	if ed == nil {
		return nil, nil
	}
	result := &model.USTC{USTCId: ustcID}
	for i := 1; i <= 8; i++ {
		k := fmt.Sprintf("author_name_%d", i)
		if v, _ := ed[k].(string); v != "" {
			for _, n := range strings.Split(v, ";") {
				n = strings.TrimSpace(n)
				if n != "" {
					result.Authors = append(result.Authors, formatUSTCName(n))
				}
			}
		}
	}
	if v, _ := ed["std_title"].(string); v != "" {
		result.ShortTitle = v
	}
	for i := 1; i <= 4; i++ {
		k := fmt.Sprintf("printer_name_%d", i)
		if v, _ := ed[k].(string); v != "" {
			for _, n := range strings.Split(v, ";") {
				n = strings.TrimSpace(n)
				if n != "" {
					result.Publishers = append(result.Publishers, formatUSTCName(n))
				}
			}
		}
	}
	if v, _ := ed["place"].(string); v != "" {
		result.City = &v
	}
	if v, _ := ed["year"].(string); v != "" {
		y, _ := strconv.Atoi(v)
		result.Year = &y
	}
	for i := 1; i <= 4; i++ {
		k := fmt.Sprintf("language_%d", i)
		if v, _ := ed[k].(string); v != "" {
			result.Languages = append(result.Languages, v)
		}
	}
	for _, d := range pageData.Props.Digitisations {
		if d.URL != "" {
			result.Digitizations = append(result.Digitizations, d.URL)
		}
	}
	if v, _ := ed["format"].(string); v != "" {
		result.Format = &v
	}
	return result, nil
}

func formatUSTCName(name string) string {
	// Remove parenthetical suffixes like " (printer)"
	re := regexp.MustCompile(`\s*\([^)]*\)\s*`)
	clean := strings.TrimSpace(re.ReplaceAllString(name, ""))
	if strings.Contains(clean, ",") {
		parts := strings.SplitN(clean, ",", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1]) + " " + strings.TrimSpace(parts[0])
		}
	}
	return clean
}

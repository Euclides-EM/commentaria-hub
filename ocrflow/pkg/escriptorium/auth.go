package escriptorium

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

func getAuthToken(username, password, basePath string) (string, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// 1. GET login page
	loginURL := basePath + "login/"
	resp, err := client.Get(loginURL)
	if err != nil {
		return "", fmt.Errorf("get login page: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("parse login page: %w", err)
	}

	// 2. extract CSRF token
	csrf := ""
	doc.Find(`input[name="csrfmiddlewaretoken"]`).Each(func(i int, s *goquery.Selection) {
		val, ok := s.Attr("value")
		if ok {
			csrf = val
		}
	})
	if csrf == "" {
		return "", errors.New("csrf token not found")
	}

	// 3. POST login the same way as Python
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	form.Set("csrfmiddlewaretoken", csrf)

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", loginURL)

	resp, err = client.Do(req)
	if err != nil {
		return "", fmt.Errorf("post login: %w", err)
	}
	defer resp.Body.Close()

	// 4. GET profile/apikey
	apiURL := basePath + "profile/apikey/"
	resp, err = client.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("get api key page: %w", err)
	}
	defer resp.Body.Close()

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("parse api page: %w", err)
	}

	// 5. extract key
	apiKey := ""
	doc.Find(`button#api-key-clipboard`).Each(func(i int, s *goquery.Selection) {
		val, ok := s.Attr("data-key")
		if ok {
			apiKey = val
		}
	})

	if apiKey == "" {
		return "", errors.New("api key not found")
	}

	return apiKey, nil
}

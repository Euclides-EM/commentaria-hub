package escriptorium

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

func get[T any](c *Client, path string) ([]T, error) {
	u := c.basePath + strings.TrimPrefix(path, "/")
	r, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %v", err)
	}
	res, err := c.Do(r)
	if err != nil {
		return nil, fmt.Errorf("unable to preform request: %v", err)
	}
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)

	var results struct {
		Results []T    `json:"results"`
		Next    string `json:"next"`
	}
	if err = decoder.Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	if results.Next != "" {
		// results.Next
		up, err := url.Parse(results.Next)
		if err != nil {
			return nil, fmt.Errorf("unable to parse next url: %v", err)
		}
		more, err := get[T](c, up.Path+"?"+up.RawQuery)
		if err != nil {
			return nil, err
		}
		results.Results = append(results.Results, more...)
	}
	return results.Results, nil
}

func post[T any](c *Client, path string, body T) (T, error) {
	var result T

	//u, err := url.JoinPath(c.basePath, path)
	//if err != nil {
	//	return result, fmt.Errorf("unable to create url: %v", err)
	//}
	u := c.basePath + path
	b, err := json.Marshal(body)
	if err != nil {
		return result, fmt.Errorf("unable to marshal body: %v", err)
	}

	r, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return result, fmt.Errorf("unable to create request: %v", err)
	}
	r.Header.Set("Content-Type", "application/json")

	res, err := c.Do(r)
	if err != nil {
		return result, fmt.Errorf("unable to preform request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return result, fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, string(b))
	}

	decoder := json.NewDecoder(res.Body)

	if err = decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("failed to decode response: %w", err)
	}
	return result, nil
}

func doDelete(c *Client, path string) error {
	u, err := url.JoinPath(c.basePath, path)
	if err != nil {
		return fmt.Errorf("unable to create url: %v", err)
	}
	r, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return fmt.Errorf("unable to create request: %v", err)
	}
	res, err := c.Do(r)
	if err != nil {
		return fmt.Errorf("unable to preform request: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	return nil
}

func (c *Client) UploadAnnotation(documentIdentifier, filePath string) error {
	documentId, err := c.inferDocumentID(documentIdentifier)
	if err != nil {
		return err
	}
	additionalParams := map[string]string{
		"task":     "import-xml",
		"name":     "",
		"document": fmt.Sprintf("%d", documentId),
	}
	return upload(c, fmt.Sprintf("/api/documents/%d/imports/", documentId), "upload_file", filePath, additionalParams)
}

func upload(c *Client, urlPath string, fileFieldName, filePath string, additionalParams map[string]string) error {
	u, err := url.JoinPath(c.basePath, urlPath)
	if err != nil {
		return fmt.Errorf("unable to create url: %w", err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if additionalParams != nil {
		for key, val := range additionalParams {
			_ = writer.WriteField(key, val)
		}
	}

	fw, err := writer.CreateFormFile(fileFieldName, path.Base(filePath))
	if err != nil {
		return fmt.Errorf("unable to create form file: %w", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("unable to open file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(fw, f)
	if err != nil {
		return fmt.Errorf("unable to copy file: %w", err)
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, u, &buf)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("unable to upload file: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	fmt.Println("Status:", res.Status)
	fmt.Println(string(body))
	return nil
}

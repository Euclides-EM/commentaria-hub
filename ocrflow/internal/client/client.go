package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
)

type Client struct {
	apiKey   string
	basePath string

	httpClient *http.Client
}

func NewClient(apiKey string, basePath string) *Client {
	return &Client{
		apiKey:     apiKey,
		basePath:   strings.TrimSuffix(basePath, "/") + "/",
		httpClient: &http.Client{},
	}
}

func (c *Client) apiURL(path string) (string, error) {
	return url.JoinPath(c.basePath, "api/v1", path)
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	return c.httpClient.Do(req)
}

func (c *Client) Authenticate() error {
	u, err := c.apiURL("auth/validate")
	if err != nil {
		return fmt.Errorf("commentaria auth url: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("commentaria auth failed: %d %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) CreateDataset(ds *model.Dataset) (*model.Dataset, error) {
	u, err := c.apiURL("datasets")
	if err != nil {
		return nil, fmt.Errorf("commentaria create dataset url: %w", err)
	}
	body, err := json.Marshal(ds)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("commentaria create dataset: %d %s", resp.StatusCode, string(respBody))
	}
	var out model.Dataset
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("commentaria decode dataset: %w", err)
	}
	return &out, nil
}

func (c *Client) UploadAnnotation(datasetID string, upm *annotation.UploadMetadata, altoDir string) (*annotation.Annotation, error) {
	tmpZip, err := os.CreateTemp("", "commentaria-alto-*.zip")
	if err != nil {
		return nil, fmt.Errorf("commentaria create temp zip: %w", err)
	}
	tmpPath := tmpZip.Name()
	defer os.Remove(tmpPath)
	if err := tmpZip.Close(); err != nil {
		return nil, err
	}
	if err := futils.Zip(altoDir, tmpPath); err != nil {
		return nil, fmt.Errorf("commentaria zip alto dir: %w", err)
	}
	zipFile, err := os.Open(tmpPath)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	u, err := c.apiURL("datasets/" + datasetID + "/annotations/fromzip")
	if err != nil {
		return nil, fmt.Errorf("commentaria upload annotation url: %w", err)
	}
	q := url.Values{}
	q.Set("format", "ALTO")
	q.Set("name", upm.Name)
	q.Set("description", upm.Description)
	q.Set("segmented", strconv.FormatBool(upm.Segmented))
	q.Set("ocred", strconv.FormatBool(upm.Ocred))
	q.Set("ground_truth", strconv.FormatBool(upm.GroundTruth))
	if upm.OriginAnnotationID != "" {
		q.Set("origin_annotation_id", upm.OriginAnnotationID)
	}
	if upm.OCRModelID != "" {
		q.Set("ocr_model_id", upm.OCRModelID)
	}
	if upm.SegmentModelID != "" {
		q.Set("segment_model_id", upm.SegmentModelID)
	}
	u += "?" + q.Encode()

	var buf bytes.Buffer
	mp := multipart.NewWriter(&buf)
	part, err := mp.CreateFormFile("file", filepath.Base(tmpPath))
	if err != nil {
		return nil, fmt.Errorf("commentaria create form file: %w", err)
	}
	if _, err := io.Copy(part, zipFile); err != nil {
		return nil, fmt.Errorf("commentaria copy zip to form: %w", err)
	}
	if err := mp.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mp.FormDataContentType())
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("commentaria upload annotation: %d %s", resp.StatusCode, string(respBody))
	}
	var out annotation.Annotation
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("commentaria decode annotation: %w", err)
	}
	return &out, nil
}

func (c *Client) GetFacsimilesByEditionID(editionID string) ([]*model.Facsimile, error) {
	u, err := c.apiURL("/facsimilies")
	if err != nil {
		return nil, fmt.Errorf("commentaria get facsimiles url: %w", err)
	}
	q := url.Values{}
	q.Set("edition_id", editionID)
	u += "?" + q.Encode()
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("commentaria get facsimiles: %d %s", resp.StatusCode, string(respBody))
	}
	var out []*model.Facsimile
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("commentaria decode facsimiles: %w", err)
	}
	return out, nil
}

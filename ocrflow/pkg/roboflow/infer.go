package roboflow

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/formatcov"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Recognize(imgPath string, outPath string, modelName string, apiKey string, filenames []string) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		errCh <- infer(imgPath, outPath, modelName, apiKey, filenames)
	}()

	return errCh
}

func infer(imgPath string, outPath string, modelName string, apiKey string, filenames []string) error {
	if err := os.MkdirAll(outPath, 0o755); err != nil {
		return err
	}
	log.Printf("Starting inference on %d images using model %s", len(filenames), modelName)
	for _, filename := range filenames {
		inputFile, err := futils.SafeJoin(imgPath, filename)
		if err != nil {
			return err
		}
		modelURL := fmt.Sprintf("https://serverless.roboflow.com/%s?api_key=%s", modelName, apiKey)

		data, err := os.ReadFile(inputFile)
		if err != nil {
			return err
		}

		encoded := base64.StdEncoding.EncodeToString(data)

		req, err := http.NewRequest("POST", modelURL, bytes.NewBuffer([]byte(encoded)))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "text/plain")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		altoFile, err := formatcov.Roboflow2ALTO(string(body), filepath.Base(filename), "eSc_dummypage_")
		if err != nil {
			return fmt.Errorf("failed to convert to ALTO: %v", err)
		}

		altoPath := filepath.Join(outPath, strings.TrimSuffix(filepath.Base(filename), ".png")+".xml")
		if err = os.WriteFile(altoPath, altoFile, 0o644); err != nil {
			return fmt.Errorf("write alto file failed: %v", err)
		}

		log.Printf("Processed image %s, saved to %s", filename, altoPath)
	}

	return nil
}

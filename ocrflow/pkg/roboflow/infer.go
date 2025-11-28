package roboflow

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func Recognize(imgPath string, outPath string, modelName string, filenames []string, apiKey string) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		errCh <- infer(imgPath, outPath, modelName, filenames, apiKey)
	}()

	return errCh
}

func infer(imgPath string, outPath string, modelName string, filenames []string, apiKey string) error {
	if err := os.MkdirAll(filepath.Join(outPath, "test"), 0o755); err != nil {
		return err
	}

	jsonStrs := make(map[string]string)
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

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		jsonStrs[filename] = string(body)

		// Write response to output file
		outImgPath := filepath.Join(outPath, "test", filename)

		// copy the img
		futils.CopyFile(inputFile, outImgPath)
		// my_dataset/
		//├── train/
		//│   ├── image1.jpg
		//│   ├── image2.jpg
		//│   ├── image3.jpg
		//│   └── _annotations.coco.json
		//├── valid/
		//│   ├── image4.jpg
		//│   └── _annotations.coco.json
		//└── test/
		//    ├── image5.jpg
		//    └── _annotations.coco.json
	}

	asCoco, err := formatcov.Roboflow2Coco(jsonStrs)
	if err != nil {
		return err
	}
	outputPath := filepath.Join(outPath, "test", "_annotations.coco.json")
	err = os.WriteFile(outputPath, []byte(asCoco), 0o644)
	if err != nil {
		return err
	}
	return nil
}

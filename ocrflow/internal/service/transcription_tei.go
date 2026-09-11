package service

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/markdown"
	tei2 "github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei"
	teimodel "github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei/model"
)

const transcriptionAlternativeDivType = "transcription-alternative"

func buildTEIFromTranscriptionFiles(
	pageKey string,
	files []filesys.TranscriptionFile,
	entities []tei2.EntityItem,
	imageURL string,
	biblMetadata *teimodel.BiblFull,
) (*teimodel.TEI, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no transcription files provided")
	}

	var combined *teimodel.TEI
	for i, file := range files {
		filePageKey := pageKey
		fileEntities := entities
		fileImageURL := imageURL
		if i > 0 {
			filePageKey += "_alternative_" + strconv.Itoa(i)
			fileEntities = nil
			fileImageURL = ""
		}

		doc, err := buildTEIFromTranscriptionFile(filePageKey, file, fileEntities, fileImageURL, biblMetadata)
		if err != nil {
			return nil, fmt.Errorf("build TEI from transcription file %s: %w", file.Path, err)
		}
		if len(doc.Text.Body.Divs) == 0 {
			return nil, fmt.Errorf("transcription file %s produced no text divisions", file.Path)
		}

		if i == 0 {
			combined = doc
			continue
		}

		for _, div := range doc.Text.Body.Divs {
			if div.Type != "transcription" {
				continue
			}
			div.Type = transcriptionAlternativeDivType
			div.N = file.Label
			clearAlternativeFacsimileReferences(&div)
			combined.Text.Body.Divs = append(combined.Text.Body.Divs, div)
		}
	}

	return combined, nil
}

func buildTEIFromTranscriptionFile(
	pageKey string,
	file filesys.TranscriptionFile,
	entities []tei2.EntityItem,
	imageURL string,
	biblMetadata *teimodel.BiblFull,
) (*teimodel.TEI, error) {
	switch file.Format {
	case filesys.TranscriptionFileFormatALTO:
		a, err := alto.LoadFromFile(file.Path)
		if err != nil {
			return nil, fmt.Errorf("load ALTO: %w", err)
		}
		return tei2.BuildTEIFromALTO(pageKey, a, entities, imageURL, biblMetadata)
	case filesys.TranscriptionFileFormatMarkdown:
		data, err := os.ReadFile(file.Path)
		if err != nil {
			return nil, fmt.Errorf("read Markdown: %w", err)
		}
		return tei2.BuildTEIFromMarkdown(pageKey, &markdown.Markdown{Content: string(data)}, biblMetadata)
	case filesys.TranscriptionFileFormatText:
		data, err := os.ReadFile(file.Path)
		if err != nil {
			return nil, fmt.Errorf("read TXT: %w", err)
		}
		return tei2.BuildTEIFromLines(pageKey, tei2.Lines{
			TranscriptionLines: strings.Split(string(data), "\n"),
		}, entities, imageURL, biblMetadata)
	default:
		return nil, fmt.Errorf("unsupported transcription format %q", file.Format)
	}
}

func clearAlternativeFacsimileReferences(div *teimodel.Div) {
	for i := range div.Abs {
		div.Abs[i].Facs = ""
		for j := range div.Abs[i].Lines {
			div.Abs[i].Lines[j].Facs = ""
		}
	}
}

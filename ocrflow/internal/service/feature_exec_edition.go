package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/llm"
)

func (fe *Execution) editionApplyFunc(editionKey string, actions *executionActions, execID string) applyFunc {
	return func() ([]*feature.Result, error) {
		edition, err := fe.editionSvc.GetEditionByID(editionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata for edition %s: %w", editionKey, err)
		}

		results := make([]*feature.Result, 0)
		var execErrs []error
		if len(actions.promptRevisions) > 0 {
			for _, group := range groupPromptRevisionsByAIConfig(actions.promptRevisions, actions.promptFeatures) {
				promptResults, err := fe.editionPromptApplyFunc(edition, group.revisions, group.features, execID)()
				if err != nil {
					execErrs = append(execErrs, err)
				}
				results = append(results, promptResults...)
			}
		}
		return results, errors.Join(execErrs...)
	}
}

func formatEditionInfo(ed *model.Edition) string {
	var b strings.Builder

	if ed.IsManuscript {
		intro := "This is a manuscript"
		if ed.ManuscriptClass != "" {
			intro += " of the " + ed.ManuscriptClass + " class"
			if ed.ManuscriptSubclass != nil {
				intro += " (" + *ed.ManuscriptSubclass + ")"
			}
		}
		switch {
		case ed.ManuscriptYearFrom != nil && ed.ManuscriptYearTo != nil:
			intro += fmt.Sprintf(", dated approximately %d–%d", *ed.ManuscriptYearFrom, *ed.ManuscriptYearTo)
		case ed.ManuscriptYearFrom != nil:
			intro += fmt.Sprintf(", dated from approximately %d", *ed.ManuscriptYearFrom)
		case ed.ManuscriptYearTo != nil:
			intro += fmt.Sprintf(", dated up to approximately %d", *ed.ManuscriptYearTo)
		}
		if ed.ShortTitle != "" {
			intro += ", known as \"" + ed.ShortTitle + "\""
		}
		b.WriteString(intro + ".\n")
	} else {
		intro := "This is a printed edition"
		if ed.ShortTitle != "" {
			intro += " known as \"" + ed.ShortTitle + "\""
		}
		var where []string
		if ed.Year != nil {
			where = append(where, "published in "+*ed.Year)
		}
		if len(ed.Cities) > 0 {
			where = append(where, "in "+strings.Join(ed.Cities, " and "))
		}
		if len(ed.Languages) > 0 {
			where = append(where, "in "+strings.Join(ed.Languages, " and "))
		}
		if len(where) > 0 {
			intro += ", " + strings.Join(where, ", ")
		}
		b.WriteString(intro + ".\n")

		var people []string
		if len(ed.Editor) > 0 {
			people = append(people, strings.Join(ed.Editor, " and ")+" edited it")
		}
		if len(ed.Publisher) > 0 {
			people = append(people, "published by "+strings.Join(ed.Publisher, " and "))
		}
		if len(people) > 0 {
			b.WriteString(strings.Join(people, "; ") + ".\n")
		}

		if ed.Format != nil || ed.Volumes != nil {
			var phys []string
			if ed.Format != nil {
				phys = append(phys, fmt.Sprintf("%d°", *ed.Format))
			}
			if ed.Volumes != nil {
				phys = append(phys, fmt.Sprintf("%d volume(s)", *ed.Volumes))
			}
			b.WriteString("Physical format: " + strings.Join(phys, ", ") + ".\n")
		}

		if ed.ReprintOf != nil {
			b.WriteString("This edition is a reprint.\n")
		}

		if ed.Title != nil {
			b.WriteString("\nTitle page reads: " + *ed.Title + "\n")
		}
		if ed.Imprint != nil {
			b.WriteString("\nImprint reads: " + *ed.Imprint + "\n")
		}
		if ed.Colophon != nil {
			b.WriteString("\nColophon reads: " + *ed.Colophon + "\n")
		}
		if ed.Frontispiece != nil {
			b.WriteString("\nFrontispiece reads: " + *ed.Frontispiece + "\n")
		}
	}

	if ed.IsElements {
		content := "\nThe edition covers Euclid's Elements"
		if len(ed.Books) > 0 {
			bookStrs := make([]string, len(ed.Books))
			for i, n := range ed.Books {
				bookStrs[i] = fmt.Sprintf("%d", n)
			}
			content += ", specifically books " + strings.Join(bookStrs, ", ")
		}
		content += "."
		b.WriteString(content + "\n")
	}
	if len(ed.AdditionalContent) > 0 {
		b.WriteString("Additional content included: " + strings.Join(ed.AdditionalContent, ", ") + ".\n")
	}
	if ed.HasDiagrams != nil {
		if *ed.HasDiagrams {
			b.WriteString("The edition contains diagrams.\n")
		} else {
			b.WriteString("The edition does not contain diagrams.\n")
		}
	}
	if ed.Notes != "" {
		b.WriteString("\nNotes: " + ed.Notes + "\n")
	}

	return strings.TrimSpace(b.String())
}

func (fe *Execution) editionPromptApplyFunc(ed *model.Edition, frs []*feature.Revision, fes []*feature.Feature, execID string) applyFunc {
	return func() ([]*feature.Result, error) {
		aiProvider := frs[0].AIProvider
		aiModel := frs[0].AIModel
		featureNameToIndex, definitions, outputFormat := buildPromptComponents(frs, fes)
		prompt := fmt.Sprintf(`You are an AI agent designed to extract structured metadata about historical textbook editions.

You will be given structured metadata about a specific edition.

Your task is to answer specific questions about the edition based on the provided metadata and return them as a JSON object.
Each field should contain the exact value(s) from the input metadata, with no modifications, rephrasing, or interpretation.
Some information may apply to more than one field, so you may return the same values in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  %s
}

Definitions:
%s

Edition metadata:
%s
`, outputFormat, strings.Join(definitions, "\n"), formatEditionInfo(ed))

		contextDesc := fmt.Sprintf("edition %s", ed.Key)
		rawResponse, err := fe.llmClient.Exec(aiProvider.ToLLMAIProvider(), aiModel, prompt, "")
		if err != nil {
			return nil, fmt.Errorf("failed to execute LLM prompt for %s using %s/%s: %w", contextDesc, aiProvider, aiModel, err)
		}
		rawFields, err := llm.ParseJSON[map[string]json.RawMessage](rawResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse LLM response for %s: %w", contextDesc, err)
		}
		return parseLLMResults(rawFields, frs, fes, featureNameToIndex, execID, feature.NewEditionExecScope(), ed.Key, contextDesc)
	}
}

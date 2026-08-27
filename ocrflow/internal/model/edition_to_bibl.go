package model

import (
	"fmt"
	"strconv"
	"strings"

	teim "github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei/model"
)

func EditionToBiblFull(ed *Edition) *teim.BiblFull {
	if ed == nil {
		return nil
	}

	bf := &teim.BiblFull{
		TitleStmt:       BiblTitleStmt(ed),
		PublicationStmt: BiblPublicationStmt(ed),
		Extent:          buildExtent(ed),
		NotesStmt:       buildNotesStmt(ed),
	}

	if isEmptyBiblFull(bf) {
		return nil
	}

	return bf
}

func BiblTitleStmt(ed *Edition) *teim.BiblTitleStmt {
	var titles []teim.Title

	if s := trimPtr(ed.Title); s != "" {
		titles = append(titles, teim.Title{
			Type:    "full",
			Content: s,
		})
	}

	if ed.ShortTitle != "" {
		titles = append(titles, teim.Title{
			Type:    "short",
			Content: ed.ShortTitle,
		})
	}

	if s := trimPtr(ed.TitleEN); s != "" {
		titles = append(titles, teim.Title{
			Type:    "translated",
			Lang:    "en",
			Content: s,
		})
	}

	editors := nonEmptyStrings(ed.Editor)

	if len(titles) == 0 && len(editors) == 0 {
		return nil
	}

	return &teim.BiblTitleStmt{
		Titles: titles,
		Editor: editors,
	}
}

func BiblPublicationStmt(ed *Edition) *teim.BiblPublicationStmt {
	ps := make([]teim.P, 0)

	if s := trimPtr(ed.Imprint); s != "" {
		ps = append(ps, teim.P{Text: "Imprint: " + s})
	}
	if s := trimPtr(ed.ImprintEN); s != "" {
		ps = append(ps, teim.P{Lang: "en", Text: "Imprint translation: " + s})
	}
	if s := trimPtr(ed.Colophon); s != "" {
		ps = append(ps, teim.P{Text: "Colophon: " + s})
	}
	if s := trimPtr(ed.ColophonEN); s != "" {
		ps = append(ps, teim.P{Lang: "en", Text: "Colophon translation: " + s})
	}
	if s := trimPtr(ed.Frontispiece); s != "" {
		ps = append(ps, teim.P{Text: "Frontispiece: " + s})
	}
	if s := trimPtr(ed.FrontispieceEN); s != "" {
		ps = append(ps, teim.P{Lang: "en", Text: "Frontispiece translation: " + s})
	}

	for _, sh := range ed.Shelfmarks {
		if sh.Scan != "" {
			ps = append(ps, teim.P{Text: "Digital scan: " + sh.Scan})
		}
		if sh.Shelfmark != "" {
			ps = append(ps, teim.P{Text: "Shelfmark: " + sh.Shelfmark})
		}
	}

	pubPlace := ""
	if len(ed.Cities) > 0 {
		pubPlace = strings.Join(nonEmptyStrings(ed.Cities), "; ")
	}

	var date *teim.Date
	if s := trimPtr(ed.Year); s != "" {
		date = &teim.Date{
			When: s,
			Text: s,
		}
	}

	publisher := strings.Join(nonEmptyStrings(ed.Publisher), "; ")

	if pubPlace == "" && date == nil && publisher == "" && len(ps) == 0 {
		return nil
	}

	return &teim.BiblPublicationStmt{
		PubPlace:  pubPlace,
		Date:      date,
		Publisher: publisher,
		Ps:        ps,
	}
}

func buildExtent(ed *Edition) *teim.Extent {
	var measures []teim.Measure

	if ed.Volumes != nil && *ed.Volumes > 0 {
		measures = append(measures, teim.Measure{
			Unit:     "volume",
			Quantity: *ed.Volumes,
			Text:     fmt.Sprintf("%d volume(s)", *ed.Volumes),
		})
	}

	if ed.Format != nil && *ed.Format > 0 {
		measures = append(measures, teim.Measure{
			Unit:     "format",
			Quantity: *ed.Format,
			Text:     strconv.Itoa(*ed.Format),
		})
	}

	if len(measures) == 0 {
		return nil
	}

	return &teim.Extent{
		Measures: measures,
	}
}

func buildNotesStmt(ed *Edition) *teim.NotesStmt {
	var notes []teim.Note

	addNote := func(noteType, text string) {
		if strings.TrimSpace(text) == "" {
			return
		}
		notes = append(notes, teim.Note{
			Type: noteType,
			Text: text,
		})
	}

	addNote("key", ed.Key)
	addNote("shortTitleSource", ed.ShortTitleSource)
	addNote("notes", ed.Notes)
	addNote("verified", strconv.FormatBool(ed.Verified))
	addNote("diagramCropsAvailable", strconv.FormatBool(ed.DiagramCropsAvailable))

	if ed.HasDiagrams != nil {
		addNote("hasDiagrams", strconv.FormatBool(*ed.HasDiagrams))
	}

	addNote("isManuscript", strconv.FormatBool(ed.IsManuscript))

	if ed.ManuscriptYearFrom != nil {
		addNote("manuscriptYearFrom", strconv.Itoa(*ed.ManuscriptYearFrom))
	}
	if ed.ManuscriptYearTo != nil {
		addNote("manuscriptYearTo", strconv.Itoa(*ed.ManuscriptYearTo))
	}

	addNote("manuscriptClass", ed.ManuscriptClass)
	if ed.ManuscriptSubclass != nil {
		addNote("manuscriptSubclass", *ed.ManuscriptSubclass)
	}

	if ed.USTCId != nil {
		addNote("ustcId", *ed.USTCId)
	}

	addNote("titlePageStatus", string(ed.TitlePageStatus))
	addNote("isElements", strconv.FormatBool(ed.IsElements))

	if len(ed.Languages) > 0 {
		addNote("languages", strings.Join(nonEmptyStrings(ed.Languages), ", "))
	}
	if len(ed.Corpus) > 0 {
		addNote("corpus", strings.Join(nonEmptyStrings(ed.Corpus), ", "))
	}
	if len(ed.Books) > 0 {
		addNote("books", intsToCSV(ed.Books))
	}
	if len(ed.Bibliography) > 0 {
		addNote("bibliography", strings.Join(nonEmptyStrings(ed.Bibliography), " | "))
	}
	if ed.ReprintOf != nil {
		addNote("reprintOf", *ed.ReprintOf)
	}
	if len(ed.AdditionalContent) > 0 {
		addNote("additionalContent", strings.Join(nonEmptyStrings(ed.AdditionalContent), " | "))
	}

	for i, sh := range ed.Shelfmarks {
		prefix := fmt.Sprintf("shelfmark[%d].", i)

		if sh.Volume != nil {
			addNote(prefix+"volume", strconv.Itoa(*sh.Volume))
		}
		addNote(prefix+"scan", sh.Scan)
		addNote(prefix+"shelfmark", sh.Shelfmark)
		addNote(prefix+"titlePageImg", sh.TitlePageImg)
		addNote(prefix+"frontispieceImg", sh.FrontispieceImg)
		addNote(prefix+"annotations", sh.Annotations)
		addNote(prefix+"copyright", sh.Copyright)
		addNote(prefix+"transcriptionAvailable", string(sh.TranscriptionAvailable))
		addNote(prefix+"note", sh.Note)
	}

	if len(notes) == 0 {
		return nil
	}

	return &teim.NotesStmt{
		Notes: notes,
	}
}

func isEmptyBiblFull(bf *teim.BiblFull) bool {
	return bf == nil ||
		(bf.TitleStmt == nil &&
			bf.PublicationStmt == nil &&
			bf.Extent == nil &&
			bf.NotesStmt == nil)
}

func trimPtr(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func nonEmptyStrings(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func intsToCSV(nums []int) string {
	if len(nums) == 0 {
		return ""
	}

	parts := make([]string, 0, len(nums))
	for _, n := range nums {
		parts = append(parts, strconv.Itoa(n))
	}
	return strings.Join(parts, ", ")
}

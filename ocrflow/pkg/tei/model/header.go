package model

type Header struct {
	FileDesc FileDesc  `xml:"fileDesc"`
	StandOff *StandOff `xml:"standOff,omitempty"`
}

type FileDesc struct {
	TitleStmt       TitleStmt       `xml:"titleStmt"`
	PublicationStmt PublicationStmt `xml:"publicationStmt"`
	SourceDesc      SourceDesc      `xml:"sourceDesc"`
}

type TitleStmt struct {
	Title string `xml:"title"`
}

type PublicationStmt struct {
	Publisher string `xml:"publisher,omitempty"`
	P         string `xml:"p,omitempty"`
}

type SourceDesc struct {
	BiblFull *BiblFull `xml:"biblFull,omitempty"`
	P        string    `xml:"p,omitempty"` // keep fallback if needed
}

type BiblFull struct {
	TitleStmt       *BiblTitleStmt       `xml:"titleStmt,omitempty"`
	PublicationStmt *BiblPublicationStmt `xml:"publicationStmt,omitempty"`
	Extent          *Extent              `xml:"extent,omitempty"`
	NotesStmt       *NotesStmt           `xml:"notesStmt,omitempty"`
}

type BiblTitleStmt struct {
	Titles []Title  `xml:"title"`
	Editor []string `xml:"editor,omitempty"`
}

type Title struct {
	Type    string `xml:"type,attr,omitempty"`
	Lang    string `xml:"xml:lang,attr,omitempty"`
	Content string `xml:",chardata"`
}

type BiblPublicationStmt struct {
	PubPlace  string `xml:"pubPlace,omitempty"`
	Date      *Date  `xml:"date,omitempty"`
	Publisher string `xml:"publisher,omitempty"`
	Ps        []P    `xml:"p,omitempty"` // imprint, colophon, translations
}

type Date struct {
	When string `xml:"when,attr,omitempty"`
	Text string `xml:",chardata"`
}

type P struct {
	Lang string `xml:"xml:lang,attr,omitempty"`
	Text string `xml:",chardata"`
}

type Extent struct {
	Measures []Measure `xml:"measure"`
}

type Measure struct {
	Unit     string `xml:"unit,attr,omitempty"`
	Quantity int    `xml:"quantity,attr,omitempty"`
	Text     string `xml:",chardata"`
}

type NotesStmt struct {
	Notes []Note `xml:"note"`
}

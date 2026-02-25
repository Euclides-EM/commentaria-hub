package model

type Header struct {
	FileDesc     FileDesc      `xml:"fileDesc"`
	EncodingDesc *EncodingDesc `xml:"encodingDesc,omitempty"`
	ProfileDesc  *ProfileDesc  `xml:"profileDesc,omitempty"`
	StandOff     *StandOff     `xml:"standOff,omitempty"`
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
	P string `xml:"p"`
}

package model

// EncodingDesc holds encoding description (e.g. taxonomy for features and facts).
type EncodingDesc struct {
	ClassDecl *ClassDecl `xml:"classDecl,omitempty"`
}

// ClassDecl wraps taxonomy (TEI classDecl).
type ClassDecl struct {
	Taxonomy *Taxonomy `xml:"taxonomy,omitempty"`
}

// Taxonomy defines a vocabulary of categories (features and fact types).
type Taxonomy struct {
	XmlID      string     `xml:"xml:id,attr,omitempty"`
	Categories []Category `xml:"category,omitempty"`
}

// Category is a single category in a taxonomy (TEI category with catDesc).
type Category struct {
	XmlID   string `xml:"xml:id,attr,omitempty"`
	CatDesc string `xml:"catDesc,omitempty"`
}

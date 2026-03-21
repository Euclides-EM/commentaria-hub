package tei

import "github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"

func buildFileDesc(biblMetadata *model.BiblFull) model.FileDesc {
	return model.FileDesc{
		TitleStmt: model.TitleStmt{Title: "Converted from lines"},
		PublicationStmt: model.PublicationStmt{
			P: "Unpublished research data",
		},
		SourceDesc: model.SourceDesc{
			P:        "Derived from extracted text lines",
			BiblFull: biblMetadata,
		},
	}
}

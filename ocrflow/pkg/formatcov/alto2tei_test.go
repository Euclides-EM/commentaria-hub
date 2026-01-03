package formatcov

import (
	_ "embed"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/alto/page-0502.xml
var testALTOPage0502 []byte

//go:embed testdata/tei/page-0502.xml
var expectedTEIPage0502 []byte

func TestConvertALTOToTEI(t *testing.T) {
	var a alto.Alto
	if err := xml.Unmarshal(testALTOPage0502, &a); err != nil {
		assert.NoError(t, err)
		return
	}

	tei, err := ConvertALTOToTEI(&a, ALTOToTEIOptions{
		RowTolPx:     6,
		ParaGapPx:    28,
		KeepEmpty:    false,
		Title:        "Converted from ALTO",
		FacsFromPage: true,
	})
	if err != nil {
		assert.NoError(t, err)
		return
	}

	assert.Equal(t, strings.TrimSpace(string(expectedTEIPage0502)), strings.TrimSpace(string(tei)))
}

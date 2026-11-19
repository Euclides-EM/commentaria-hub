package annotationrule

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelDetectApplicableStagesByModelType(t *testing.T) {
	tests := []struct {
		name string
		rule *ModelDetect
		want PipelineStage
	}{
		{
			name: "segmentation model starts from raw",
			rule: NewModelDetect("segment-model"),
			want: PipelineStageRaw,
		},
		{
			name: "OCR model requires text lines",
			rule: NewOCRModelDetect("ocr-model"),
			want: PipelineStageTextLineSegmentation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.rule)
			require.NoError(t, err)

			var payload struct {
				ApplicableStages []PipelineStage `json:"applicable_stages"`
			}
			require.NoError(t, json.Unmarshal(data, &payload))
			require.Equal(t, []PipelineStage{tt.want}, payload.ApplicableStages)
		})
	}
}

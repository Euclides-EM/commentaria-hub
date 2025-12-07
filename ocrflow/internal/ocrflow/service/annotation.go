package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/annotationrule"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/tiendc/go-deepcopy"
	"math/rand"
	"os"
	"path"
)

// todo: add interfaces to all services

type Annotation struct {
	m           map[string]*model.Annotation
	datasetSvc  *Dataset
	ruleApplier *AnnotationRuleApplier
}

func NewAnnotationsService(
	datasetSvc *Dataset,
	ruleApplier *AnnotationRuleApplier,
) *Annotation {
	annotations := map[string]*model.Annotation{}
	manuallyAnnotated := map[string]*model.Annotation{
		// Manually annotated

		// Only one big MainZone, without any subtype like paragraph, enunciation, etc.
		// In Robolflow it's here: https://app.roboflow.com/mia-workplace/paris-1615-withmznosubtypes-tkgii/1
		"gog80x": {
			Meta:        model.NewMeta("gog80x"),
			Description: "Manually annotated (ground truth); Only one big MainZone, without any subtype like paragraph, enunciation, etc.",
			Pages:       "66,160,197,303,497,20,49,91,95-97,148-149,153,183,195-196,255,257,295-297,315,388,395-397,450-465,495-496,508,596,603,624,256,339,387,595,597,609",
			DatasetID:   "rrpbnk",
			YoloDir:     "store/data/annotations/gog80x/yolo",
		},

		// Polygons for subtypes like paragraph, enunciation and main zones.
		// In Roboflow it's here: https://app.roboflow.com/mia-workplace/paris-1615-polygonswithmz-wsrge/1
		"f0k3ks": {
			Meta:        model.NewMeta("f0k3ks"),
			Description: "Manually annotated (ground truth); Polygons for subtypes like paragraph, enunciation and main zones.",
			Pages:       "66,160,197,303,497,20,49,91,95-97,148-149,153,183,195-196,255,257,295-297,315,388,395-397,450-465,495-496,508,596,603,624,256,339,387,595,597,609",
			DatasetID:   "rrpbnk",
			YoloDir:     "store/data/annotations/f0k3ks/yolo",
		},

		// Includes a big MainZone, in addition to subtypes like paragraph, enunciation, etc.
		// In some cases, instead of boxes polygons were drawn to better fit the text areas.
		// In Roboflow it's here: https://app.roboflow.com/mia-workplace/paris-1615-polygonswithmz-wsrge/1
		"idim36": {
			Meta:        model.NewMeta("idim36"),
			Description: "Manually annotated (ground truth); Includes a big MainZone, in addition to subtypes like paragraph, enunciation, etc. In some cases, instead of boxes polygons were drawn to better fit the text areas.",
			Pages:       "66,160,197,303,497,20,49,91,95-97,148-149,153,183,195-196,255,257,295-297,315,388,395-397,450-465,495-496,508,596,603,624,256,339,387,595,597,609",
			DatasetID:   "rrpbnk",
			YoloDir:     "store/data/annotations/idim36/yolo",
		},

		// Inferred Annotations

		"4s48pk": {
			Meta:        model.NewMeta("4s48pk"),
			Description: "Inferred annotations from Kraken on the same pages as above",
			Pages:       "15-655",
			DatasetID:   "uk5wbj",
			AltoDir:     "store/data/uk5wbj/annotations/4s48pk/alto",
			AppliedRules: []annotationrule.AnnotationRule{
				annotationrule.NewSegment("1615FineTunedCapricciosaM_0312"),
			},
		},
	}

	for k, v := range manuallyAnnotated {
		annotations[k] = v
		annotations[k].YoloDir = store.DatasetAnnotationYoloDir(annotations[k], datasetSvc.dataDir)
	}

	return &Annotation{
		m:           annotations,
		datasetSvc:  datasetSvc,
		ruleApplier: ruleApplier,
	}

}

func (a *Annotation) ListAnnotations(id string) ([]*model.Annotation, error) {
	annotations := make([]*model.Annotation, 0)
	for _, annotation := range a.m {
		if annotation.DatasetID == id {
			var dst *model.Annotation
			if err := deepcopy.Copy(&dst, &annotation); err != nil {
				return nil, fmt.Errorf("failed to copy annotation: %w", err)
			}
			annotations = append(annotations, dst)
		}
	}
	return annotations, nil
}

func (a *Annotation) Get(datasetId, id string) (*model.Annotation, error) {
	annotation, ok := a.m[id]
	if !ok || annotation.DatasetID != datasetId {
		return nil, fmt.Errorf("annotation not found")
	}
	var dst *model.Annotation
	if err := deepcopy.Copy(&dst, &annotation); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	return dst, nil
}

func (a *Annotation) Create(datasetID string, ann *model.Annotation, randomPages int) (*model.Annotation, error) {
	// validate dataset exists
	ds, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	if ds.ImagesPath == "" {
		return nil, fmt.Errorf("no JPGs path found for dataset")
	}

	// assign basic fields
	ann.ID = idgen.GenerateID()
	ann.DatasetID = datasetID

	// select random pages if requested
	if ann.Pages == "" && randomPages > 0 {
		allPages, err := store.InferPages(ds.ImagesPath, "")
		if err != nil {
			return nil, fmt.Errorf("failed to count dataset pages: %w", err)
		}
		if len(allPages) < randomPages {
			return nil, fmt.Errorf("requested random pages %d exceeds total pages %d", randomPages, len(allPages))
		}
		rand.Shuffle(len(allPages), func(i, j int) {
			allPages[i], allPages[j] = allPages[j], allPages[i]
		})
		ann.Pages = pagesparser.ToString(allPages[:randomPages])
	}

	// verify page images exist for all specified pages
	pages, err := pagesparser.Parse(ann.Pages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pages: %w", err)
	}

	for _, p := range pages {
		filename := pagesparser.PageToPNGFilename(p)
		if _, err := os.Stat(path.Join(ds.ImagesPath, filename)); err != nil {
			return nil, fmt.Errorf("no such file %s in existing dataset", filename)
		}
	}

	var dst *model.Annotation
	if err := deepcopy.Copy(&dst, &ann); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	a.m[ann.ID] = dst
	return a.m[ann.ID], nil
}

func (a *Annotation) CreateFromZip(datasetID string, format model.AnnotationFormat, save func(dstPath string) error) (*model.Annotation, error) {
	_, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	ann := &model.Annotation{
		Meta:      model.NewMeta(idgen.GenerateID()),
		DatasetID: datasetID,
	}
	ann.AltoDir = store.DatasetAnnotationAltoDir(ann, a.datasetSvc.dataDir)
	dstPath := ann.AltoDir
	if format == model.AnnotationFormatYolo {
		dstPath = store.DatasetAnnotationYoloDir(ann, a.datasetSvc.dataDir)
		ann.YoloDir = dstPath
	}
	if err := save(dstPath); err != nil {
		return nil, fmt.Errorf("failed to store uploaded annotations: %w", err)
	}
	if format == model.AnnotationFormatYolo {
		if err := formatcov.Yolo2Alto(ann.YoloDir, ann.AltoDir); err != nil {
			return nil, fmt.Errorf("failed to convert YOLO annotations to ALTO: %w", err)
		}
	}

	pages, err := store.InferPages(dstPath, format)
	if err != nil {
		return nil, fmt.Errorf("failed to infer pages from uploaded annotations: %w", err)
	}
	ann.Pages = pagesparser.ToString(pages)

	var dst *model.Annotation
	if err := deepcopy.Copy(&dst, &ann); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	a.m[ann.ID] = dst
	return a.m[ann.ID], nil
}

func (a *Annotation) ApplyRules(datasetID string, id string, aar *annotationrule.ApplyRules) (*model.Annotation, error) {
	fromDB, ok := a.m[id]
	if !ok || fromDB.DatasetID != datasetID {
		return nil, fmt.Errorf("annotation not found")
	}
	ds, err := a.datasetSvc.Get(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}

	var ann *model.Annotation
	if err := deepcopy.Copy(&ann, &fromDB); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}

	if aar.Action == annotationrule.ApplyRulesActionCreateNew {
		ann.ID = idgen.GenerateID()
		if fromDB.AltoDir != "" {
			ann.AltoDir = store.DatasetAnnotationAltoDir(ann, a.datasetSvc.dataDir)
			if err := futils.CopyDir(store.DatasetAnnotationAltoDir(fromDB, a.datasetSvc.dataDir), ann.AltoDir); err != nil {
				return nil, fmt.Errorf("failed to copy annotations for new annotation: %w", err)
			}
		}
		ann.OriginAnnotationID = id
		var dst2 *model.Annotation
		if err := deepcopy.Copy(&dst2, &ann); err != nil {
			return nil, fmt.Errorf("failed to copy annotation: %w", err)
		}
		a.m[ann.ID] = dst2
	}

	// apply rules...
	if err := a.ruleApplier.ApplyRules(ds.ImagesPath, ann, aar.Rules); err != nil {
		return nil, fmt.Errorf("failed to apply annotation rules: %w", err)
	}

	ann.AppliedRules = append(ann.AppliedRules, aar.Rules...)

	var dst2 *model.Annotation
	if err := deepcopy.Copy(&dst2, &ann); err != nil {
		return nil, fmt.Errorf("failed to copy annotation: %w", err)
	}
	a.m[ann.ID] = dst2

	return a.m[ann.ID], nil
}

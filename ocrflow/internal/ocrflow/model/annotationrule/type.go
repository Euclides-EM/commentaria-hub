package annotationrule

type Type string

const (
	TypeSegment                      Type = "segment"
	TypeSlicePages                   Type = "slice_pages"
	TypeStretch                      Type = "stretch"
	TypeAddMargin                    Type = "add_margin"
	TypeLinesDetect                  Type = "lines_detect"
	TypeRemoveCategories             Type = "remove_categories"
	TypeRemoveOverlap                Type = "remove_overlap"
	TypeReassignTextLinesByTolerance Type = "reassign_text_lines_by_tolerance"
	TypeTextBlocksCorrections        Type = "text_blocks_corrections"
)

var ruleFactories = map[Type]func() AnnotationRule{
	TypeSlicePages:                   func() AnnotationRule { return NewSlicePagesFixed("") },
	TypeStretch:                      func() AnnotationRule { return NewZeroStretch() },
	TypeAddMargin:                    func() AnnotationRule { return NewZeroAddMargin() },
	TypeLinesDetect:                  func() AnnotationRule { return NewLinesDetect(nil, nil) },
	TypeSegment:                      func() AnnotationRule { return NewSegment("") },
	TypeRemoveCategories:             func() AnnotationRule { return NewRemoveCategories(nil) },
	TypeRemoveOverlap:                func() AnnotationRule { return NewRemoveOverlap(nil, 0.0) },
	TypeReassignTextLinesByTolerance: func() AnnotationRule { return NewReassignTextLinesByTolerance("", "", 0.0, 0.0) },
	TypeTextBlocksCorrections:        func() AnnotationRule { return NewTextBlockCorrections(nil) },
}

var applicableStagesByType = map[Type][]PipelineStage{
	TypeSlicePages:                   {PipelineStageRaw, PipelineStageZoneSegmentation, PipelineStageTextLineSegmentation, PipelineStageOCR},
	TypeStretch:                      {PipelineStageZoneSegmentation},
	TypeAddMargin:                    {PipelineStageZoneSegmentation},
	TypeLinesDetect:                  {PipelineStageZoneSegmentation},
	TypeSegment:                      {PipelineStageRaw},
	TypeRemoveCategories:             {PipelineStageZoneSegmentation},
	TypeRemoveOverlap:                {PipelineStageZoneSegmentation},
	TypeReassignTextLinesByTolerance: {PipelineStageTextLineSegmentation},
	TypeTextBlocksCorrections:        {PipelineStageOCR},
}

var minEnsuredStageByType = map[Type]PipelineStage{
	TypeSlicePages:                   PipelineStageRaw,
	TypeStretch:                      PipelineStageZoneSegmentation,
	TypeAddMargin:                    PipelineStageZoneSegmentation,
	TypeLinesDetect:                  PipelineStageTextLineSegmentation,
	TypeSegment:                      PipelineStageZoneSegmentation,
	TypeRemoveCategories:             PipelineStageZoneSegmentation,
	TypeRemoveOverlap:                PipelineStageZoneSegmentation,
	TypeReassignTextLinesByTolerance: PipelineStageTextLineSegmentation,
	TypeTextBlocksCorrections:        PipelineStageOCR,
}

func ZeroFromType(t Type) AnnotationRule {
	if factory, ok := ruleFactories[t]; ok {
		return factory()
	}
	return nil
}

var AllAnnotationRuleTypes = []Type{
	TypeSegment,
	TypeSlicePages,
	TypeStretch,
	TypeAddMargin,
	TypeLinesDetect,
	TypeRemoveCategories,
	TypeRemoveOverlap,
	TypeReassignTextLinesByTolerance,
	TypeTextBlocksCorrections,
}

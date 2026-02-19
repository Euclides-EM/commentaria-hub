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
	TypeResolveOverlapWithPriority   Type = "resolve_overlap_with_priority"
	TypeRecategorizeByAlignment      Type = "recategorize_by_alignment"
	TypeReassignTextLinesByTolerance Type = "reassign_text_lines_by_tolerance"
	TypeTextBlocksCorrections        Type = "text_blocks_corrections"
	TypeDetectText                   Type = "detect_text"
)

var ruleFactories = map[Type]func() AnnotationRule{
	TypeSlicePages:                   func() AnnotationRule { return NewSlicePagesFixed("") },
	TypeStretch:                      func() AnnotationRule { return NewZeroStretch() },
	TypeAddMargin:                    func() AnnotationRule { return NewZeroAddMargin() },
	TypeLinesDetect:                  func() AnnotationRule { return NewLinesDetect(nil, nil) },
	TypeSegment:                      func() AnnotationRule { return NewSegment("") },
	TypeRemoveCategories:             func() AnnotationRule { return NewRemoveCategories(nil) },
	TypeRemoveOverlap:                func() AnnotationRule { return NewRemoveOverlap(nil, 0.0) },
	TypeResolveOverlapWithPriority:   func() AnnotationRule { return NewResolveOverlapWithPriority("", "", 0.0) },
	TypeRecategorizeByAlignment:      func() AnnotationRule { return NewRecategorizeByAlignment("", "", "", AlignmentHorizontal, 0.0) },
	TypeReassignTextLinesByTolerance: func() AnnotationRule { return NewReassignTextLinesByTolerance("", "", 0.0, 0.0) },
	TypeTextBlocksCorrections:        func() AnnotationRule { return NewTextBlockCorrections(nil) },
	TypeDetectText:                   func() AnnotationRule { return NewDetectText("") },
}

var applicableStagesByType = map[Type][]PipelineStage{
	TypeSlicePages:                   {PipelineStageRaw, PipelineStageZoneSegmentation, PipelineStageTextLineSegmentation, PipelineStageOCR},
	TypeStretch:                      {PipelineStageZoneSegmentation},
	TypeAddMargin:                    {PipelineStageZoneSegmentation},
	TypeLinesDetect:                  {PipelineStageZoneSegmentation},
	TypeSegment:                      {PipelineStageRaw},
	TypeRemoveCategories:             {PipelineStageZoneSegmentation},
	TypeRemoveOverlap:                {PipelineStageZoneSegmentation},
	TypeResolveOverlapWithPriority:   {PipelineStageZoneSegmentation},
	TypeRecategorizeByAlignment:      {PipelineStageZoneSegmentation},
	TypeReassignTextLinesByTolerance: {PipelineStageTextLineSegmentation},
	TypeTextBlocksCorrections:        {PipelineStageOCR},
	TypeDetectText:                   {PipelineStageTextLineSegmentation},
}

var minEnsuredStageByType = map[Type]PipelineStage{
	TypeSlicePages:                   PipelineStageRaw,
	TypeStretch:                      PipelineStageZoneSegmentation,
	TypeAddMargin:                    PipelineStageZoneSegmentation,
	TypeLinesDetect:                  PipelineStageTextLineSegmentation,
	TypeSegment:                      PipelineStageZoneSegmentation,
	TypeRemoveCategories:             PipelineStageZoneSegmentation,
	TypeRemoveOverlap:                PipelineStageZoneSegmentation,
	TypeResolveOverlapWithPriority:   PipelineStageZoneSegmentation,
	TypeRecategorizeByAlignment:      PipelineStageZoneSegmentation,
	TypeReassignTextLinesByTolerance: PipelineStageTextLineSegmentation,
	TypeTextBlocksCorrections:        PipelineStageOCR,
	TypeDetectText:                   PipelineStageOCR,
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
	TypeResolveOverlapWithPriority,
	TypeRecategorizeByAlignment,
	TypeReassignTextLinesByTolerance,
	TypeTextBlocksCorrections,
	TypeDetectText,
}

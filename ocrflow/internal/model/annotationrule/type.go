package annotationrule

type Type string

const (
	TypeModelDetect                  Type = "model_detect"
	TypeSlicePages                   Type = "slice_pages"
	TypeStretch                      Type = "stretch"
	TypeAddMargin                    Type = "add_margin"
	TypeLinesDetect                  Type = "lines_detect"
	TypeRemoveCategories             Type = "remove_categories"
	TypeRenameCategories             Type = "rename_categories"
	TypeRemoveOverlap                Type = "remove_overlap"
	TypeResolveOverlapWithPriority   Type = "resolve_overlap_with_priority"
	TypeRecategorizeByAlignment      Type = "recategorize_by_alignment"
	TypeLimitCategoryZones           Type = "limit_category_zones"
	TypeReassignTextLinesByTolerance Type = "reassign_text_lines_by_tolerance"
	TypeTextBlocksCorrections        Type = "text_blocks_corrections"
	TypeLLMTranscriptionCorrector    Type = "llm_transcription_corrector"
)

var ruleFactories = map[Type]func() AnnotationRule{
	TypeSlicePages:                   func() AnnotationRule { return NewSlicePagesFixed("") },
	TypeStretch:                      func() AnnotationRule { return NewZeroStretch() },
	TypeAddMargin:                    func() AnnotationRule { return NewZeroAddMargin() },
	TypeLinesDetect:                  func() AnnotationRule { return NewLinesDetect(nil, nil) },
	TypeModelDetect:                  func() AnnotationRule { return NewModelDetect("") },
	TypeRemoveCategories:             func() AnnotationRule { return NewRemoveCategories(nil) },
	TypeRenameCategories:             func() AnnotationRule { return NewRenameCategories(nil) },
	TypeRemoveOverlap:                func() AnnotationRule { return NewRemoveOverlap(nil, 0.0) },
	TypeResolveOverlapWithPriority:   func() AnnotationRule { return NewResolveOverlapWithPriority("", "", 0.0) },
	TypeRecategorizeByAlignment:      func() AnnotationRule { return NewRecategorizeByAlignment("", "", "", AlignmentHorizontal, 0.0) },
	TypeLimitCategoryZones:           func() AnnotationRule { return NewLimitCategoryZones("", 0, KeepPositionTop) },
	TypeReassignTextLinesByTolerance: func() AnnotationRule { return NewReassignTextLinesByTolerance("", "", 0.0, 0.0) },
	TypeTextBlocksCorrections:        func() AnnotationRule { return NewTextBlockCorrections(nil) },
	TypeLLMTranscriptionCorrector:    func() AnnotationRule { return NewLLMTranscriptionCorrector("", "", nil, false) },
}

var applicableStagesByType = map[Type][]PipelineStage{
	TypeSlicePages:                   {PipelineStageRaw, PipelineStageZoneSegmentation, PipelineStageTextLineSegmentation, PipelineStageOCR},
	TypeStretch:                      {PipelineStageZoneSegmentation},
	TypeAddMargin:                    {PipelineStageZoneSegmentation},
	TypeLinesDetect:                  {PipelineStageZoneSegmentation},
	TypeModelDetect:                  {PipelineStageRaw},
	TypeRemoveCategories:             {PipelineStageZoneSegmentation},
	TypeRenameCategories:             {PipelineStageZoneSegmentation},
	TypeRemoveOverlap:                {PipelineStageZoneSegmentation},
	TypeResolveOverlapWithPriority:   {PipelineStageZoneSegmentation},
	TypeRecategorizeByAlignment:      {PipelineStageZoneSegmentation},
	TypeLimitCategoryZones:           {PipelineStageZoneSegmentation},
	TypeReassignTextLinesByTolerance: {PipelineStageTextLineSegmentation},
	TypeTextBlocksCorrections:        {PipelineStageOCR},
	TypeLLMTranscriptionCorrector:    {PipelineStageOCR},
}

var minEnsuredStageByType = map[Type]PipelineStage{
	TypeSlicePages:                   PipelineStageRaw,
	TypeStretch:                      PipelineStageZoneSegmentation,
	TypeAddMargin:                    PipelineStageZoneSegmentation,
	TypeLinesDetect:                  PipelineStageTextLineSegmentation,
	TypeModelDetect:                  PipelineStageZoneSegmentation,
	TypeRemoveCategories:             PipelineStageZoneSegmentation,
	TypeRenameCategories:             PipelineStageZoneSegmentation,
	TypeRemoveOverlap:                PipelineStageZoneSegmentation,
	TypeResolveOverlapWithPriority:   PipelineStageZoneSegmentation,
	TypeRecategorizeByAlignment:      PipelineStageZoneSegmentation,
	TypeLimitCategoryZones:           PipelineStageZoneSegmentation,
	TypeReassignTextLinesByTolerance: PipelineStageTextLineSegmentation,
	TypeTextBlocksCorrections:        PipelineStageOCR,
	TypeLLMTranscriptionCorrector:    PipelineStageOCR,
}

func ZeroFromType(t Type) AnnotationRule {
	if factory, ok := ruleFactories[t]; ok {
		return factory()
	}
	return nil
}

var AllAnnotationRuleTypes = []Type{
	TypeModelDetect,
	TypeSlicePages,
	TypeStretch,
	TypeAddMargin,
	TypeLinesDetect,
	TypeRemoveCategories,
	TypeRenameCategories,
	TypeRemoveOverlap,
	TypeResolveOverlapWithPriority,
	TypeRecategorizeByAlignment,
	TypeLimitCategoryZones,
	TypeReassignTextLinesByTolerance,
	TypeTextBlocksCorrections,
	TypeLLMTranscriptionCorrector,
}

var PreferAsyncTypes = []Type{
	TypeModelDetect,
	TypeLinesDetect,
	TypeLLMTranscriptionCorrector,
}

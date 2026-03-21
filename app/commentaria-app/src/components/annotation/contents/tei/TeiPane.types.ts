import type {
  feature_Feature,
  feature_Result,
  feature_ResultValue,
} from '@hub-api'

export type ResolvedTeiFeature = {
  id: string
  label: string
  description: string
  color: string
  isDefault: boolean
  renderMode: 'fill' | 'outline'
}

export type TeiTooltipItem = {
  id: string
  featureId: string
  categoryId: string
  label: string
  description: string
  surface: string
  normalized: string
  institution: string
  ancientPersona: string
  paragraphIndex: number
  start: number
  end: number
  fromAnchorId: string
  toAnchorId: string
  color: string
}

export type TeiTooltipState = {
  x: number
  y: number
  items: TeiTooltipItem[]
}

export type SelectionDraft = {
  paragraphIndex: number
  start: number
  end: number
  surface: string
  x: number
  y: number
}

export type DraftHighlight = {
  localId: string
  sourceId: string
  paragraphIndex: number
  start: number
  end: number
  featureId: string
  categoryId: string
  surface: string
  normalized: string
  institution: string
  ancientPersona: string
  fromAnchorId: string
  toAnchorId: string
}

export type FeatureModalState = {
  selection: SelectionDraft
}

export type SourceOption = {
  value: 'annotation' | 'edition'
  label: string
}

export type FeatureOption = {
  value: string
  label: string
  color?: string
  description: string
  isAction?: boolean
}

export type TeiFeatureResult = feature_Result
export type TeiFeatureResultValue = feature_ResultValue
export type TeiDatasetFeature = feature_Feature

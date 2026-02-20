import { useQuery } from '@tanstack/react-query'
import { MetadataService } from '@hub-api'

export const useAnnotationRules = () =>
  useQuery({
    queryKey: ['annotationRules'],
    queryFn: () => MetadataService.getAnnotationRules(),
  })

export const usePipelineStages = () =>
  useQuery({
    queryKey: ['pipelineStages'],
    queryFn: () => MetadataService.getPipelineStages(),
  })

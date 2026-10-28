import { useMutation, useQueryClient } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type { ResearchPipeline } from '#/types/pipeline.type'
import { fetchHelper } from '#/utils/fetchHelper'

export const useResumePipeline = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (pipelineId: string) => {
      const response = await fetchHelper<ApiResponse<ResearchPipeline>>(
        `/api/v1/pipelines/${encodeURIComponent(pipelineId)}/resume`,
        { method: 'POST' },
      )
      return response.data
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['pipeline', data.id], data)
      void queryClient.invalidateQueries({ queryKey: ['pipelines'] })
    },
  })
}

export const useCancelPipeline = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (pipelineId: string) => {
      const response = await fetchHelper<ApiResponse<null>>(
        `/api/v1/pipelines/${encodeURIComponent(pipelineId)}/cancel`,
        { method: 'POST' },
      )
      return response.data
    },
    onSuccess: (_, pipelineId) => {
      void queryClient.invalidateQueries({ queryKey: ['pipeline', pipelineId] })
      void queryClient.invalidateQueries({ queryKey: ['pipelines'] })
    },
  })
}

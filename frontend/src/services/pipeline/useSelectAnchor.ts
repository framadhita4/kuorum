import { useMutation, useQueryClient } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type { ResearchPipeline } from '#/types/pipeline.type'
import { fetchHelper } from '#/utils/fetchHelper'

export interface SelectAnchorPayload {
  pipelineId: string
  anchorPaperId: string
}

export const useSelectAnchor = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ pipelineId, anchorPaperId }: SelectAnchorPayload) => {
      const response = await fetchHelper<
        ApiResponse<ResearchPipeline>,
        { anchorPaperId: string }
      >(`/api/v1/pipelines/${encodeURIComponent(pipelineId)}/select-anchor`, {
        method: 'POST',
        body: { anchorPaperId },
      })
      return response.data
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['pipeline', data.id], data)
      void queryClient.invalidateQueries({ queryKey: ['pipelines'] })
    },
  })
}

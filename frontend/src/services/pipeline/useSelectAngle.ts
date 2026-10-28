import { useMutation, useQueryClient } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type { ResearchPipeline } from '#/types/pipeline.type'
import { fetchHelper } from '#/utils/fetchHelper'

export interface SelectAnglePayload {
  pipelineId: string
  angleId: string
}

export const useSelectAngle = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ pipelineId, angleId }: SelectAnglePayload) => {
      const response = await fetchHelper<
        ApiResponse<ResearchPipeline>,
        { angleId: string }
      >(`/api/v1/pipelines/${encodeURIComponent(pipelineId)}/select-angle`, {
        method: 'POST',
        body: { angleId },
      })
      return response.data
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['pipeline', data.id], data)
      void queryClient.invalidateQueries({ queryKey: ['pipelines'] })
    },
  })
}

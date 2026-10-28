import { useMutation, useQueryClient } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import { fetchHelper } from '#/utils/fetchHelper'

interface CancelPipelinePayload {
  pipelineId: string
}

export const useCancelPipeline = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ pipelineId }: CancelPipelinePayload) => {
      const response = await fetchHelper<ApiResponse<null>, null>(
        `/api/v1/pipelines/${encodeURIComponent(pipelineId)}/cancel`,
        {
          method: 'POST',
        },
      )
      return response.data
    },
    onSuccess: (_, { pipelineId }) => {
      void queryClient.invalidateQueries({
        queryKey: ['pipeline', pipelineId],
      })
      void queryClient.invalidateQueries({
        queryKey: ['pipelines'],
      })
    },
  })
}

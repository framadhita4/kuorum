import { useMutation, useQueryClient } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type {
  ExecutionMode,
  ResearchApproach,
  ResearchPipeline,
} from '#/types/pipeline.type'
import { fetchHelper } from '#/utils/fetchHelper'

export interface StartPipelinePayload {
  prompt: string
  mode?: ExecutionMode
  approach?: ResearchApproach
}

export const useStartPipeline = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (payload: StartPipelinePayload) => {
      const response = await fetchHelper<
        ApiResponse<ResearchPipeline>,
        StartPipelinePayload
      >('/api/v1/pipelines', {
        method: 'POST',
        body: payload,
      })
      return response.data
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['pipeline', data.id], data)
      void queryClient.invalidateQueries({ queryKey: ['pipelines'] })
    },
  })
}

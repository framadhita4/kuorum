import { useQuery } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type { ResearchPipeline } from '#/types/pipeline.type'
import { fetchHelper } from '#/utils/fetchHelper'

export const useGetPipeline = (pipelineId: string | undefined | null) => {
  return useQuery({
    queryKey: ['pipeline', pipelineId],
    queryFn: async () => {
      if (!pipelineId) throw new Error('Pipeline ID is required')
      const response = await fetchHelper<ApiResponse<ResearchPipeline>>(
        `/api/v1/pipelines/${encodeURIComponent(pipelineId)}`,
      )
      return response.data
    },
    enabled: Boolean(pipelineId),
    staleTime: 5000,
  })
}

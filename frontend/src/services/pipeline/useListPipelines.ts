import { useQuery } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type { ResearchPipeline } from '#/types/pipeline.type'
import { fetchHelper } from '#/utils/fetchHelper'

export const useListPipelines = (limit = 20) => {
  return useQuery({
    queryKey: ['pipelines', limit],
    queryFn: async () => {
      const response = await fetchHelper<ApiResponse<Array<ResearchPipeline>>>(
        `/api/v1/pipelines?limit=${limit}`,
      )
      return response.data
    },
    staleTime: 10000,
  })
}

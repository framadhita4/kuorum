import { useQuery } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type { GetModelsResult } from '#/types/llm.type'
import { fetchHelper } from '#/utils/fetchHelper'

export const GET_MODELS_QUERY_KEY = ['llm', 'models'] as const

export const useGetModels = () => {
  return useQuery({
    queryKey: GET_MODELS_QUERY_KEY,
    queryFn: async () => {
      const response = await fetchHelper<ApiResponse<GetModelsResult>>(
        '/api/v1/models',
        { method: 'GET' },
      )
      return response.data
    },
    staleTime: 10 * 60 * 1000, // 10 minutes
  })
}

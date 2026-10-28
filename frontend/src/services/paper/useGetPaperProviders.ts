import { useQuery } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type { PaperProvidersResult } from '#/types/paper.type'
import { fetchHelper } from '#/utils/fetchHelper'

export function useGetPaperProviders() {
  return useQuery({
    queryKey: ['papers', 'providers'],
    queryFn: async () => {
      const response = await fetchHelper<ApiResponse<PaperProvidersResult>>(
        '/api/v1/papers/providers',
        { method: 'GET' },
      )
      return response.data
    },
    staleTime: 1000 * 60 * 60, // 1 hour cache
  })
}

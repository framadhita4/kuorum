import { useQuery } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type { SearchPapersParams, SearchPapersResult } from '#/types/paper.type'
import { fetchHelper } from '#/utils/fetchHelper'

export function useSearchPapers(params: SearchPapersParams, enabled = true) {
  const queryStr = new URLSearchParams()
  if (params.query) queryStr.set('query', params.query)
  if (params.limit) queryStr.set('limit', String(params.limit))
  if (params.providers && params.providers.length > 0) {
    queryStr.set('providers', params.providers.join(','))
  }

  const endpoint = `/api/v1/papers/search?${queryStr.toString()}`

  return useQuery({
    queryKey: ['papers', 'search', params],
    queryFn: async () => {
      const response = await fetchHelper<ApiResponse<SearchPapersResult>>(
        endpoint,
        { method: 'GET' },
      )
      return response.data
    },
    enabled: enabled && Boolean(params.query.trim()),
    staleTime: 1000 * 60 * 5, // 5 minutes cache
  })
}

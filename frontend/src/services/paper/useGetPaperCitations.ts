import { useQuery } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type { PaperCitationsResult } from '#/types/paper.type'
import { fetchHelper } from '#/utils/fetchHelper'

export function useGetPaperCitations(paperId: string | null | undefined) {
  return useQuery({
    queryKey: ['papers', 'citations', paperId],
    queryFn: async () => {
      if (!paperId) throw new Error('Paper ID is required')
      const response = await fetchHelper<ApiResponse<PaperCitationsResult>>(
        `/api/v1/papers/${encodeURIComponent(paperId)}/citations`,
        { method: 'GET' },
      )
      return response.data
    },
    enabled: Boolean(paperId),
    staleTime: 1000 * 60 * 10,
  })
}

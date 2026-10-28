import { useQuery } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type { GetAnglesResult } from '#/types/exploration.type'
import { fetchHelper } from '#/utils/fetchHelper'

export function useGetAngles(concept: string | null | undefined) {
  return useQuery({
    queryKey: ['exploration', 'angles', concept],
    queryFn: async () => {
      const url = concept
        ? `/api/v1/exploration/angles?concept=${encodeURIComponent(concept)}`
        : '/api/v1/exploration/angles'
      const response = await fetchHelper<ApiResponse<GetAnglesResult>>(url, {
        method: 'GET',
      })
      return response.data
    },
    enabled: Boolean(concept),
    staleTime: 1000 * 60 * 10,
  })
}

import { useQuery } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type { SystemHealth } from '#/types/exploration.type'
import { fetchHelper } from '#/utils/fetchHelper'

export function useHealthCheck() {
  return useQuery({
    queryKey: ['system', 'health'],
    queryFn: async () => {
      const response =
        await fetchHelper<ApiResponse<SystemHealth>>('/api/v1/health')
      return response.data
    },
    refetchInterval: 30000,
  })
}

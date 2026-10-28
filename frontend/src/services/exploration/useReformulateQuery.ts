import { useMutation } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type {
  ReformulateQueryPayload,
  ReformulateQueryResult,
} from '#/types/exploration.type'
import { fetchHelper } from '#/utils/fetchHelper'

export function useReformulateQuery() {
  return useMutation({
    mutationFn: async (payload: ReformulateQueryPayload) => {
      const response = await fetchHelper<
        ApiResponse<ReformulateQueryResult>,
        ReformulateQueryPayload
      >('/api/v1/exploration/reformulate', {
        method: 'POST',
        body: payload,
      })
      return response.data
    },
  })
}

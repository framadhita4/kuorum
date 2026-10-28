import { useMutation } from '@tanstack/react-query'

import type { ApiResponse } from '#/types/api.type'
import type { DecomposeConceptResult } from '#/types/exploration.type'
import { fetchHelper } from '#/utils/fetchHelper'

interface DecomposeConceptPayload {
  prompt: string
  model?: string
}

export function useDecomposeConcept() {
  return useMutation({
    mutationFn: async (payload: DecomposeConceptPayload) => {
      const response = await fetchHelper<
        ApiResponse<DecomposeConceptResult>,
        DecomposeConceptPayload
      >('/api/v1/exploration/decompose', {
        method: 'POST',
        body: payload,
      })
      return response.data
    },
  })
}

import { useQuery } from '@tanstack/react-query'

import { useAuth } from '../providers/AuthProvider'
import { fetchEntitlement } from '../services/billing'

export function useEntitlement() {
  const { accessToken, clearSession, userId } = useAuth()
  return useQuery({
    queryKey: ['entitlement', userId],
    queryFn: async () => {
      if (!accessToken) return null
      const resp = await fetchEntitlement(accessToken)
      if (resp.error) {
        if (resp.status === 401) {
          await clearSession()
          return null
        }
        throw new Error(resp.error.message)
      }
      return resp.data ?? null
    },
    enabled: !!accessToken,
    staleTime: 60 * 1000,
  })
}

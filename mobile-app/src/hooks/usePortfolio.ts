import { useQuery } from '@tanstack/react-query'

import { useAuth } from '../providers/AuthProvider'
import { fetchActivePortfolio } from '../services/portfolio'

export function useActivePortfolio() {
  const { accessToken, clearSession, userId } = useAuth()
  return useQuery({
    queryKey: ['portfolio', 'active', userId],
    queryFn: async () => {
      if (!accessToken) return null
      const resp = await fetchActivePortfolio(accessToken)
      if (resp.error) {
        if (resp.status === 401) {
          await clearSession()
          return null
        }
        if (resp.status === 404) return null
        throw new Error(resp.error.message)
      }
      return resp.data ?? null
    },
    enabled: !!accessToken,
    staleTime: 30 * 1000,
  })
}

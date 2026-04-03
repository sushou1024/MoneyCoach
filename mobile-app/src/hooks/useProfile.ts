import { useQuery } from '@tanstack/react-query'

import { useAuth } from '../providers/AuthProvider'
import { fetchProfile } from '../services/profile'

export function useProfile() {
  const { accessToken, clearSession, userId } = useAuth()
  return useQuery({
    queryKey: ['profile', userId],
    queryFn: async () => {
      if (!accessToken) return null
      const resp = await fetchProfile(accessToken)
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

import { useEffect, useState } from 'react'

import { useAuth } from '../providers/AuthProvider'
import { joinWaitlist } from '../services/waitlist'

type WaitlistState = {
  rank: number | null
  error: string | null
  isLoading: boolean
}

export function useWaitlistEntry(strategyId: string, calculationId: string) {
  const { accessToken } = useAuth()
  const [state, setState] = useState<WaitlistState>({ rank: null, error: null, isLoading: false })

  useEffect(() => {
    let isActive = true
    const submit = async () => {
      if (!accessToken || !strategyId || !calculationId) return
      setState((prev) => ({ ...prev, isLoading: true }))
      const resp = await joinWaitlist(accessToken, { strategy_id: strategyId, calculation_id: calculationId })
      if (!isActive) return
      if (resp.error) {
        const message = resp.error.message
        setState((prev) => ({ ...prev, error: message, isLoading: false }))
        return
      }
      setState((prev) => ({ ...prev, rank: resp.data?.rank ?? null, isLoading: false }))
    }
    submit()
    return () => {
      isActive = false
    }
  }, [accessToken, strategyId, calculationId])

  return {
    rank: state.rank,
    error: state.error,
    isLoading: state.isLoading,
    isReady: !state.error && state.rank !== null,
  }
}

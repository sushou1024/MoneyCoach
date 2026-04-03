import React, { createContext, useContext, useMemo } from 'react'

import { Entitlement } from '../types/api'

type PurchaseResult = Entitlement | null

type WebIAPItemDetails = {
  productId: string
  price?: string
  offerToken?: string
}

interface IAPContextValue {
  supported: boolean
  ready: boolean
  products: Record<string, WebIAPItemDetails>
  loadProducts: (productIds: string[]) => Promise<void>
  purchaseProduct: (productId: string) => Promise<PurchaseResult>
  restorePurchases: () => Promise<PurchaseResult>
}

const IAPContext = createContext<IAPContextValue | null>(null)

export function IAPProvider({ children }: { children: React.ReactNode }) {
  const value = useMemo<IAPContextValue>(
    () => ({
      supported: false,
      ready: false,
      products: {},
      loadProducts: async () => {},
      purchaseProduct: async () => {
        throw new Error('IAP not supported on web')
      },
      restorePurchases: async () => {
        throw new Error('IAP not supported on web')
      },
    }),
    []
  )

  return <IAPContext.Provider value={value}>{children}</IAPContext.Provider>
}

export function useIAP() {
  const ctx = useContext(IAPContext)
  if (!ctx) {
    throw new Error('useIAP must be used within IAPProvider')
  }
  return ctx
}

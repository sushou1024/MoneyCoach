import * as Crypto from 'expo-crypto'
import React, { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { Platform } from 'react-native'
import {
  endConnection,
  fetchProducts,
  finishTransaction,
  getAvailablePurchases,
  getReceiptDataIOS,
  initConnection,
  purchaseErrorListener,
  purchaseUpdatedListener,
  requestPurchase,
  type ProductSubscription,
  type Purchase,
  type PurchaseError,
  type PurchaseIOS,
  type SubscriptionOffer,
} from 'react-native-iap'

import { useAuth } from './AuthProvider'
import { submitAppleReceipt, submitGooglePlayReceipt } from '../services/billing'
import { Entitlement } from '../types/api'
import { isExpoGo } from '../utils/expoEnvironment'

type PurchaseResult = Entitlement | null

type IapProduct = {
  productId: string
  price?: string
  offerToken?: string
}

interface IAPContextValue {
  supported: boolean
  ready: boolean
  products: Record<string, IapProduct>
  loadProducts: (productIds: string[]) => Promise<void>
  purchaseProduct: (productId: string) => Promise<PurchaseResult>
  restorePurchases: () => Promise<PurchaseResult>
}

type PendingPurchase = {
  productId: string
  resolve: (value: PurchaseResult) => void
  reject: (error: Error) => void
}

type AccountIdentifiers = {
  obfuscatedAccountId: string
  obfuscatedProfileId: string
}

const IAPContext = createContext<IAPContextValue | null>(null)

const selectAndroidOfferToken = (offers?: SubscriptionOffer[] | null) => {
  if (!offers || offers.length === 0) return undefined
  const basePlan = offers.find((offer) => (offer.offerTagsAndroid?.length ?? 0) === 0)
  return basePlan?.offerTokenAndroid ?? offers[0]?.offerTokenAndroid ?? undefined
}

const buildProductMap = (items: ProductSubscription[]) => {
  const next: Record<string, IapProduct> = {}
  for (const item of items) {
    if (item.type !== 'subs') continue
    const productId = item.id
    const offerToken = Platform.OS === 'android' ? selectAndroidOfferToken(item.subscriptionOffers) : undefined
    next[productId] = {
      productId,
      price: item.displayPrice,
      offerToken,
    }
  }
  return next
}

export function IAPProvider({ children }: { children: React.ReactNode }) {
  const { accessToken, userId } = useAuth()
  const supportedPlatform = Platform.OS === 'ios' || Platform.OS === 'android'
  const supported = supportedPlatform && !isExpoGo()

  const [ready, setReady] = useState(false)
  const [products, setProducts] = useState<Record<string, IapProduct>>({})
  const [accountIdentifiers, setAccountIdentifiers] = useState<AccountIdentifiers | undefined>(undefined)

  const accessTokenRef = useRef<string | null>(null)
  const connectedRef = useRef(false)
  const pendingRef = useRef<PendingPurchase | null>(null)
  const listenerRef = useRef<{ purchase?: { remove: () => void }; error?: { remove: () => void } } | null>(null)

  useEffect(() => {
    accessTokenRef.current = accessToken
  }, [accessToken])

  useEffect(() => {
    if (!userId) {
      setAccountIdentifiers(undefined)
      return
    }
    let active = true
    const digest = async () => {
      const hashed = await Crypto.digestStringAsync(Crypto.CryptoDigestAlgorithm.SHA256, userId)
      if (active) {
        setAccountIdentifiers({ obfuscatedAccountId: hashed, obfuscatedProfileId: hashed })
      }
    }
    digest()
    return () => {
      active = false
    }
  }, [userId])

  const rejectPending = (err: Error) => {
    if (pendingRef.current) {
      pendingRef.current.reject(err)
      pendingRef.current = null
    }
  }

  const verifyPurchase = async (purchase: Purchase) => {
    const token = accessTokenRef.current
    if (!token) {
      throw new Error('Missing auth token')
    }
    if (purchase.purchaseState !== 'purchased') {
      throw new Error('Purchase not completed')
    }
    if (Platform.OS === 'ios') {
      const iosPurchase = purchase as PurchaseIOS
      const receiptData = await getReceiptDataIOS()
      if (!receiptData) {
        throw new Error('Missing receipt')
      }
      const transactionId = iosPurchase.transactionId || iosPurchase.originalTransactionIdentifierIOS || ''
      if (!transactionId) {
        throw new Error('Missing transaction id')
      }
      const resp = await submitAppleReceipt(token, {
        receiptData,
        productId: iosPurchase.productId,
        transactionId,
      })
      if (resp.error ?? !resp.data) {
        throw new Error(resp.error?.message ?? 'Receipt verification failed')
      }
      return resp.data
    }
    if (Platform.OS === 'android') {
      const purchaseToken = purchase.purchaseToken ?? ''
      if (!purchaseToken) {
        throw new Error('Missing purchase token')
      }
      const resp = await submitGooglePlayReceipt(token, {
        purchaseToken,
        productId: purchase.productId,
      })
      if (resp.error ?? !resp.data) {
        throw new Error(resp.error?.message ?? 'Purchase verification failed')
      }
      return resp.data
    }
    return null
  }

  const handlePurchaseUpdate = async (purchase: Purchase) => {
    if (!purchase || purchase.purchaseState !== 'purchased') return
    try {
      const entitlement = await verifyPurchase(purchase)
      await finishTransaction({ purchase, isConsumable: false })
      if (pendingRef.current?.productId === purchase.productId) {
        pendingRef.current.resolve(entitlement)
        pendingRef.current = null
      }
    } catch (err) {
      rejectPending(err instanceof Error ? err : new Error('Verification failed'))
    }
  }

  const handlePurchaseError = (error: PurchaseError) => {
    const message = error?.message || 'Purchase failed'
    rejectPending(new Error(message))
  }

  useEffect(() => {
    if (!supported || !accessToken) return
    let active = true
    const connect = async () => {
      try {
        await initConnection()
        connectedRef.current = true
        if (!active) return
        setReady(true)
        const purchaseSub = purchaseUpdatedListener((purchase) => {
          handlePurchaseUpdate(purchase).catch(() => {})
        })
        const errorSub = purchaseErrorListener(handlePurchaseError)
        listenerRef.current = { purchase: purchaseSub, error: errorSub }
      } catch {
        if (active) {
          setReady(false)
        }
      }
    }
    connect()
    return () => {
      active = false
      if (listenerRef.current) {
        listenerRef.current.purchase?.remove()
        listenerRef.current.error?.remove()
      }
      listenerRef.current = null
      if (connectedRef.current) {
        endConnection().catch(() => {})
        connectedRef.current = false
      }
      setReady(false)
    }
  }, [accessToken, supported])

  const ensureConnected = useCallback(async () => {
    if (!supported) {
      throw new Error('IAP not supported')
    }
    if (!connectedRef.current) {
      await initConnection()
      connectedRef.current = true
      setReady(true)
    }
  }, [supported])

  const loadProducts = useCallback(
    async (productIds: string[]) => {
      const ids = Array.from(new Set(productIds.filter((id) => id)))
      if (ids.length === 0) return
      await ensureConnected()
      const response = await fetchProducts({ skus: ids, type: 'subs' })
      const items = (response ?? []).filter((item): item is ProductSubscription => item.type === 'subs')
      const next = buildProductMap(items)
      setProducts((prev) => ({ ...prev, ...next }))
    },
    [ensureConnected]
  )

  const purchaseProduct = useCallback(
    async (productId: string) => {
      if (!productId) {
        throw new Error('Missing product id')
      }
      await ensureConnected()
      if (pendingRef.current) {
        throw new Error('Purchase in progress')
      }
      return new Promise<PurchaseResult>((resolve, reject) => {
        pendingRef.current = { productId, resolve, reject }
        const offerToken = Platform.OS === 'android' ? products[productId]?.offerToken : undefined
        if (Platform.OS === 'android' && !offerToken) {
          pendingRef.current = null
          reject(new Error('Missing subscription offer'))
          return
        }
        requestPurchase({
          type: 'subs',
          request:
            Platform.OS === 'android'
              ? {
                  google: {
                    skus: [productId],
                    subscriptionOffers: offerToken ? [{ sku: productId, offerToken }] : undefined,
                    obfuscatedAccountId: accountIdentifiers?.obfuscatedAccountId,
                    obfuscatedProfileId: accountIdentifiers?.obfuscatedProfileId,
                  },
                }
              : {
                  apple: {
                    sku: productId,
                  },
                },
        }).catch((err) => {
          pendingRef.current = null
          reject(err instanceof Error ? err : new Error('Purchase failed'))
        })
      })
    },
    [accountIdentifiers, ensureConnected, products]
  )

  const restorePurchases = useCallback(async () => {
    await ensureConnected()
    const purchases = await getAvailablePurchases({
      onlyIncludeActiveItemsIOS: true,
      includeSuspendedAndroid: false,
    })
    if (!purchases || purchases.length === 0) {
      return null
    }
    const sorted = purchases.slice().sort((a, b) => b.transactionDate - a.transactionDate)
    let restored: PurchaseResult = null
    for (const purchase of sorted) {
      try {
        restored = await verifyPurchase(purchase)
        await finishTransaction({ purchase, isConsumable: false })
        if (restored) {
          return restored
        }
      } catch {
        continue
      }
    }
    return restored
  }, [ensureConnected])

  const value = useMemo(
    () => ({
      supported,
      ready,
      products,
      loadProducts,
      purchaseProduct,
      restorePurchases,
    }),
    [supported, ready, products, loadProducts, purchaseProduct, restorePurchases]
  )

  return <IAPContext.Provider value={value}>{children}</IAPContext.Provider>
}

export function useIAP() {
  const ctx = useContext(IAPContext)
  if (!ctx) {
    throw new Error('IAPProvider missing')
  }
  return ctx
}

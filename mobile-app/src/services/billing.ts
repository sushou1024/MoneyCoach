import { apiRequest } from './api'
import { BillingPlan, Entitlement } from '../types/api'

export async function fetchBillingPlans(token: string) {
  return apiRequest<{ plans: BillingPlan[] }>('/v1/billing/plans', { token })
}

export async function fetchEntitlement(token: string) {
  return apiRequest<Entitlement>('/v1/billing/entitlement', { token })
}

export async function devActivateEntitlement(token: string) {
  return apiRequest<Entitlement>('/v1/billing/dev/entitlement', {
    method: 'POST',
    token,
    body: { status: 'active', plan_id: 'dev_pro' },
  })
}

export async function devClearEntitlement(token: string) {
  return apiRequest<Entitlement>('/v1/billing/dev/entitlement', {
    method: 'DELETE',
    token,
  })
}

export async function createStripeCheckoutSession(
  token: string,
  params: { planId: string; successUrl: string; cancelUrl: string }
) {
  return apiRequest<{ checkout_url: string }>('/v1/billing/stripe/session', {
    method: 'POST',
    token,
    body: {
      plan_id: params.planId,
      success_url: params.successUrl,
      cancel_url: params.cancelUrl,
    },
  })
}

export async function submitAppleReceipt(
  token: string,
  params: { receiptData: string; productId: string; transactionId: string }
) {
  return apiRequest<Entitlement>('/v1/billing/receipt/ios', {
    method: 'POST',
    token,
    body: {
      receipt_data: params.receiptData,
      product_id: params.productId,
      transaction_id: params.transactionId,
    },
  })
}

export async function submitGooglePlayReceipt(token: string, params: { purchaseToken: string; productId: string }) {
  return apiRequest<Entitlement>('/v1/billing/receipt/android', {
    method: 'POST',
    token,
    body: {
      purchase_token: params.purchaseToken,
      product_id: params.productId,
    },
  })
}

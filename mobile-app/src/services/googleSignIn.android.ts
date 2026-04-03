import {
  GoogleSignin,
  isCancelledResponse,
  isErrorWithCode,
  isSuccessResponse,
  statusCodes,
} from '@react-native-google-signin/google-signin'

export type GoogleSignInResult = {
  idToken: string
}

export class GoogleSignInCancelledError extends Error {
  constructor() {
    super('Google sign-in was cancelled')
    this.name = 'GoogleSignInCancelledError'
  }
}

export class GoogleSignInConfigError extends Error {
  readonly missingVars: string[]

  constructor(missingVars: string[]) {
    super(`Google sign-in missing env vars: ${missingVars.join(', ')}`)
    this.name = 'GoogleSignInConfigError'
    this.missingVars = missingVars
  }
}

let configuredWebClientId: string | null = null

const GOOGLE_SCOPES = ['email', 'profile']

function ensureConfigured(webClientId: string) {
  const trimmedClientId = webClientId.trim()
  if (!trimmedClientId) {
    throw new GoogleSignInConfigError(['EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID'])
  }
  if (configuredWebClientId === trimmedClientId) {
    return
  }
  GoogleSignin.configure({
    webClientId: trimmedClientId,
    scopes: GOOGLE_SCOPES,
    offlineAccess: false,
  })
  configuredWebClientId = trimmedClientId
}

export async function signInWithGoogle(options: { webClientId: string }): Promise<GoogleSignInResult> {
  ensureConfigured(options.webClientId)
  await GoogleSignin.hasPlayServices({ showPlayServicesUpdateDialog: true })
  const response = await GoogleSignin.signIn()
  if (isCancelledResponse(response)) {
    throw new GoogleSignInCancelledError()
  }
  if (!isSuccessResponse(response)) {
    throw new Error('Google sign-in did not complete successfully')
  }

  const idToken = response.data.idToken?.trim()
  if (!idToken) {
    throw new Error('Google sign-in did not return an id token')
  }

  return { idToken }
}

export function isGoogleSignInCancelledError(error: unknown): error is GoogleSignInCancelledError {
  if (error instanceof GoogleSignInCancelledError) {
    return true
  }
  return isErrorWithCode(error) && [statusCodes.SIGN_IN_CANCELLED, statusCodes.IN_PROGRESS].includes(error.code)
}

export function isGoogleSignInConfigError(error: unknown): error is GoogleSignInConfigError {
  return error instanceof GoogleSignInConfigError
}

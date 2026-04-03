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

export async function signInWithGoogle(_: { webClientId: string }): Promise<GoogleSignInResult> {
  throw new Error('Native Google sign-in is only available on Android')
}

export function isGoogleSignInCancelledError(error: unknown): error is GoogleSignInCancelledError {
  return error instanceof GoogleSignInCancelledError
}

export function isGoogleSignInConfigError(error: unknown): error is GoogleSignInConfigError {
  return error instanceof GoogleSignInConfigError
}

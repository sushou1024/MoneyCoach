import 'react-native-gesture-handler/jestSetup'

//

jest.mock('@react-native-async-storage/async-storage', () =>
  require('@react-native-async-storage/async-storage/jest/async-storage-mock')
)

jest.mock('expo-file-system', () => {
  class FileMock {
    uri: string

    constructor(uri: string) {
      this.uri = uri
    }

    bytes = jest.fn(async () => new Uint8Array())

    info = jest.fn(() => ({ exists: true, size: 0 }))

    get size() {
      return 0
    }
  }

  return { File: FileMock }
})

jest.mock('expo-crypto', () => ({
  digestStringAsync: jest.fn(async () => 'mocked-digest'),
  digest: jest.fn(async () => new ArrayBuffer(0)),
  CryptoDigestAlgorithm: { SHA256: 'SHA256' },
  CryptoEncoding: { BASE64: 'base64', HEX: 'hex' },
}))

jest.mock('expo-localization', () => ({
  getLocales: () => [{ languageTag: 'en', languageCode: 'en' }],
  timezone: 'UTC',
}))

jest.mock('expo-router', () => ({
  useRouter: () => ({ push: jest.fn(), replace: jest.fn(), back: jest.fn() }),
  useLocalSearchParams: () => ({}),
  Redirect: () => null,
}))

jest.mock('@react-native-google-signin/google-signin', () => ({
  GoogleSignin: {
    configure: jest.fn(),
    hasPlayServices: jest.fn(async () => true),
    signIn: jest.fn(async () => ({
      type: 'success',
      data: {
        user: {
          id: 'google-user-id',
          name: 'Money Coach Reviewer',
          email: 'reviewer@example.com',
          photo: null,
          familyName: 'Reviewer',
          givenName: 'Money Coach',
        },
        scopes: ['email', 'profile'],
        idToken: 'mock-google-id-token',
        serverAuthCode: null,
      },
    })),
    signOut: jest.fn(async () => null),
    revokeAccess: jest.fn(async () => null),
    hasPreviousSignIn: jest.fn(() => false),
    getCurrentUser: jest.fn(() => null),
    clearCachedAccessToken: jest.fn(async () => null),
    getTokens: jest.fn(async () => ({ idToken: 'mock-google-id-token', accessToken: 'mock-access-token' })),
  },
  statusCodes: {
    SIGN_IN_CANCELLED: 'SIGN_IN_CANCELLED',
    IN_PROGRESS: 'IN_PROGRESS',
    PLAY_SERVICES_NOT_AVAILABLE: 'PLAY_SERVICES_NOT_AVAILABLE',
    SIGN_IN_REQUIRED: 'SIGN_IN_REQUIRED',
    NULL_PRESENTER: 'NULL_PRESENTER',
  },
  isErrorWithCode: (error: unknown) => typeof error === 'object' && error !== null && 'code' in error,
  isSuccessResponse: (response: { type?: string }) => response?.type === 'success',
  isCancelledResponse: (response: { type?: string }) => response?.type === 'cancelled',
}))

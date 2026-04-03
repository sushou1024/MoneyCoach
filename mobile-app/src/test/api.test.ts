import { getApiBaseUrl } from '../services/api'

describe('getApiBaseUrl', () => {
  it('uses the env override when provided', () => {
    process.env.EXPO_PUBLIC_API_BASE_URL = 'http://example.com'
    expect(getApiBaseUrl()).toBe('http://example.com')
  })

  afterEach(() => {
    delete process.env.EXPO_PUBLIC_API_BASE_URL
  })
})

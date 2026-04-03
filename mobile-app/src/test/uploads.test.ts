import { reviewUploadBatch } from '../services/uploads'

function mockFetch(status: number, body: any) {
  global.fetch = jest.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    headers: { get: () => 'application/json' },
    json: async () => body,
  } as any)
}

describe('reviewUploadBatch', () => {
  beforeEach(() => {
    process.env.EXPO_PUBLIC_API_BASE_URL = 'http://example.com'
  })

  afterEach(() => {
    delete process.env.EXPO_PUBLIC_API_BASE_URL
  })

  it('posts review payload with edits and resolutions', async () => {
    mockFetch(200, { status: 'processing' })

    const payload = {
      platform_overrides: [],
      resolutions: [
        {
          symbol_raw: 'ABC',
          asset_type: 'stock',
          symbol: 'ABC',
          asset_key: 'stock:ABC',
          exchange_mic: 'XNAS',
        },
      ],
      edits: [
        {
          asset_id: 'asset_1',
          action: 'remove',
        },
      ],
      duplicate_overrides: [],
    }

    await reviewUploadBatch('token', 'batch_1', payload)

    const [url, init] = (global.fetch as jest.Mock).mock.calls[0]
    expect(url).toBe('http://example.com/v1/upload-batches/batch_1/review')
    expect(init.method).toBe('POST')
    expect(JSON.parse(init.body)).toEqual(payload)
  })
})

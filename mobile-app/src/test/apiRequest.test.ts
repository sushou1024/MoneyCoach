import { apiRequest } from '../services/api'

function mockFetch(status: number, body: any) {
  global.fetch = jest.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    headers: { get: () => 'application/json' },
    json: async () => body,
  } as any)
}

describe('apiRequest', () => {
  it('returns api error payload even when status is accepted', async () => {
    mockFetch(202, { error: { code: 'NOT_READY', message: 'preview not ready' } })
    const resp = await apiRequest('/v1/reports/preview/123')
    expect(resp.status).toBe(202)
    expect(resp.error?.code).toBe('NOT_READY')
  })
})

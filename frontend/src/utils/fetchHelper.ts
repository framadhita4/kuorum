export interface FetchOptions<TBody = unknown> {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  headers?: Record<string, string>
  body?: TBody
  signal?: AbortSignal
}

export async function fetchHelper<TResponse, TBody = unknown>(
  url: string,
  options: FetchOptions<TBody> = {},
): Promise<TResponse> {
  const { method = 'GET', headers = {}, body, signal } = options

  const requestHeaders: Record<string, string> = {
    'Content-Type': 'application/json',
    ...headers,
  }

  const config: RequestInit = {
    method,
    headers: requestHeaders,
    signal,
  }

  if (body !== undefined && method !== 'GET') {
    config.body = JSON.stringify(body)
  }

  const res = await fetch(url, config)

  if (!res.ok) {
    let errorMessage = `HTTP error! status: ${res.status}`
    try {
      const errorJson = (await res.json()) as {
        error?: string
        message?: string
      }
      if (errorJson.error) {
        errorMessage = String(errorJson.error)
      } else if (errorJson.message) {
        errorMessage = String(errorJson.message)
      }
    } catch {
      // JSON parse failed, use fallback message
    }
    throw new Error(errorMessage)
  }

  return res.json() as Promise<TResponse>
}

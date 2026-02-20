import ky from 'ky';
import pLimit from 'p-limit';

const limit = pLimit(5);

const api = ky.create({
  prefixUrl: '',
  timeout: 30000,
  retry: {
    limit: 3,
    methods: ['get', 'post', 'put', 'delete'],
    statusCodes: [404, 408, 429, 500, 502, 503, 504],
  },
});

let csrfToken: string | null = null;

export async function fetchCSRFToken(): Promise<string> {
  const response = await ky.get('/api/csrf').json<{ token: string }>();
  csrfToken = response.token;
  return response.token;
}

export function getCSRFToken(): string | null {
  return csrfToken;
}

export function updateCSRFToken(token: string | null): void {
  csrfToken = token;
}

export async function ensureCSRFToken(): Promise<string> {
  const token = getCSRFToken();
  if (token) return token;
  return fetchCSRFToken();
}

export async function fetchWithRetry<T>(url: string, options: RequestInit = {}): Promise<T> {
  return api.get(url, { ...options }).json<T>();
}

export async function fetchWithCSRF<T>(url: string, options: { json?: unknown } = {}): Promise<T> {
  const csrf = await ensureCSRFToken();
  return api
    .post(url, {
      json: options.json,
      headers: {
        'X-CSRF-Token': csrf,
      },
    })
    .json<T>();
}

export async function fetchWithCSRFAndFormData(url: string, formData: FormData): Promise<void> {
  const csrf = await ensureCSRFToken();
  formData.set('csrf_token', csrf);
  await api.post(url, {
    body: formData,
    headers: {
      'X-CSRF-Token': csrf,
    },
  });
}

export { limit };

import { z } from 'zod';

const token = { value: null as string | null };

const csrfTokenSchema = z.object({ token: z.string() });

export async function fetchCSRFToken(): Promise<string> {
  const response = await fetch('/api/csrf');
  if (!response.ok) throw new Error('Failed to fetch CSRF token');
  const data = csrfTokenSchema.parse(await response.json());
  token.value = data.token;
  return data.token;
}

function getCSRFToken(): string | null {
  return token.value;
}

function updateCSRFToken(response: Response): void {
  const newToken = response.headers.get('X-CSRF-Token');
  if (newToken) token.value = newToken;
}

async function ensureCSRFToken(): Promise<string> {
  const t = getCSRFToken();
  if (t) return t;
  return fetchCSRFToken();
}

async function fetchWithRetry(
  url: string,
  options: RequestInit = {},
  retries = 3
): Promise<Response> {
  let lastError: Error | null = null;

  for (let i = 0; i < retries; i++) {
    try {
      const response = await fetch(url, options);
      if (response.ok) return response;

      if (response.status === 404 && i < retries - 1) {
        lastError = new Error(`API endpoint not ready, retrying... (${i + 1}/${retries})`);
        await new Promise((r) => setTimeout(r, 1000 * (i + 1)));
        continue;
      }

      lastError = new Error(`Request failed: ${response.statusText}`);
      throw lastError;
    } catch (err) {
      lastError = err instanceof Error ? err : new Error('Unknown error');
      if (i < retries - 1) await new Promise((r) => setTimeout(r, 1000 * (i + 1)));
    }
  }

  throw lastError || new Error('Request failed after retries');
}

export const documentSchema = z.object({
  url: z.string().url(),
  domain: z.string(),
  title: z.string(),
  text: z.string(),
  favicon: z.string(),
  score: z.number(),
  added: z.number(),
});

export type Document = z.infer<typeof documentSchema>;

export const urlCountSchema = z.object({
  url: z.string(),
  title: z.string(),
  count: z.number().int().min(0),
});

export type URLCount = z.infer<typeof urlCountSchema>;

export const searchQuerySchema = z.object({
  text: z.string().min(1),
  highlight: z.string().optional(),
  fields: z.array(z.string()).optional(),
  limit: z.number().int().min(1).optional(),
  sort: z.string().optional(),
  date_from: z.number().int().optional(),
  date_to: z.number().int().optional(),
});

export type SearchQuery = z.infer<typeof searchQuerySchema>;

export const searchResultsSchema = z.object({
  total: z.number().int().min(0),
  query: searchQuerySchema,
  documents: z.array(documentSchema),
  history: z.array(urlCountSchema),
  search_duration: z.string(),
  query_suggestion: z.string(),
});

export type SearchResults = z.infer<typeof searchResultsSchema>;

export const rulesSchema = z.object({
  skip: z.array(z.string()),
  priority: z.array(z.string()),
  aliases: z.record(z.string(), z.string()),
});

export type Rules = z.infer<typeof rulesSchema>;

export const historyItemSchema = z.object({
  query: z.string(),
  title: z.string(),
  url: z.string(),
  favicon: z.string().optional(),
});

export type HistoryItem = z.infer<typeof historyItemSchema>;

export const addEntryRequestSchema = z.object({
  url: z.string().url(),
  title: z.string(),
  text: z.string(),
});

export type AddEntryRequest = z.infer<typeof addEntryRequestSchema>;

const wsCallbacks = new Map<string, (results: SearchResults) => void>();
let ws: WebSocket | null = null;

export function initWebSocket(): WebSocket {
  if (ws?.readyState === WebSocket.OPEN) return ws;

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/search`;

  ws = new WebSocket(wsUrl);

  ws.onopen = () => console.log('WebSocket connected');

  ws.onmessage = (event) => {
    try {
      const data = searchResultsSchema.parse(JSON.parse(event.data));
      const callback = wsCallbacks.get(data.query.text);
      callback?.(data);
    } catch (err) {
      console.error('Failed to parse WebSocket message:', err);
    }
  };

  ws.onerror = (error) => console.error('WebSocket error:', error);

  ws.onclose = () => {
    console.log('WebSocket disconnected');
    ws = null;
  };

  return ws;
}

export function search(query: SearchQuery, callback: (results: SearchResults) => void): void {
  const validatedQuery = searchQuerySchema.parse(query);
  const socket = initWebSocket();

  const send = () => {
    wsCallbacks.set(validatedQuery.text, callback);
    socket.send(JSON.stringify(validatedQuery));
  };

  if (socket.readyState === WebSocket.OPEN) {
    send();
    return;
  }

  socket.onopen = send;
}

export async function searchHttp(query: SearchQuery): Promise<SearchResults> {
  const validatedQuery = searchQuerySchema.parse(query);
  const params = new URLSearchParams({ q: validatedQuery.text });

  if (validatedQuery.date_from)
    params.set('date_from', new Date(validatedQuery.date_from * 1000).toISOString().split('T')[0]);
  if (validatedQuery.date_to)
    params.set('date_to', new Date(validatedQuery.date_to * 1000).toISOString().split('T')[0]);

  const response = await fetchWithRetry(`/search?${params.toString()}`, {
    headers: { Accept: 'application/json' },
  });
  return searchResultsSchema.parse(await response.json());
}

export async function fetchRules(): Promise<Rules> {
  const response = await fetchWithRetry('/api/rules', { headers: { Accept: 'application/json' } });
  return rulesSchema.parse(await response.json());
}

async function fetchWithCSRF(url: string, options: RequestInit): Promise<void> {
  const csrf = await ensureCSRFToken();
  const response = await fetch(url, {
    ...options,
    headers: { ...options.headers, 'X-CSRF-Token': csrf },
  });
  if (!response.ok) throw new Error(`Request failed: ${response.statusText}`);
  updateCSRFToken(response);
}

async function fetchWithCSRFAndFormData(url: string, formData: FormData): Promise<void> {
  const csrf = await ensureCSRFToken();
  formData.set('csrf_token', csrf);
  const response = await fetch(url, {
    method: 'POST',
    body: formData,
    headers: { 'X-CSRF-Token': csrf },
  });
  if (!response.ok) throw new Error(`Request failed: ${response.statusText}`);
  updateCSRFToken(response);
}

export async function saveRules(rules: Rules): Promise<void> {
  const validatedRules = rulesSchema.parse(rules);
  const csrf = await ensureCSRFToken();
  const response = await fetch('/api/rules', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrf },
    body: JSON.stringify(validatedRules),
  });
  if (!response.ok) throw new Error(`Failed to save rules: ${response.statusText}`);
  updateCSRFToken(response);
}

export async function addAlias(keyword: string, value: string): Promise<void> {
  const formData = new FormData();
  formData.set('alias-keyword', keyword);
  formData.set('alias-value', value);
  await fetchWithCSRFAndFormData('/add_alias', formData);
}

export async function deleteAlias(alias: string): Promise<void> {
  const formData = new FormData();
  formData.set('alias', alias);
  await fetchWithCSRFAndFormData('/delete_alias', formData);
}

export async function addEntry(entry: AddEntryRequest): Promise<void> {
  const validatedEntry = addEntryRequestSchema.parse(entry);
  await fetchWithCSRF('/add', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(validatedEntry),
  });
}

export async function fetchHistory(limit = 40): Promise<HistoryItem[]> {
  const response = await fetchWithRetry(`/api/history?limit=${limit}`, {
    headers: { Accept: 'application/json' },
  });
  return z.array(historyItemSchema).parse(await response.json());
}

export async function updateHistory(query: string, url: string, title: string): Promise<void> {
  await fetchWithCSRF('/history', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query, url, title }),
  });
}

export async function trackSearch(query: string): Promise<void> {
  const searchUrl = `/search?q=${encodeURIComponent(query)}`;
  await updateHistory(query, searchUrl, `Search: ${query}`);
}

export async function deleteHistoryItem(query: string, url: string): Promise<void> {
  await fetchWithCSRF('/history', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query, url, delete: true }),
  });
}

export async function deleteDocument(url: string): Promise<void> {
  const formData = new FormData();
  formData.set('url', url);
  await fetchWithCSRFAndFormData('/delete', formData);
}

export const statsSchema = z.object({
  pagesIndexed: z.number().int().min(0),
  domains: z.number().int().min(0),
  dateRange: z.string(),
  minDate: z.number().int().optional(),
  maxDate: z.number().int().optional(),
});

export type Stats = z.infer<typeof statsSchema>;

// Global stats store for background loading
let globalStats: Stats | null = null;
let statsPromise: Promise<Stats> | null = null;
let lastFetchTime = 0;
const STATS_CACHE_TTL = 5 * 60 * 1000; // 5 minutes
let refreshInterval: ReturnType<typeof setInterval> | null = null;

export async function fetchStats(forceRefresh = false): Promise<Stats> {
  const now = Date.now();

  // Return cached stats if fresh and not forcing refresh
  if (globalStats && !forceRefresh && now - lastFetchTime < STATS_CACHE_TTL) {
    return globalStats;
  }

  if (statsPromise) {
    return statsPromise;
  }

  statsPromise = (async () => {
    try {
      const response = await fetchWithRetry('/api/stats', {
        headers: { Accept: 'application/json' },
      });
      globalStats = statsSchema.parse(await response.json());
      lastFetchTime = now;
      return globalStats;
    } finally {
      statsPromise = null;
    }
  })();

  return statsPromise;
}

export function initBackgroundStats() {
  fetchStats();

  if (!refreshInterval) {
    refreshInterval = setInterval(() => {
      fetchStats(true).catch((err) => {
        console.error('Background stats refresh failed:', err);
      });
    }, STATS_CACHE_TTL);
  }
}

export function cleanupBackgroundStats() {
  if (refreshInterval) {
    clearInterval(refreshInterval);
    refreshInterval = null;
  }
}

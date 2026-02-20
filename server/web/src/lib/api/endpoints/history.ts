import { fetchWithRetry, fetchWithCSRF } from '../client';
import { parse, array } from 'valibot';
import { historyItemSchema, type HistoryItem } from './schemas';

export async function fetchHistory(limit = 40): Promise<HistoryItem[]> {
  const response = await fetchWithRetry<HistoryItem[]>(`/api/history?limit=${limit}`);
  return parse(array(historyItemSchema), response);
}

export async function updateHistory(query: string, url: string, title: string): Promise<void> {
  await fetchWithCSRF<void>('/history', {
    json: { query, url, title },
  });
}

export async function trackSearch(query: string): Promise<void> {
  await updateHistory(query, `/search?q=${encodeURIComponent(query)}`, `Search: ${query}`);
}

export async function deleteHistoryItem(query: string, url: string): Promise<void> {
  await fetchWithCSRF<void>('/history', {
    json: { query, url, delete: true },
  });
}

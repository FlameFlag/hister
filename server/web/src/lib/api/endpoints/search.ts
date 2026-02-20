import { fetchWithRetry } from '../client';
import { parse } from 'valibot';
import {
  searchQuerySchema,
  searchResultsSchema,
  type SearchQuery,
  type SearchResults,
} from './schemas';

export async function search(query: SearchQuery): Promise<SearchResults> {
  const validated = parse(searchQuerySchema, query);
  const params = new URLSearchParams({ q: validated.text });

  if (validated.date_from) {
    params.set('date_from', new Date(validated.date_from * 1000).toISOString().split('T')[0]);
  }
  if (validated.date_to) {
    params.set('date_to', new Date(validated.date_to * 1000).toISOString().split('T')[0]);
  }

  return parse(
    searchResultsSchema,
    await fetchWithRetry<SearchResults>(`/search?${params.toString()}`)
  );
}

export { search as searchWebSocket };
export type { SearchQuery, SearchResults };

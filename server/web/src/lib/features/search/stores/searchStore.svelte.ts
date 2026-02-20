import type { SearchResults } from '$lib/api';

export interface SearchState {
  query: string;
  results: SearchResults | null;
  loading: boolean;
  sortBy: string;
  dateFrom: number | undefined;
  dateTo: number | undefined;
}

const initialState: SearchState = {
  query: '',
  results: null,
  loading: false,
  sortBy: 'relevance',
  dateFrom: undefined,
  dateTo: undefined,
};

export function createSearchStore() {
  let state = $state<SearchState>({ ...initialState });

  return {
    get state() {
      return state;
    },
    reset() {
      state = { ...initialState };
    },
  };
}

export type SearchStore = ReturnType<typeof createSearchStore>;

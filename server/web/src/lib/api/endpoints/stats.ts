import { fetchWithRetry } from '../client';
import { parse } from 'valibot';
import { statsSchema, type Stats } from './schemas';
import { format, fromUnixTime } from 'date-fns';

export const STATS_CACHE_TTL = 5 * 60 * 1000;

let globalStats: Stats | null = null;
let statsPromise: Promise<Stats> | null = null;
let lastFetchTime = 0;
let refreshInterval: ReturnType<typeof setInterval> | null = null;

export async function fetchStats(forceRefresh = false): Promise<Stats> {
  const now = Date.now();

  if (globalStats && !forceRefresh && now - lastFetchTime < STATS_CACHE_TTL) {
    return globalStats;
  }

  if (statsPromise) {
    return statsPromise;
  }

  statsPromise = (async () => {
    try {
      const data = parse(statsSchema, await fetchWithRetry<Stats>('/api/stats'));
      const result = { ...data, dateRange: data.dateRange || 'Unknown' };

      if (data.minDate && data.maxDate) {
        result.dateRange = `${format(fromUnixTime(data.minDate), 'MMM d, yyyy')} - ${format(fromUnixTime(data.maxDate), 'MMM d, yyyy')}`;
      }

      globalStats = result;
      lastFetchTime = now;
      return result;
    } finally {
      statsPromise = null;
    }
  })();

  return statsPromise;
}

export function initBackgroundStats(): void {
  fetchStats();

  if (!refreshInterval) {
    refreshInterval = setInterval(() => {
      fetchStats(true).catch((err) => {
        console.error('Background stats refresh failed:', err);
      });
    }, STATS_CACHE_TTL);
  }
}

export function cleanupBackgroundStats(): void {
  if (refreshInterval) {
    clearInterval(refreshInterval);
    refreshInterval = null;
  }
}

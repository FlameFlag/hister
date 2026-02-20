<script lang="ts">
  import { onMount } from 'svelte';
  import { Clock, Trash2 } from 'lucide-svelte';
  import SearchInput from '$lib/features/search/components/SearchInput.svelte';
  import { fetchStats, fetchHistory, deleteHistoryItem, type HistoryItem } from '$lib/api';
  import * as Tooltip from '$lib/components/ui/tooltip';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { goto } from '$app/navigation';

  let searchQuery = $state('');
  let recentSearches = $state<string[]>([]);
  let stats = $state({
    pagesIndexed: 0,
    domains: 0,
    dateRange: 'Last 30 days',
  });
  let loading = $state(true);
  let historyItems = $state<HistoryItem[]>([]);
  let tooltipShown = $state(false);

  const STATS_CACHE_KEY = 'hister_stats_cache';
  const STATS_CACHE_DURATION = 5 * 60 * 1000;

  function getCachedStats(): { data: unknown; timestamp: number } | null {
    try {
      const cached = localStorage.getItem(STATS_CACHE_KEY);
      if (cached) {
        const parsed = JSON.parse(cached) as {
          data: unknown;
          timestamp: number;
        };
        if (Date.now() - parsed.timestamp < STATS_CACHE_DURATION) {
          return parsed;
        }
      }
    } catch (e) {
      console.error('Failed to parse cached stats:', e);
    }
    return null;
  }

  function setCachedStats(data: unknown): void {
    try {
      localStorage.setItem(
        STATS_CACHE_KEY,
        JSON.stringify({
          data,
          timestamp: Date.now(),
        })
      );
    } catch (e) {
      console.error('Failed to cache stats:', e);
    }
  }

  onMount(async () => {
    try {
      const cachedStats = getCachedStats();

      const [statsData, historyData] = await Promise.all([
        cachedStats ? Promise.resolve(cachedStats.data as typeof stats) : fetchStats(),
        fetchHistory(10),
      ]);

      stats = statsData;

      if (!cachedStats) {
        setCachedStats(statsData);
      }

      historyItems = historyData;
      const queries = historyData
        .map((h: HistoryItem) => h.query)
        .filter((q: string): q is string => q != null && q.trim() !== '') as string[];
      recentSearches = [...new Set(queries)].slice(0, 5);
    } catch (err) {
      console.error('Failed to load home data:', err);
    } finally {
      loading = false;
    }
  });

  function handleInput() {
    if (searchQuery.trim()) {
      goto(`/search?q=${encodeURIComponent(searchQuery)}`, { keepFocus: true });
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter' && searchQuery.trim()) {
      goto(`/search?q=${encodeURIComponent(searchQuery)}`, { keepFocus: true });
    }
  }

  function setSearchQuery(term: string) {
    searchQuery = term;
    goto(`/search?q=${encodeURIComponent(term)}`, { keepFocus: true });
  }

  function formatNumber(num: number): string {
    return num.toLocaleString();
  }

  async function handleDeleteSearch(term: string) {
    const historyItem = historyItems.find((h) => h.query === term);
    if (!historyItem) return;

    try {
      await deleteHistoryItem(term, historyItem.url);
      recentSearches = recentSearches.filter((s) => s !== term);
      historyItems = historyItems.filter((h) => h.query !== term);
    } catch (err) {
      console.error('Failed to delete search:', err);
    }
  }

  async function handleDeleteAllSearches() {
    if (!confirm('Are you sure you want to delete all recent searches?')) {
      return;
    }

    try {
      await Promise.all(recentSearches.map((term) => handleDeleteSearch(term)));
      recentSearches = [];
      historyItems = [];
    } catch (err) {
      console.error('Failed to delete all searches:', err);
    }
  }
</script>

<svelte:head>
  <title>Home - Hister</title>
</svelte:head>

<main class="flex flex-1 flex-col items-center justify-center p-12">
  <h1
    class="mb-8 flex items-center gap-3 font-display text-[32px] font-extrabold tracking-[-1px] text-foreground"
  >
    hister
  </h1>

  <div class="w-full max-w-200">
    <SearchInput
      bind:value={searchQuery}
      placeholder="Search your browsing history..."
      oninput={handleInput}
      onkeydown={handleKeyDown}
    />
  </div>

  {#if recentSearches.length > 0}
    <section class="mt-8 w-full max-w-200">
      <header class="mb-3 flex items-center justify-center gap-2">
        <h2
          class="
            text-center text-xs font-medium tracking-[0.5px]
            text-muted-foreground uppercase
          "
        >
          Recent Searches
        </h2>
        <Button
          variant="ghost"
          size="sm"
          onclick={handleDeleteAllSearches}
          class="
            flex cursor-pointer items-center gap-1 text-xs text-muted-foreground
            hover:bg-destructive/10 hover:text-destructive
          "
        >
          <Trash2 class="size-3" />
          Clear All
        </Button>
      </header>
      <div class="flex flex-wrap justify-center gap-2">
        {#each recentSearches as term (term)}
          <Tooltip.Root
            open={tooltipShown ? false : undefined}
            onOpenChange={(isOpen) => {
              if (isOpen) tooltipShown = true;
            }}
          >
            <Tooltip.Trigger>
              <Button
                variant="outline"
                size="sm"
                class="
                  cursor-pointer rounded-[24px] border-border bg-card
                  text-muted-foreground
                  hover:bg-secondary hover:text-foreground
                "
                onclick={() => setSearchQuery(term)}
              >
                <Clock class="mr-2 size-3.5" />
                {term}
              </Button>
            </Tooltip.Trigger>
            <Tooltip.Content>
              <p class="text-xs">Right-click to delete</p>
            </Tooltip.Content>
          </Tooltip.Root>
        {/each}
      </div>
    </section>
  {/if}
</main>

<footer
  class="
    mt-auto flex h-12 items-center justify-center gap-12 border-t border-border
    bg-muted text-sm
  "
>
  {#if loading}
    <span class="text-muted-foreground">Loading stats…</span>
  {:else}
    <span class="flex items-center gap-2">
      <Badge variant="secondary" class="font-display text-base font-semibold"
        >{formatNumber(stats.pagesIndexed)}</Badge
      >
      <span class="text-muted-foreground">pages indexed</span>
    </span>
    <span class="flex items-center gap-2">
      <Badge variant="secondary" class="font-display text-base font-semibold"
        >{formatNumber(stats.domains)}</Badge
      >
      <span class="text-muted-foreground">domains</span>
    </span>
    <span class="font-medium text-muted-foreground">{stats.dateRange}</span>
  {/if}
</footer>

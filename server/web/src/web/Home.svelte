<script lang="ts">
  import { onMount } from 'svelte';
  import { Search, Clock, Moon, Sun } from 'lucide-svelte';
  import { toggleMode } from 'mode-watcher';
  import { fetchStats, fetchHistory, type HistoryItem } from '$lib/api';
  import { Button, Input, Badge } from '$lib/components/ui';
  import { Separator } from '$lib/components/ui/separator';
  import * as NavigationMenu from '$lib/components/ui/navigation-menu';

  let searchQuery = $state('');
  let recentSearches = $state<string[]>([]);
  let stats = $state({
    pagesIndexed: 0,
    domains: 0,
    dateRange: 'Last 30 days',
  });
  let loading = $state(true);

  onMount(async () => {
    try {
      const [statsData, historyData] = await Promise.all([fetchStats(), fetchHistory(10)]);

      stats = statsData;

      // Extract unique queries from history (only items with a query field, which represents searches that led to opening links)
      const queries = historyData
        .map((h: HistoryItem) => h.query)
        .filter((q: string) => q && q.trim() !== '');
      recentSearches = [...new Set(queries)].slice(0, 5);
    } catch (err) {
      console.error('Failed to load home data:', err);
    } finally {
      loading = false;
    }
  });

  function handleInput() {
    // Transition to search page on any input
    if (searchQuery.trim()) {
      window.location.href = `/search?q=${encodeURIComponent(searchQuery)}`;
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter' && searchQuery.trim()) {
      window.location.href = `/search?q=${encodeURIComponent(searchQuery)}`;
    }
  }

  function setSearchQuery(term: string) {
    window.location.href = `/search?q=${encodeURIComponent(term)}`;
  }

  function formatNumber(num: number): string {
    return num.toLocaleString();
  }
</script>

<svelte:head>
  <title>Hister - Search Your Browsing History</title>
</svelte:head>

<div class="flex min-h-screen flex-col bg-background font-sans antialiased">
  <!-- Header -->
  <header
    class="
      flex h-18 items-center justify-between border-b border-border
      bg-background px-12
    "
  >
    <!-- Logo Section -->
    <div class="flex items-center gap-2">
      <span
        class="
          font-display text-[28px] font-bold tracking-[-0.5px] text-primary
        ">Hister</span
      >
    </div>

    <!-- Navigation -->
    <nav class="flex items-center gap-2">
      <Badge variant="default">History</Badge>
      <NavigationMenu.Root>
        <NavigationMenu.List>
          <NavigationMenu.Item>
            <NavigationMenu.Link onclick={() => (window.location.href = '/rules')}>
              Rules
            </NavigationMenu.Link>
          </NavigationMenu.Item>
          <NavigationMenu.Item>
            <NavigationMenu.Link onclick={() => (window.location.href = '/add')}>
              Add
            </NavigationMenu.Link>
          </NavigationMenu.Item>
        </NavigationMenu.List>
      </NavigationMenu.Root>
    </nav>

    <!-- Theme Toggle -->
    <Button variant="ghost" size="icon" onclick={toggleMode} aria-label="Toggle theme">
      <Sun
        class="
          size-5
          dark:hidden
        "
      />
      <Moon
        class="
          hidden size-5
          dark:block
        "
      />
    </Button>
  </header>

  <!-- Main Content -->
  <main class="flex flex-1 flex-col items-center justify-center p-12">
    <!-- Logo -->
    <div class="mb-8 flex items-center gap-3">
      <span
        class="
          font-display text-[32px] font-extrabold tracking-[-1px]
          text-foreground
        ">hister</span
      >
    </div>

    <!-- Search Container -->
    <div class="w-full max-w-200">
      <!-- Search Bar -->
      <div
        class="
          relative flex h-18 items-center gap-4 rounded-[36px] bg-[#F4F4F5] px-6
        "
      >
        <Search class="size-6 shrink-0 self-center text-[#A1A1AA]" />
        <Input
          type="text"
          bind:value={searchQuery}
          oninput={handleInput}
          onkeydown={handleKeyDown}
          placeholder="Search your browsing history..."
          class="
            h-full flex-1 self-center border-0 bg-transparent py-0! text-lg
            shadow-none
            placeholder:text-[#A1A1AA]
            focus-visible:ring-0 focus-visible:ring-offset-0
          "
        />
      </div>
    </div>

    <!-- Recent Searches -->
    {#if recentSearches.length > 0}
      <div class="mt-8 w-full max-w-200">
        <p
          class="
            mb-3 text-center text-xs font-medium tracking-[0.5px] text-[#A1A1AA]
            uppercase
          "
        >
          Recent Searches
        </p>
        <div class="flex flex-wrap justify-center gap-2">
          {#each recentSearches as term (term)}
            <Button variant="outline" size="sm" onclick={() => setSearchQuery(term)}>
              <Clock class="size-3.5" />
              {term}
            </Button>
          {/each}
        </div>
      </div>
    {/if}
  </main>

  <!-- Stats Bar -->
  <footer
    class="
      flex h-12 items-center justify-center gap-6 border-t border-border
      bg-[#FAFAFA] px-12 text-sm
    "
  >
    {#if loading}
      <span class="text-muted-foreground">Loading stats…</span>
    {:else}
      <div class="flex items-center gap-2">
        <span class="font-display text-base font-semibold text-foreground"
          >{formatNumber(stats.pagesIndexed)}</span
        >
        <span class="text-[#71717A]">pages indexed</span>
      </div>
      <Separator orientation="vertical" class="h-4" />
      <div class="flex items-center gap-2">
        <span class="font-display text-base font-semibold text-foreground"
          >{formatNumber(stats.domains)}</span
        >
        <span class="text-[#71717A]">domains</span>
      </div>
      <Separator orientation="vertical" class="h-4" />
      <span class="font-medium text-[#A1A1AA]">{stats.dateRange}</span>
    {/if}
  </footer>
</div>

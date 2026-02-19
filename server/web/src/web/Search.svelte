<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import {
    Search,
    X,
    Clock,
    ChevronDown,
    Download,
    Moon,
    Sun,
    Trash2,
    Calendar,
  } from 'lucide-svelte';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import * as Popover from '$lib/components/ui/popover';
  import * as ContextMenu from '$lib/components/ui/context-menu';
  import { Calendar as CalendarComponent } from '$lib/components/ui/calendar';
  import { toggleMode } from 'mode-watcher';
  import {
    search,
    deleteDocument,
    updateHistory,
    deleteHistoryItem,
    fetchStats,
    type SearchResults,
    type Stats,
  } from '$lib/api';
  import Button from '$lib/components/ui/button/button.svelte';
  import { Input } from '$lib/components/ui/input';
  import { Badge } from '$lib/components/ui/badge';
  import * as Item from '$lib/components/ui/item';
  import { Separator } from '$lib/components/ui/separator';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import * as Empty from '$lib/components/ui/empty';
  import { Search as SearchIcon } from 'lucide-svelte';
  import * as NavigationMenu from '$lib/components/ui/navigation-menu';
  import { CalendarDate, getLocalTimeZone, type DateValue } from '@internationalized/date';
  import { cn } from '$lib/utils';
  import { buttonVariants } from '$lib/components/ui/button';

  interface Props {
    query?: string;
  }

  let { query = '' }: Props = $props();
  let searchQuery = $state('');
  let results = $state<SearchResults | null>(null);
  let loading = $state(false);
  let dateFrom = $state<DateValue | undefined>();
  let dateTo = $state<DateValue | undefined>();
  let lastTrackedQuery = $state('');
  let hoveredDate = $state<DateValue | undefined>();
  let datePickerOpen = $state(false);
  let calendarMinDate = $state<DateValue | undefined>();
  let calendarMaxDate = $state<DateValue | undefined>();
  let stats = $state<Stats | null>(null);

  let sortBy = $state('relevance');

  const sortOptions = [
    { value: 'relevance', label: 'Relevance' },
    { value: 'date', label: 'Date' },
    { value: 'domain', label: 'Domain' },
  ];

  let searchTimeout: ReturnType<typeof setTimeout>;

  function performSearch() {
    if (!searchQuery.trim()) {
      results = null;
      return;
    }

    loading = true;

    const queryObj = {
      text: searchQuery,
      highlight: 'HTML',
      limit: 100,
      sort: sortBy === 'relevance' ? '' : sortBy,
      date_from: dateFrom
        ? Math.floor(dateFrom.toDate(getLocalTimeZone()).getTime() / 1000)
        : undefined,
      date_to: dateTo ? Math.floor(dateTo.toDate(getLocalTimeZone()).getTime() / 1000) : undefined,
    };

    search(queryObj, (searchResults) => {
      results = searchResults;
      loading = false;
    });
  }

  async function trackSearchInteraction() {
    if (!searchQuery.trim() || lastTrackedQuery === searchQuery) {
      return;
    }
    try {
      const searchUrl = `/search?q=${encodeURIComponent(searchQuery)}`;
      await updateHistory(searchQuery, searchUrl, `Search: ${searchQuery}`);
      lastTrackedQuery = searchQuery;
    } catch (err) {
      console.error('Failed to track search:', err);
    }
  }

  // Parse date from YYYY-MM-DD string to CalendarDate
  function parseDateFromString(dateStr: string): DateValue | undefined {
    if (!dateStr) return undefined;
    const [year, month, day] = dateStr.split('-').map(Number);
    if (year && month && day) {
      return new CalendarDate(year, month, day);
    }
    return undefined;
  }

  // Format CalendarDate to YYYY-MM-DD string
  function formatDateToString(date: DateValue | undefined): string {
    if (!date) return '';
    return `${date.year}-${String(date.month).padStart(2, '0')}-${String(date.day).padStart(2, '0')}`;
  }

  // Update URL with current search params
  function updateUrl() {
    const url = new URL(window.location.href);
    if (searchQuery) {
      url.searchParams.set('q', searchQuery);
    } else {
      url.searchParams.delete('q');
    }

    if (dateFrom) {
      url.searchParams.set('date_from', formatDateToString(dateFrom));
    } else {
      url.searchParams.delete('date_from');
    }

    if (dateTo) {
      url.searchParams.set('date_to', formatDateToString(dateTo));
    } else {
      url.searchParams.delete('date_to');
    }

    window.history.replaceState({}, '', url.toString());
  }

  // Sync searchQuery with query prop on mount only
  $effect.pre(() => {
    searchQuery = query;
  });

  // Perform search when query prop changes
  $effect(() => {
    const currentQuery = searchQuery;
    clearTimeout(searchTimeout);
    if (currentQuery) {
      searchTimeout = setTimeout(() => {
        performSearch();
        updateUrl();
      }, 300);
    }
  });

  onMount(async () => {
    // Parse date params from URL first
    const urlParams = new URLSearchParams(window.location.search);
    const dateFromParam = urlParams.get('date_from');
    const dateToParam = urlParams.get('date_to');

    if (dateFromParam) {
      dateFrom = parseDateFromString(dateFromParam);
    }
    if (dateToParam) {
      dateTo = parseDateFromString(dateToParam);
    }

    // Use the cached stats from background loader
    try {
      stats = await fetchStats();
      if (stats.minDate && stats.maxDate) {
        const minDateObj = new Date(stats.minDate * 1000);
        const maxDateObj = new Date(stats.maxDate * 1000);
        calendarMinDate = new CalendarDate(
          minDateObj.getFullYear(),
          minDateObj.getMonth() + 1,
          minDateObj.getDate()
        );
        calendarMaxDate = new CalendarDate(
          maxDateObj.getFullYear(),
          maxDateObj.getMonth() + 1,
          maxDateObj.getDate()
        );
      }
    } catch (err) {
      console.error('Failed to fetch stats:', err);
    }

    // Trigger search after dates are set
    if (query) {
      performSearch();
    }
  });

  onDestroy(() => {
    clearTimeout(searchTimeout);
  });

  function handleSearch(e: KeyboardEvent) {
    if (e.key === 'Enter' && searchQuery.trim()) {
      clearTimeout(searchTimeout);
      performSearch();
      updateUrl();
    }
  }

  function clearSearch() {
    searchQuery = '';
    results = null;
    window.location.href = '/';
  }

  function formatDate(timestamp: number): string {
    const date = new Date(timestamp * 1000);
    const now = new Date();
    const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24));

    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return `${diffDays} days ago`;
    if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`;
    return date.toLocaleDateString();
  }

  function isDateSelectedFrom(date: DateValue): boolean {
    return isDateRangeStart(date);
  }

  function isDateSelectedTo(date: DateValue): boolean {
    return isDateRangeEnd(date);
  }

  function formatCalendarDate(date: DateValue | undefined): string {
    if (!date) return 'Select date';
    return date.toDate(getLocalTimeZone()).toLocaleDateString();
  }

  function isDateDisabled(date: DateValue): boolean {
    if (calendarMinDate && date.compare(calendarMinDate) < 0) return true;
    if (calendarMaxDate && date.compare(calendarMaxDate) > 0) return true;
    return false;
  }

  function isDateRangeStart(date: DateValue): boolean {
    if (!dateFrom) return false;
    const end = dateTo || hoveredDate;
    if (!end) return false;
    const start = dateFrom.compare(end) <= 0 ? dateFrom : end;
    return date.compare(start) === 0;
  }

  function isDateRangeEnd(date: DateValue): boolean {
    if (!dateFrom) return false;
    const end = dateTo || hoveredDate;
    if (!end) return false;
    const stop = dateFrom.compare(end) <= 0 ? end : dateFrom;
    return date.compare(stop) === 0;
  }

  function handleDateClick(date: DateValue) {
    if (!dateFrom) {
      dateFrom = date;
    } else if (dateFrom && !dateTo) {
      // Don't allow selecting end date before start date
      if (date.compare(dateFrom) >= 0) {
        dateTo = date;
        datePickerOpen = false;
        trackSearchInteraction();
        updateUrl();
        // Trigger search when date range is complete
        if (searchQuery.trim()) {
          clearTimeout(searchTimeout);
          searchTimeout = setTimeout(performSearch, 100);
        }
      } else {
        // If clicked before start date, reset and start new range
        dateFrom = date;
        dateTo = undefined;
      }
    } else {
      dateFrom = date;
      dateTo = undefined;
    }
  }

  async function handleDeleteDocument(url: string, e: Event) {
    e.preventDefault();
    e.stopPropagation();

    if (!confirm('Are you sure you want to delete this document?')) {
      return;
    }

    try {
      await deleteDocument(url);
      if (results) {
        results.documents = results.documents.filter((d) => d.url !== url);
        results.total = Math.max(0, results.total - 1);
      }
    } catch (err) {
      console.error('Failed to delete document:', err);
      alert('Failed to delete document');
    }
  }

  async function handleDeleteHistoryItem(url: string, e: Event) {
    e.preventDefault();
    e.stopPropagation();

    try {
      await deleteHistoryItem(searchQuery, url);
      if (results && results.history) {
        results.history = results.history.filter((h) => h.url !== url);
      }
    } catch (err) {
      console.error('Failed to delete history item:', err);
      alert('Failed to delete history item');
    }
  }

  function exportResults(format: string) {
    if (!results) return;

    let content = '';
    let filename = `search-results-${new Date().toISOString().split('T')[0]}`;
    let mimeType = '';

    switch (format) {
      case 'json':
        content = JSON.stringify(results.documents, null, 2);
        filename += '.json';
        mimeType = 'application/json';
        break;
      case 'csv':
        content = 'Title,URL,Domain,Date\n';
        content += results.documents
          .map(
            (d) =>
              `"${d.title.replace(/"/g, '""')}","${d.url}","${d.domain}","${new Date(d.added * 1000).toISOString()}"`
          )
          .join('\n');
        filename += '.csv';
        mimeType = 'text/csv';
        break;
      case 'html':
        content = `<!DOCTYPE html>
<html>
<head><title>Search Results</title></head>
<body>
<h1>Search Results for "${searchQuery}"</h1>
<ul>
${results.documents.map((d) => `  <li><a href="${d.url}">${d.title}</a> - ${d.domain}</li>`).join('\n')}
</ul>
</body>
</html>`;
        filename += '.html';
        mimeType = 'text/html';
        break;
    }

    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }

  $effect(() => {
    if (sortBy) {
      clearTimeout(searchTimeout);
      if (searchQuery) {
        searchTimeout = setTimeout(() => {
          performSearch();
          updateUrl();
        }, 300);
      }
    }
  });

  // Watch for date changes and trigger search
  $effect(() => {
    const from = dateFrom;
    const to = dateTo;
    // Only trigger if we have a search query and dates are defined
    if (searchQuery.trim() && (from !== undefined || to !== undefined)) {
      clearTimeout(searchTimeout);
      searchTimeout = setTimeout(() => {
        performSearch();
        updateUrl();
      }, 300);
    }
  });
</script>

<svelte:head>
  <title>{searchQuery ? `${searchQuery} - Search` : 'Search'} - Hister</title>
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
  <main class="flex flex-col items-center px-12 py-10">
    <!-- Search Section -->
    <div class="w-full max-w-200">
      <!-- Search Bar -->
      <div
        class="
          flex h-15 items-center gap-4 rounded-2xl border border-[#E4E4E7]
          bg-white px-6
        "
      >
        <Search class="size-6 shrink-0 self-center text-[#A1A1AA]" />
        <Input
          type="text"
          bind:value={searchQuery}
          placeholder="Search your browsing history..."
          onkeydown={handleSearch}
          class="
            h-full flex-1 self-center border-0 bg-transparent py-0! text-lg
            focus-visible:ring-0 focus-visible:outline-none
          "
        />
        {#if searchQuery}
          <Button
            variant="ghost"
            size="icon-sm"
            onclick={clearSearch}
            aria-label="Clear search"
            class="self-center"
          >
            <X class="size-4" />
          </Button>
        {/if}
      </div>

      <!-- Filters Section-->
      <div class="mt-4 flex items-center gap-5 rounded-2xl bg-[#FAFAFA] p-4">
        <!-- Sort By -->
        <div class="flex items-center gap-2">
          <span class="text-sm whitespace-nowrap text-[#71717A]">Sort by:</span>
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              {#snippet child({ props })}
                <button
                  {...props}
                  class="
                    flex h-9 items-center gap-2 rounded-md border
                    border-[#E4E4E7] bg-white px-4 text-sm
                  "
                >
                  {sortOptions.find((o) => o.value === sortBy)?.label}
                  <ChevronDown class="size-3" />
                </button>
              {/snippet}
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="start">
              {#each sortOptions as option (option.value)}
                <DropdownMenu.Item
                  onclick={() => {
                    sortBy = option.value;
                    trackSearchInteraction();
                  }}
                >
                  {option.label}
                </DropdownMenu.Item>
              {/each}
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        </div>

        <!-- Date Range -->
        <div class="flex items-center gap-2">
          <span class="text-sm whitespace-nowrap text-[#71717A]">Date:</span>
          <Popover.Root bind:open={datePickerOpen}>
            <Popover.Trigger>
              {#snippet child({ props })}
                <Button
                  {...props}
                  variant="outline"
                  class="
                    h-9 w-64 justify-between border-[#E4E4E7] bg-white text-sm
                    font-normal
                  "
                >
                  {#if dateFrom && dateTo}
                    {formatCalendarDate(dateFrom)} - {formatCalendarDate(dateTo)}
                  {:else if dateFrom}
                    {formatCalendarDate(dateFrom)} - Select end date
                  {:else}
                    Select date range
                  {/if}
                  <Calendar class="size-3.5" />
                </Button>
              {/snippet}
            </Popover.Trigger>
            <Popover.Content class="w-auto p-0" align="start">
              <CalendarComponent
                type="single"
                bind:value={dateFrom}
                captionLayout="dropdown"
              >
                {#snippet day({ day: date, outsideMonth })}
                  <div
                    class="
                      flex size-(--cell-size) items-center justify-center
                      rounded-md
                    "
                  >
                    <button
                      type="button"
                      disabled={isDateDisabled(date)}
                      class={cn(
                        buttonVariants({ variant: 'ghost' }),
                        `
                          size-full flex-col items-center justify-center gap-1
                          rounded-md p-0 leading-none font-normal
                          whitespace-nowrap select-none
                        `,
                        isDateSelectedFrom(date) &&
                          `
                            rounded-full! bg-primary! text-primary-foreground!
                            hover:bg-primary!
                          `,
                        isDateSelectedTo(date) &&
                          `
                            rounded-full! bg-primary! text-primary-foreground!
                            hover:bg-primary!
                          `,
                        isDateDisabled(date) && 'cursor-not-allowed opacity-30',
                        outsideMonth && 'text-muted-foreground'
                      )}
                      onmouseenter={() => {
                        if (!isDateDisabled(date)) {
                          hoveredDate = date;
                        }
                      }}
                      onmouseleave={() => {
                        hoveredDate = undefined;
                      }}
                      onfocus={() => {
                        if (!isDateDisabled(date)) {
                          hoveredDate = date;
                        }
                      }}
                      onblur={() => {
                        hoveredDate = undefined;
                      }}
                      onclick={() => {
                        if (!isDateDisabled(date)) {
                          handleDateClick(date);
                        }
                      }}
                    >
                      <span class="text-xs opacity-70">{date.day}</span>
                    </button>
                  </div>
                {/snippet}
              </CalendarComponent>
            </Popover.Content>
          </Popover.Root>
        </div>

        <!-- Export Button -->
        <div class="ml-auto">
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              {#snippet child({ props })}
                <button
                  {...props}
                  class="
                    flex h-9 items-center gap-2 rounded-lg border
                    border-[#E4E4E7] bg-white px-4 text-sm
                  "
                >
                  <Download class="size-4" />
                  Export
                </button>
              {/snippet}
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="end">
              <DropdownMenu.Label>Export Format</DropdownMenu.Label>
              <DropdownMenu.Separator />
              <DropdownMenu.Item onclick={() => exportResults('json')}>
                Export as JSON
              </DropdownMenu.Item>
              <DropdownMenu.Item onclick={() => exportResults('csv')}>
                Export as CSV
              </DropdownMenu.Item>
              <DropdownMenu.Item onclick={() => exportResults('html')}>
                Export as HTML
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        </div>
      </div>

      <!-- Results Meta -->
      {#if results}
        <div class="mt-4 flex items-center gap-4 px-2">
          <span class="text-sm font-medium text-[#A1A1AA]">{results.total} results</span>
          <span class="text-sm text-[#A1A1AA]">{results.search_duration}</span>
        </div>
      {:else if loading}
        <div class="mt-6 w-full max-w-200 space-y-3">
          <div class="flex items-center gap-4">
            <Skeleton class="h-5 w-20" />
            <Skeleton class="h-5 w-24" />
          </div>
          <Skeleton class="h-28 w-full rounded-2xl" />
          <Skeleton class="h-28 w-full rounded-2xl" />
          <Skeleton class="h-28 w-full rounded-2xl" />
        </div>
      {/if}
    </div>

    <!-- Results List -->
    {#if results}
      <div class="mt-6 w-full max-w-200 space-y-3">
        <!-- History Results -->
        {#if results.history && results.history.length > 0}
          {#each results.history as historyItem (historyItem.url)}
            <ContextMenu.Root>
              <ContextMenu.Trigger>
                <Item.Root
                  class="
                    group border-[#E4E4E7] bg-white transition-shadow
                    hover:shadow-md
                  "
                >
                  {#snippet child({ props })}
                    <a {...props} href={historyItem.url} target="_blank" rel="noopener noreferrer">
                      <Item.Media variant="icon">
                        <Clock class="size-5 text-[#71717A]" />
                      </Item.Media>
                      <Item.Content>
                        <Item.Title class="text-base font-semibold">
                          {historyItem.title}
                        </Item.Title>
                        <Item.Description class="text-xs text-[#71717A]">
                          {historyItem.url}
                        </Item.Description>
                        <div
                          class="
                            mt-1 flex items-center gap-2 text-xs text-[#71717A]
                          "
                        >
                          <span class="font-medium">History</span>
                          <span>•</span>
                          <span>{historyItem.count} visits</span>
                        </div>
                      </Item.Content>
                      <Item.Actions>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          class="
                            opacity-0 transition-opacity
                            group-hover:opacity-100
                            hover:bg-destructive/10 hover:text-destructive
                          "
                          onclick={(e) => handleDeleteHistoryItem(historyItem.url, e)}
                          aria-label="Delete history"
                        >
                          <Trash2 class="size-4" />
                        </Button>
                      </Item.Actions>
                    </a>
                  {/snippet}
                </Item.Root>
              </ContextMenu.Trigger>
              <ContextMenu.Content>
                <ContextMenu.Item
                  class="cursor-pointer"
                  onclick={(e) => handleDeleteHistoryItem(historyItem.url, e)}
                >
                  <Trash2 class="mr-2 size-4" />
                  Delete from history
                </ContextMenu.Item>
              </ContextMenu.Content>
            </ContextMenu.Root>
          {/each}
          <Separator class="my-2" />
        {/if}

        <!-- Document Results -->
        {#each results.documents as result (result.url)}
          <Item.Root
            class="
              group border-[#E4E4E7] bg-white transition-shadow
              hover:shadow-md
            "
          >
            {#snippet child({ props })}
              <a {...props} href={result.url} target="_blank" rel="noopener noreferrer">
                <Item.Media>
                  {#if result.favicon}
                    <img
                      src={result.favicon}
                      alt=""
                      class="
                        size-10 rounded-lg bg-[#F4F4F5] object-contain p-1
                      "
                    />
                  {:else}
                    <div
                      class="
                        flex size-10 items-center justify-center rounded-lg
                        bg-[#F4F4F5]
                      "
                    >
                      <span class="text-xs font-bold text-[#71717A]">
                        {result.domain.slice(0, 2).toUpperCase()}
                      </span>
                    </div>
                  {/if}
                </Item.Media>
                <Item.Content>
                  <Item.Title class="text-base font-semibold">
                    {#if result.title}
                      {result.title}
                    {:else}
                      Untitled
                    {/if}
                  </Item.Title>
                  <Item.Description class="text-xs text-[#71717A]">
                    {result.url}
                  </Item.Description>
                  {#if result.text}
                    <Item.Description
                      class="
                        mt-1 line-clamp-2 text-sm text-secondary-foreground
                      "
                    >
                      {result.text}
                    </Item.Description>
                  {/if}
                  <div
                    class="mt-2 flex items-center gap-2 text-xs text-[#71717A]"
                  >
                    <Clock class="size-3.5" />
                    <span class="font-medium">{formatDate(result.added)}</span>
                    <span>•</span>
                    <span>{result.domain}</span>
                  </div>
                </Item.Content>
                <Item.Actions>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    class="
                      opacity-0 transition-opacity
                      group-hover:opacity-100
                      hover:bg-destructive/10 hover:text-destructive
                    "
                    onclick={(e) => handleDeleteDocument(result.url, e)}
                    aria-label="Delete document"
                  >
                    <Trash2 class="size-4" />
                  </Button>
                </Item.Actions>
              </a>
            {/snippet}
          </Item.Root>
        {/each}

        {#if results.documents.length === 0 && (!results.history || results.history.length === 0)}
          <div class="w-full max-w-200">
            <Empty.Root>
              <Empty.Header>
                <Empty.Media variant="icon">
                  <SearchIcon class="size-8 text-muted-foreground" />
                </Empty.Media>
                <Empty.Title>No Results Found</Empty.Title>
                <Empty.Description>
                  No results found for "{searchQuery}". Try a different search term or check your
                  spelling.
                </Empty.Description>
              </Empty.Header>
            </Empty.Root>
          </div>
        {/if}
      </div>
    {:else if !loading && searchQuery}
      <div class="mt-6 w-full max-w-200">
        <Empty.Root>
          <Empty.Header>
            <Empty.Media variant="icon">
              <SearchIcon class="size-8 text-muted-foreground" />
            </Empty.Media>
            <Empty.Title>Start Your Search</Empty.Title>
            <Empty.Description>
              Type a query above to search your browsing history
            </Empty.Description>
          </Empty.Header>
        </Empty.Root>
      </div>
    {/if}
  </main>
</div>

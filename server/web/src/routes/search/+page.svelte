<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Search, X, ChevronDown, Download, Trash2, Clock, Calendar as CalendarIcon } from 'lucide-svelte';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import * as Dialog from '$lib/components/ui/dialog';
  import * as Popover from '$lib/components/ui/popover';
  import * as ContextMenu from '$lib/components/ui/context-menu';
  import { Calendar } from '$lib/components/ui/calendar';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Card, CardContent } from '$lib/components/ui/card';
  import { Badge } from '$lib/components/ui/badge';
  import {
    search,
    deleteDocument,
    trackSearch,
    deleteHistoryItem,
    fetchStats,
    type SearchResults,
    type Stats,
  } from '$lib/api';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { getLocalTimeZone, CalendarDate, type DateValue } from '@internationalized/date';
  import { buttonVariants } from '$lib/components/ui/button';
  import { cn } from '$lib/utils';
  import { sanitizeHtml, stripHtml } from '$lib/sanitize';

  let searchQuery = $state('');
  let results = $state<SearchResults | null>(null);
  let loading = $state(false);
  let dateFrom = $state<DateValue | undefined>(undefined);
  let dateTo = $state<DateValue | undefined>(undefined);
  let calendarValue = $state<DateValue | undefined>(undefined);
  let sortBy = $state('relevance');
  let searchInput: HTMLInputElement | null = $state(null);
  let searchTimeout: ReturnType<typeof setTimeout>;
  let datePickerOpen = $state(false);
  let hoveredDate = $state<DateValue | undefined>();
  let minDate = $state<DateValue | undefined>();
  let maxDate = $state<DateValue | undefined>();
  let stats = $state<Stats | null>(null);

  // Delete dialog state
  let deleteDialogOpen = $state(false);
  let deleteUrl = $state<string | null>(null);
  let deleteTitle = $state<string | null>(null);

  const sortOptions = [
    { value: 'relevance', label: 'Relevance' },
    { value: 'date', label: 'Date' },
    { value: 'domain', label: 'Domain' },
  ];

  function performSearch() {
    if (!searchQuery.trim()) {
      results = null;
      return;
    }

    loading = true;

    trackSearch(searchQuery).catch((err) => {
      console.error('Failed to track search:', err);
    });

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

  // Only sync from URL on initial mount, not reactively
  onMount(async () => {
    const urlQuery = page.url.searchParams.get('q') || '';
    if (urlQuery && urlQuery !== searchQuery) {
      searchQuery = urlQuery;
      performSearch();
    }
    // Focus the search input when page mounts for seamless typing
    searchInput?.focus();

    // Fetch stats to get date range
    try {
      stats = await fetchStats();
      if (stats.minDate && stats.maxDate) {
        const minDateObj = new Date(stats.minDate * 1000);
        const maxDateObj = new Date(stats.maxDate * 1000);
        minDate = new CalendarDate(
          minDateObj.getFullYear(),
          minDateObj.getMonth() + 1,
          minDateObj.getDate()
        );
        maxDate = new CalendarDate(
          maxDateObj.getFullYear(),
          maxDateObj.getMonth() + 1,
          maxDateObj.getDate()
        );
      }
    } catch (err) {
      console.error('Failed to fetch stats:', err);
    }
  });

  onDestroy(() => {
    clearTimeout(searchTimeout);
  });

  function handleInput() {
    clearTimeout(searchTimeout);
    if (searchQuery.trim()) {
      goto(`/search?q=${encodeURIComponent(searchQuery)}`, { keepFocus: true, noScroll: true });
      searchTimeout = setTimeout(performSearch, 300);
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter' && searchQuery.trim()) {
      clearTimeout(searchTimeout);
      performSearch();
      goto(`/search?q=${encodeURIComponent(searchQuery)}`, { keepFocus: true, noScroll: true });
    }
  }

  function clearSearch() {
    searchQuery = '';
    results = null;
    goto('/', { keepFocus: true });
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

  function formatCalendarDate(date: DateValue | undefined): string {
    if (!date) return '';
    return date.toDate(getLocalTimeZone()).toLocaleDateString();
  }

  function isDateSelectedFrom(date: DateValue): boolean {
    if (!dateFrom) return false;
    const end = dateTo || hoveredDate;
    if (!end) return false;
    const start = dateFrom.compare(end) <= 0 ? dateFrom : end;
    const stop = dateFrom.compare(end) <= 0 ? end : dateFrom;
    const inRange = date.compare(start) >= 0 && date.compare(stop) <= 0;
    return inRange && !isDateSelectedTo(date);
  }

  function isDateSelectedTo(date: DateValue): boolean {
    return dateTo ? date.compare(dateTo) === 0 : false;
  }

  function isDateDisabled(date: DateValue): boolean {
    if (minDate && date.compare(minDate) < 0) return true;
    if (maxDate && date.compare(maxDate) > 0) return true;
    return false;
  }

  function handleDateClick(date: DateValue) {
    if (!dateFrom) {
      dateFrom = date;
    } else if (dateFrom && !dateTo) {
      // Don't allow selecting end date before start date
      if (date.compare(dateFrom) >= 0) {
        dateTo = date;
        datePickerOpen = false;
      } else {
        // If clicked before start date, reset and start new range
        dateFrom = date;
        dateTo = undefined;
      }
    } else {
      dateFrom = date;
      dateTo = undefined;
    }
    // Reset calendar value to prevent it from showing current day
    calendarValue = undefined;
  }

  function openDeleteDialog(url: string, title: string, e: Event) {
    e.preventDefault();
    e.stopPropagation();
    deleteUrl = url;
    deleteTitle = stripHtml(title);
    deleteDialogOpen = true;
  }

  async function handleDeleteDocument() {
    if (!deleteUrl) return;

    try {
      await deleteDocument(deleteUrl);
      if (results) {
        results.documents = results.documents.filter((d) => d.url !== deleteUrl);
        results.total = Math.max(0, results.total - 1);
      }
      deleteDialogOpen = false;
      deleteUrl = null;
      deleteTitle = null;
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

  // Watch sortBy changes and re-search when sorting changes
  $effect(() => {
    if (sortBy) {
      clearTimeout(searchTimeout);
      if (searchQuery) {
        searchTimeout = setTimeout(performSearch, 300);
      }
    }
  });

  // Watch date changes and re-search when dates change
  $effect(() => {
    if (dateFrom !== undefined || dateTo !== undefined) {
      clearTimeout(searchTimeout);
      if (searchQuery) {
        searchTimeout = setTimeout(performSearch, 300);
      }
    }
  });
</script>

<svelte:head>
  <title>{searchQuery ? `${searchQuery} - Search` : 'Search'} - Hister</title>
</svelte:head>

<main class="flex flex-col items-center px-12 py-10">
  <div class="w-full max-w-200">
    <Card
      class="
        flex h-15 flex-row items-center gap-4 border border-border bg-background
        px-6 shadow-none
      "
    >
      <Search class="size-6 shrink-0 self-center text-muted-foreground" />
      <Input
        type="text"
        bind:value={searchQuery}
        bind:ref={searchInput}
        placeholder="Search your browsing history..."
        oninput={handleInput}
        onkeydown={handleKeyDown}
        class="
          h-full flex-1 self-center border-0! bg-transparent! py-0! text-lg
          shadow-none!
          focus-visible:ring-0! focus-visible:ring-offset-0!
        "
      />
      {#if searchQuery}
        <Button
          variant="ghost"
          size="icon"
          onclick={clearSearch}
          class="size-8 self-center"
          aria-label="Clear search"
        >
          <X class="size-4" />
        </Button>
      {/if}
    </Card>

    <div class="mt-4 flex items-center gap-5 rounded-2xl bg-muted p-4">
      <div class="flex items-center gap-2">
        <span class="text-sm whitespace-nowrap text-muted-foreground">Sort by:</span>
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            {#snippet child({ props })}
              <Button
                {...props}
                variant="outline"
                class="h-9 cursor-pointer gap-2"
              >
                {sortOptions.find((o) => o.value === sortBy)?.label}
                <ChevronDown class="size-3" />
              </Button>
            {/snippet}
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="start">
            {#each sortOptions as option (option.value)}
              <DropdownMenu.Item
                onclick={() => (sortBy = option.value)}
                class="cursor-pointer"
              >
                {option.label}
              </DropdownMenu.Item>
            {/each}
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>

      <div class="flex items-center gap-2">
        <span class="text-sm whitespace-nowrap text-muted-foreground">Date:</span>
        <Popover.Root bind:open={datePickerOpen}>
          <Popover.Trigger>
            {#snippet child({ props })}
              <Button
                {...props}
                variant="outline"
                class="h-9 cursor-pointer justify-between gap-2"
              >
                {#if dateFrom && dateTo}
                  {formatCalendarDate(dateFrom)} - {formatCalendarDate(dateTo)}
                {:else if dateFrom}
                  {formatCalendarDate(dateFrom)} - Select end date
                {:else}
                  Select date range
                {/if}
                <CalendarIcon class="size-4 text-muted-foreground" />
              </Button>
            {/snippet}
          </Popover.Trigger>
          <Popover.Content class="w-auto p-0" align="start">
            <Calendar
              type="single"
              bind:value={calendarValue}
              captionLayout="dropdown"
            >
              {#snippet day({ day: date, outsideMonth })}
                <button
                  type="button"
                  disabled={isDateDisabled(date)}
                  class={cn(
                    buttonVariants({ variant: 'ghost' }),
                    `
                      flex size-(--cell-size) cursor-pointer flex-col
                      items-center justify-center gap-1 rounded-md p-0
                      leading-none font-normal whitespace-nowrap select-none
                    `,
                    (isDateSelectedTo(date) || isDateSelectedFrom(date)) &&
                      `
                        bg-primary text-primary-foreground
                        dark:hover:bg-accent/50
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
              {/snippet}
            </Calendar>
          </Popover.Content>
        </Popover.Root>
      </div>

      <div class="ml-auto">
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            {#snippet child({ props })}
              <Button
                {...props}
                variant="outline"
                class="h-9 cursor-pointer gap-2"
              >
                <Download class="size-4" />
                Export
              </Button>
            {/snippet}
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="end">
            <DropdownMenu.Label>Export Format</DropdownMenu.Label>
            <DropdownMenu.Separator />
            <DropdownMenu.Item class="cursor-pointer" onclick={() => exportResults('json')}>
              Export as JSON
            </DropdownMenu.Item>
            <DropdownMenu.Item class="cursor-pointer" onclick={() => exportResults('csv')}>
              Export as CSV
            </DropdownMenu.Item>
            <DropdownMenu.Item class="cursor-pointer" onclick={() => exportResults('html')}>
              Export as HTML
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>
    </div>

    {#if results}
      <div class="mt-4 flex items-center gap-4 px-2">
        <span class="text-sm font-medium text-muted-foreground">{results.total} results</span>
        <span class="text-sm text-muted-foreground">{results.search_duration}</span>
      </div>
    {:else if loading}
      <div class="mt-4 flex items-center gap-4 px-2">
        <span class="text-sm text-muted-foreground">Searching…</span>
      </div>
    {/if}
  </div>

  {#if results}
    <div class="mt-6 w-full max-w-200 space-y-3">
      {#if results.history && results.history.length > 0}
        {#each results.history as historyItem (historyItem.url)}
          <ContextMenu.Root>
            <ContextMenu.Trigger>
              <Card
                class="
                  group transition-shadow
                  hover:shadow-md
                "
              >
                <CardContent class="p-0">
                  <a
                    href={historyItem.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    class="flex flex-col gap-3 p-5"
                  >
                    <div class="flex items-start gap-4">
                      <div
                        class="
                          flex size-10 shrink-0 items-center justify-center
                          rounded-lg bg-muted
                        "
                      >
                        <Clock class="size-5 text-muted-foreground" />
                      </div>
                      <div class="flex flex-1 flex-col gap-1">
                        <div class="flex items-start justify-between gap-2">
                          <h2 class="text-base font-semibold text-foreground">
                            {historyItem.title}
                          </h2>
                          <Button
                            variant="ghost"
                            size="icon"
                            onclick={(e) => handleDeleteHistoryItem(historyItem.url, e)}
                            class="
                              size-8 opacity-0
                              group-hover:opacity-100
                            "
                          >
                            <Trash2 class="size-4" />
                          </Button>
                        </div>
                        <p class="text-xs break-all text-muted-foreground">
                          {historyItem.url}
                        </p>
                      </div>
                    </div>
                    <div
                      class="
                        flex items-center gap-2 text-xs text-muted-foreground
                      "
                    >
                      <Badge variant="secondary">History</Badge>
                      <span>•</span>
                      <span>{historyItem.count} visits</span>
                    </div>
                  </a>
                </CardContent>
              </Card>
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
      {/if}

      {#each results.documents as result (result.url)}
        <Card
          class="
            group transition-shadow
            hover:shadow-md
          "
        >
          <CardContent class="p-0">
            <a
              href={result.url}
              target="_blank"
              rel="noopener noreferrer"
              class="flex flex-col gap-3 p-5"
            >
              <div class="flex items-start gap-4">
                {#if result.favicon}
                  <img
                    src={result.favicon}
                    alt=""
                    class="
                      size-10 shrink-0 rounded-lg bg-muted object-contain p-1
                    "
                  />
                {:else}
                  <div
                    class="
                      flex size-10 shrink-0 items-center justify-center
                      rounded-lg bg-muted
                    "
                  >
                    <span class="text-xs font-bold text-muted-foreground">
                      {result.domain.slice(0, 2).toUpperCase()}
                    </span>
                  </div>
                {/if}
                <div class="flex flex-1 flex-col gap-1">
                  <div class="flex items-start justify-between gap-2">
                    <h2 class="text-base font-semibold text-foreground">
                      {#if result.title}
                        {@html sanitizeHtml(result.title)}
                      {:else}
                        Untitled
                      {/if}
                    </h2>
                    <Button
                      variant="ghost"
                      size="icon"
                      onclick={(e) => openDeleteDialog(result.url, result.title || 'Untitled', e)}
                      class="
                        size-8 opacity-0
                        group-hover:opacity-100
                      "
                    >
                      <Trash2 class="size-4" />
                    </Button>
                  </div>
                  <p class="text-xs break-all text-muted-foreground">
                    {result.url}
                  </p>
                  {#if result.text}
                    <p
                      class="
                        mt-1 line-clamp-2 text-sm text-secondary-foreground
                      "
                    >
                      {@html sanitizeHtml(result.text)}
                    </p>
                  {/if}
                </div>
              </div>
              <div
                class="flex items-center gap-2 text-xs text-muted-foreground"
              >
                <Clock class="size-3.5" />
                <span class="font-medium">{formatDate(result.added)}</span>
                <span>•</span>
                <span>{result.domain}</span>
              </div>
            </a>
          </CardContent>
        </Card>
      {/each}

      {#if results.documents.length === 0 && (!results.history || results.history.length === 0)}
        <Card class="py-12">
          <CardContent
            class="flex flex-col items-center justify-center text-center"
          >
            <p class="text-muted-foreground">No results found for "{searchQuery}"</p>
            <p class="mt-2 text-sm text-muted-foreground">
              Try a different search term or check your spelling.
            </p>
          </CardContent>
        </Card>
      {/if}
    </div>
  {:else if !loading && searchQuery}
    <div class="mt-6 w-full max-w-200">
      <Card class="py-12">
        <CardContent
          class="flex flex-col items-center justify-center text-center"
        >
          <p class="text-muted-foreground">Start typing to search your browsing history</p>
        </CardContent>
      </Card>
    </div>
  {/if}
</main>

<!-- Delete Confirmation Dialog -->
<Dialog.Root bind:open={deleteDialogOpen}>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Delete Entry</Dialog.Title>
      <Dialog.Description>
        Are you sure you want to delete "{deleteTitle}"? This action cannot be undone.
      </Dialog.Description>
    </Dialog.Header>
    <Dialog.Footer>
      <Button variant="outline" onclick={() => (deleteDialogOpen = false)}>Cancel</Button>
      <Button variant="destructive" onclick={handleDeleteDocument}>Delete</Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

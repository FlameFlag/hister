<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import SearchInput from '$lib/features/search/components/SearchInput.svelte';
  import SearchResults from '$lib/features/search/components/SearchResults.svelte';
  import DateRangePicker from '$lib/features/search/components/DateRangePicker.svelte';
  import SortDropdown from '$lib/features/search/components/SortDropdown.svelte';
  import ExportMenu from '$lib/features/search/components/ExportMenu.svelte';
  import DeleteDialog from '$lib/shared/components/DeleteDialog.svelte';
  import { search as searchWebSocket } from '$lib/api/websocket';
  import { deleteDocument, trackSearch, deleteHistoryItem, fetchStats, type Stats } from '$lib/api';
  import { getLocalTimeZone, type DateValue, CalendarDate } from '@internationalized/date';
  import { fromUnixTime, getUnixTime } from 'date-fns';
  import { stripHtml } from '$lib/sanitize';
  import { debounce } from '@solid-primitives/scheduled';

  let searchQuery = $state('');
  let results = $state<import('$lib/api').SearchResults | null>();
  let loading = $state(false);
  let dateFrom = $state<DateValue | undefined>();
  let dateTo = $state<DateValue | undefined>();
  let sortBy = $state('relevance');
  let stats = $state<Stats | null>();

  let deleteDialogOpen = $state(false);
  let deleteUrl = $state<string | null>();
  let deleteTitle = $state<string | null>();

  let minDate = $state<DateValue | undefined>();
  let maxDate = $state<DateValue | undefined>();

  const sortOptions = [
    { value: 'relevance', label: 'Relevance' },
    { value: 'date', label: 'Date' },
    { value: 'domain', label: 'Domain' },
  ];

  const debouncedPerformSearch = debounce(() => {
    performSearch();
  }, 300);

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
      date_from: dateFrom ? getUnixTime(dateFrom.toDate(getLocalTimeZone())) : undefined,
      date_to: dateTo ? getUnixTime(dateTo.toDate(getLocalTimeZone())) : undefined,
    };

    searchWebSocket(queryObj, (searchResults) => {
      results = searchResults;
      loading = false;
    });
  }

  onMount(async () => {
    const urlQuery = page.url.searchParams.get('q') || '';
    if (urlQuery && urlQuery !== searchQuery) {
      searchQuery = urlQuery;
    }

    const dateFromParam = page.url.searchParams.get('date_from');
    const dateToParam = page.url.searchParams.get('date_to');

    if (dateFromParam) {
      const [year, month, day] = dateFromParam.split('-').map(Number);
      dateFrom = new CalendarDate(year, month, day);
    }

    if (dateToParam) {
      const [year, month, day] = dateToParam.split('-').map(Number);
      dateTo = new CalendarDate(year, month, day);
    }

    try {
      stats = await fetchStats();
      if (stats.minDate && stats.maxDate) {
        const minDateObj = fromUnixTime(stats.minDate);
        const maxDateObj = fromUnixTime(stats.maxDate);
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

    if (searchQuery) {
      performSearch();
    }
  });

  function updateURL() {
    const params = new URLSearchParams();
    if (searchQuery.trim()) {
      params.set('q', searchQuery);
    }
    if (dateFrom) {
      params.set(
        'date_from',
        `${dateFrom.year}-${String(dateFrom.month).padStart(2, '0')}-${String(dateFrom.day).padStart(2, '0')}`
      );
    }
    if (dateTo) {
      params.set(
        'date_to',
        `${dateTo.year}-${String(dateTo.month).padStart(2, '0')}-${String(dateTo.day).padStart(2, '0')}`
      );
    }
    const queryString = params.toString();
    goto(`/search${queryString ? '?' + queryString : ''}`, {
      keepFocus: true,
      noScroll: true,
    });
  }

  function handleInput() {
    if (searchQuery.trim()) {
      updateURL();
      debouncedPerformSearch();
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter' && searchQuery.trim()) {
      debouncedPerformSearch();
      updateURL();
    }
  }

  function handleDateFromChange(value: DateValue | undefined) {
    dateFrom = value;
    updateURL();
  }

  function handleDateToChange(value: DateValue | undefined) {
    dateTo = value;
    updateURL();
  }

  function handleOpenChange(value: boolean) {
    deleteDialogOpen = value;
  }

  function clearSearch() {
    searchQuery = '';
    dateFrom = undefined;
    dateTo = undefined;
    results = null;
    goto('/', { keepFocus: true });
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
              `"${d.title.replace(/"/g, '""')}","${d.url}","${d.domain}","${fromUnixTime(d.added).toISOString()}"`
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
      debouncedPerformSearch();
    }
  });

  $effect(() => {
    if (dateFrom || dateTo) {
      debouncedPerformSearch();
    }
  });
</script>

<svelte:head>
  <title>{searchQuery ? `${searchQuery} - Search` : 'Search'} - Hister</title>
</svelte:head>

<main class="flex flex-col items-center px-12 py-10">
  <div class="w-full max-w-200">
    <SearchInput
      bind:value={searchQuery}
      autoFocus={true}
      placeholder="Search your browsing history..."
      oninput={handleInput}
      onkeydown={handleKeyDown}
      onClear={clearSearch}
    />

    <section class="mt-4 flex items-center gap-5 rounded-2xl bg-muted p-4">
      <span class="flex items-center gap-2 text-sm whitespace-nowrap text-muted-foreground">
        Sort by:
        <SortDropdown value={sortBy} options={sortOptions} onChange={(v) => (sortBy = v)} />
      </span>

      <span class="flex items-center gap-2 text-sm whitespace-nowrap text-muted-foreground">
        Date:
        <DateRangePicker
          {dateFrom}
          onDateFromChange={handleDateFromChange}
          {dateTo}
          onDateToChange={handleDateToChange}
          {minDate}
          {maxDate}
        />
      </span>

      <div class="ml-auto">
        <ExportMenu onExport={exportResults} />
      </div>
    </section>

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
    <div class="mt-6 w-full max-w-200">
      <SearchResults
        {results}
        onDeleteDocument={openDeleteDialog}
        onDeleteHistoryItem={handleDeleteHistoryItem}
      />
    </div>
  {:else if !loading && searchQuery}
    <div class="mt-6 w-full max-w-200">
      <div class="rounded-xl border border-border p-12 text-center">
        <p class="text-muted-foreground">Start typing to search your browsing history</p>
      </div>
    </div>
  {/if}
</main>

<DeleteDialog
  open={deleteDialogOpen}
  onOpenChange={handleOpenChange}
  title="Delete Entry"
  description={`Are you sure you want to delete "${deleteTitle}"? This action cannot be undone.`}
  onConfirm={handleDeleteDocument}
  confirmText="Delete"
  variant="destructive"
/>

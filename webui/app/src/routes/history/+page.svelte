<script lang="ts">
  import { onMount } from 'svelte';
  import { fetchConfig, apiFetch } from '$lib/api';
  import { Button } from '@hister/components/ui/button';
  import { Badge } from '@hister/components/ui/badge';
  import { Separator } from '@hister/components/ui/separator';
  import { Search, Clock, Trash2 } from 'lucide-svelte';
  import StatusMessage from '$lib/components/StatusMessage.svelte';

  interface HistoryItem {
    query: string;
    url: string;
    title: string;
    updated_at: string;
  }

  let items: HistoryItem[] = $state([]);
  let loading = $state(true);
  let error = $state('');
  let filter = $state('');
  let activeGroup = $state('');
  let filterByDate = $state('');

  const groupColors = [
    'hister-indigo', 'hister-coral', 'hister-teal', 'hister-amber',
    'hister-rose', 'hister-cyan', 'hister-lime'
  ];

  function getColorVar(color: string): string {
    return `var(--${color})`;
  }

  function formatDateLabel(dateStr: string): string {
    const date = new Date(dateStr);
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);
    const itemDate = new Date(date.getFullYear(), date.getMonth(), date.getDate());

    if (itemDate.getTime() === today.getTime()) return 'Today';
    if (itemDate.getTime() === yesterday.getTime()) return 'Yesterday';
    return itemDate.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' });
  }

  function getDateKey(dateStr: string): string {
    const date = new Date(dateStr);
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
  }

  const filteredItems = $derived.by(() => {
    let result = items;
    if (filter) {
      const f = filter.toLowerCase();
      result = result.filter(item =>
        item.query.toLowerCase().includes(f) ||
        item.title.toLowerCase().includes(f) ||
        item.url.toLowerCase().includes(f)
      );
    }
    if (filterByDate) {
      result = result.filter(item => item.updated_at && getDateKey(item.updated_at) === filterByDate);
    }
    return result;
  });

  const allGroups = $derived.by(() => {
    const g: { key: string; label: string; items: HistoryItem[] }[] = [];
    const seen = new Map<string, number>();
    let baseItems = items;
    if (filter) {
      const f = filter.toLowerCase();
      baseItems = baseItems.filter(item =>
        item.query.toLowerCase().includes(f) ||
        item.title.toLowerCase().includes(f) ||
        item.url.toLowerCase().includes(f)
      );
    }
    for (const item of baseItems) {
      const key = item.updated_at ? getDateKey(item.updated_at) : 'unknown';
      const label = item.updated_at ? formatDateLabel(item.updated_at) : 'Unknown';
      if (seen.has(key)) {
        g[seen.get(key)!].items.push(item);
      } else {
        seen.set(key, g.length);
        g.push({ key, label, items: [item] });
      }
    }
    return g;
  });

  const groups = $derived.by(() => {
    const g: { key: string; label: string; items: HistoryItem[] }[] = [];
    const seen = new Map<string, number>();
    for (const item of filteredItems) {
      const key = item.updated_at ? getDateKey(item.updated_at) : 'unknown';
      const label = item.updated_at ? formatDateLabel(item.updated_at) : 'Unknown';
      if (seen.has(key)) {
        g[seen.get(key)!].items.push(item);
      } else {
        seen.set(key, g.length);
        g.push({ key, label, items: [item] });
      }
    }
    return g;
  });

  function getGroupColor(idx: number): string {
    return groupColors[idx % groupColors.length];
  }

  function scrollToGroup(key: string) {
    activeGroup = key;
    filterByDate = key;
  }

  function showAll() {
    filterByDate = '';
    activeGroup = groups.length > 0 ? groups[0].key : '';
  }

  async function deleteHistoryItem(item: HistoryItem) {
    try {
      await apiFetch('/history', {
        method: 'POST',
        headers: { 'Content-type': 'application/json; charset=UTF-8' },
        body: JSON.stringify({ url: item.url, title: item.title, query: item.query, delete: true })
      });
      items = items.filter(i => i.url !== item.url || i.query !== item.query);
    } catch (e) {
      error = String(e);
    }
  }

  async function deleteAllHistory() {
    if (!confirm('Delete all history? This cannot be undone.')) return;
    try {
      for (const item of items) {
        await apiFetch('/history', {
          method: 'POST',
          headers: { 'Content-type': 'application/json; charset=UTF-8' },
          body: JSON.stringify({ url: item.url, title: item.title, query: item.query, delete: true })
        });
      }
      items = [];
    } catch (e) {
      error = String(e);
    }
  }

  onMount(async () => {
    try {
      await fetchConfig();
      const res = await apiFetch('/history', {
        headers: { 'Accept': 'application/json' }
      });
      if (!res.ok) throw new Error('Failed to load history');
      items = await res.json();
    } catch (e) {
      error = String(e);
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>Hister - History</title>
</svelte:head>

<div class="flex items-center justify-between px-6 py-3 bg-card-surface border-b-[2px] border-border-brand-muted shrink-0">
  <h1 class="font-outfit text-lg font-extrabold text-text-brand">Search History</h1>
  <div class="flex items-center gap-3">
    <div class="flex items-center gap-2 h-8 px-3 border-[2px] border-border-brand-muted bg-page-bg">
      <Search class="size-3.5 text-text-brand-muted shrink-0" />
      <input
        type="text"
        bind:value={filter}
        placeholder="Filter..."
        class="w-40 h-full bg-transparent font-inter text-xs font-medium text-text-brand placeholder:text-text-brand-muted outline-none border-0"
      />
    </div>
    {#if items.length > 0}
      <Button
        variant="outline"
        size="sm"
        class="border-[2px] border-hister-rose text-hister-rose hover:bg-hister-rose/10 font-inter text-xs font-semibold h-8 gap-1.5"
        onclick={deleteAllHistory}
      >
        <Trash2 class="size-3.5" />
        Delete All
      </Button>
    {/if}
  </div>
</div>

{#if loading}
  <StatusMessage type="loading" message="Loading history..." />
{:else if error}
  <div class="px-6 py-4">
    <StatusMessage type="error" message={error} />
  </div>
{:else if filteredItems.length === 0}
  <StatusMessage type="empty" message={filter ? 'No matching entries' : 'No history yet'} />
{:else}
  <div class="flex flex-1 min-h-0">
    <div class="w-[180px] shrink-0 border-r-[2px] border-border-brand-muted pt-5 pr-3 overflow-y-auto">
      <div class="space-y-1">
        <span class="font-space text-[11px] font-bold tracking-[2px] text-text-brand-muted px-2.5 flex items-center gap-1.5">
          <Clock class="size-3" />
          TIMELINE
        </span>
        <Separator class="bg-border-brand-muted" />

        <button
          class="flex items-center gap-2 w-full py-2 px-2.5 text-left border-0 cursor-pointer"
          style={!filterByDate ? 'background-color: var(--hister-indigo); color: white;' : ''}
          class:bg-transparent={!!filterByDate}
          class:hover:bg-muted-surface={!!filterByDate}
          onclick={showAll}
        >
          <span class="font-inter text-[13px] font-semibold" class:text-text-brand-secondary={!!filterByDate}>
            Show All
          </span>
          <Badge
            variant="secondary"
            class="ml-auto shrink-0 text-[10px] px-1.5 py-0 h-4 border-0 {filterByDate ? 'bg-muted-surface text-text-brand-muted' : ''}"
            style={!filterByDate ? 'background-color: rgba(255,255,255,0.2); color: white;' : ''}
          >
            {filteredItems.length}
          </Badge>
        </button>

        <Separator class="bg-border-brand-muted" />

        {#each allGroups as group, i}
          {@const color = getGroupColor(i)}
          {@const isActive = filterByDate === group.key}
          <button
            class="flex items-center gap-2 w-full py-2 px-2.5 text-left border-0 cursor-pointer"
            style={isActive ? `background-color: ${getColorVar(color)}; color: white;` : ''}
            class:bg-transparent={!isActive}
            class:hover:bg-muted-surface={!isActive}
            onclick={() => scrollToGroup(group.key)}
          >
            <span
              class="w-2 h-2 shrink-0 rounded-full"
              style={isActive ? 'background-color: white;' : `background-color: ${getColorVar(color)};`}
            ></span>
            <span
              class="font-inter text-[13px] truncate"
              class:font-semibold={isActive}
              class:font-medium={!isActive}
              class:text-text-brand-secondary={!isActive}
            >
              {group.label}
            </span>
            <Badge
              variant="secondary"
              class="ml-auto shrink-0 text-[10px] px-1.5 py-0 h-4 border-0 {!isActive ? 'bg-muted-surface text-text-brand-muted' : ''}"
              style={isActive ? 'background-color: rgba(255,255,255,0.2); color: white;' : ''}
            >
              {group.items.length}
            </Badge>
          </button>
        {/each}
      </div>
    </div>

    <div class="flex-1 min-w-0 overflow-y-auto px-6 py-5 space-y-6">
      {#each groups as group, gi}
        {@const color = getGroupColor(gi)}
        <div id="group-{encodeURIComponent(group.key)}" class="space-y-2">
          <span class="font-outfit text-sm font-bold" style="color: {getColorVar(color)};">{group.label}</span>
          <div class="h-0.5" style="background-color: {getColorVar(color)};"></div>

          <div class="space-y-0">
            {#each group.items as item, ii}
              {@const itemColor = getGroupColor(gi + ii)}
              <div
                class="flex items-center gap-3 py-2.5 px-3.5 bg-card-surface border-b-[2px] border-b-border-brand-muted overflow-hidden"
                style="border-left: 3px solid {getColorVar(itemColor)};"
              >
                <div class="flex-1 min-w-0 space-y-0.5">
                  <a
                    href={item.url}
                    class="font-outfit text-[15px] font-bold hover:underline block truncate no-underline"
                    style="color: {getColorVar(itemColor)};"
                    target="_blank"
                    rel="noopener"
                  >
                    {item.title || item.url}
                  </a>
                  <span class="font-fira text-[11px] text-text-brand-muted block truncate" title={item.url}>{item.url}</span>
                </div>
                <div class="flex items-center gap-1 shrink-0">
                  <Button
                    variant="ghost"
                    size="sm"
                    class="text-xs font-inter text-text-brand-muted shrink-0 hover:text-hister-indigo gap-1.5 h-7 px-2 no-underline"
                    href="/?q={encodeURIComponent(item.query)}"
                  >
                    <Search class="size-3" />
                    Search
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    class="text-text-brand-muted hover:text-hister-rose shrink-0 size-7"
                    onclick={() => deleteHistoryItem(item)}
                  >
                    <Trash2 class="size-3.5" />
                  </Button>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  </div>
{/if}

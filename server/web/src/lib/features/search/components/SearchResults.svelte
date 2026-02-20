<script lang="ts">
  import ResultCard from './ResultCard.svelte';
  import HistoryCard from './HistoryCard.svelte';
  import EmptyState from '$lib/shared/components/EmptyState.svelte';
  import type { SearchResults } from '$lib/api';

  let {
    results,
    onDeleteDocument,
    onDeleteHistoryItem,
  }: {
    results: SearchResults;
    onDeleteDocument?: (url: string, title: string, e: Event) => void;
    onDeleteHistoryItem?: (url: string, e: Event) => void;
  } = $props();
</script>

<div class="space-y-3">
  {#if results.history && results.history.length > 0}
    {#each results.history as historyItem (historyItem.url)}
      <HistoryCard {historyItem} onDelete={onDeleteHistoryItem} />
    {/each}
  {/if}

  {#each results.documents as result (result.url)}
    <ResultCard {result} showHistoryBadge={false} onDelete={onDeleteDocument} />
  {/each}

  {#if results.documents.length === 0 && (!results.history || results.history.length === 0)}
    <EmptyState
      title="No results found"
      description="Try a different search term or check your spelling."
    />
  {/if}
</div>

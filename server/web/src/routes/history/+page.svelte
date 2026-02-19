<script lang="ts">
  import { onMount } from 'svelte';
  import { Trash2 } from 'lucide-svelte';
  import { fetchHistory, deleteHistoryItem, type HistoryItem } from '$lib/api';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent } from '$lib/components/ui/card';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import { Badge } from '$lib/components/ui/badge';

  let history = $state<HistoryItem[]>([]);
  let loading = $state(true);
  let deleteDialogOpen = $state(false);
  let deleteQuery = $state<string | null>(null);
  let deleteUrl = $state<string | null>(null);
  let deleteTitle = $state<string | null>(null);

  onMount(async () => {
    try {
      const allHistory = await fetchHistory(100);
      // Only show opened links (filter out search URL entries)
      history = allHistory.filter(
        (h: HistoryItem) => h.query !== h.url && !h.url.startsWith('/search')
      );
    } catch (err) {
      console.error('Failed to load history:', err);
    } finally {
      loading = false;
    }
  });

  function getDomain(url: string): string {
    try {
      return new URL(url).hostname;
    } catch {
      return url;
    }
  }

  function openDeleteDialog(item: HistoryItem) {
    deleteQuery = item.query;
    deleteUrl = item.url;
    deleteTitle = item.title || item.url;
    deleteDialogOpen = true;
  }

  async function handleDelete() {
    if (!deleteQuery || !deleteUrl) return;

    try {
      await deleteHistoryItem(deleteQuery, deleteUrl);
      history = history.filter((h) => h.url !== deleteUrl);
      deleteDialogOpen = false;
      deleteQuery = null;
      deleteUrl = null;
      deleteTitle = null;
    } catch (err) {
      console.error('Failed to delete:', err);
      alert('Failed to delete entry');
    }
  }
</script>

<svelte:head>
  <title>History - Hister</title>
</svelte:head>

<main class="flex flex-col items-center px-16 py-12">
  <div class="w-full max-w-300">
    <h1 class="mb-8 text-[32px] font-extrabold tracking-[-1px] text-foreground">
      Browsing History
    </h1>

    {#if loading}
      <div class="space-y-3">
        {#each Array(5)}
          <Card>
            <CardContent class="p-5">
              <div class="flex items-start gap-4">
                <Skeleton class="size-10 shrink-0 rounded-lg" />
                <div class="flex flex-1 flex-col gap-2">
                  <Skeleton class="h-5 w-3/4" />
                  <Skeleton class="h-4 w-1/2" />
                </div>
              </div>
            </CardContent>
          </Card>
        {/each}
      </div>
    {:else if history.length === 0}
      <Card class="py-12">
        <CardContent
          class="flex flex-col items-center justify-center text-center"
        >
          <p class="text-muted-foreground">No browsing history found</p>
          <p class="mt-2 text-sm text-muted-foreground">Start browsing to build your history.</p>
        </CardContent>
      </Card>
    {:else}
      <div class="space-y-3">
        {#each history as item (item.url + '|' + item.query)}
          <Card
            class="
              group transition-shadow
              hover:shadow-md
            "
          >
            <CardContent class="p-0">
              <a
                href={item.url}
                target="_blank"
                rel="noopener noreferrer"
                class="flex flex-col gap-3 p-5"
              >
                <div class="flex items-start gap-4">
                  {#if item.favicon}
                    <img
                      src={item.favicon}
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
                        {getDomain(item.url).slice(0, 2).toUpperCase()}
                      </span>
                    </div>
                  {/if}
                  <div class="flex flex-1 flex-col gap-1">
                    <div class="flex items-start justify-between gap-2">
                      <h2 class="text-base font-semibold text-foreground">
                        {item.title || item.query || 'Untitled'}
                      </h2>
                      <button
                        type="button"
                        onclick={(e) => {
                          e.preventDefault();
                          openDeleteDialog(item);
                        }}
                        class="
                          flex size-8 cursor-pointer items-center justify-center
                          rounded-lg text-muted-foreground opacity-0
                          transition-opacity
                          group-hover:opacity-100
                          hover:bg-destructive/10 hover:text-destructive
                        "
                        aria-label="Delete entry"
                      >
                        <Trash2 class="size-4" />
                      </button>
                    </div>
                    <p class="text-xs break-all text-muted-foreground">
                      {item.url}
                    </p>
                  </div>
                </div>
                <div
                  class="flex items-center gap-2 text-xs text-muted-foreground"
                >
                  <Badge variant="secondary">History</Badge>
                  <span>•</span>
                  <span>{getDomain(item.url)}</span>
                </div>
              </a>
            </CardContent>
          </Card>
        {/each}
      </div>
    {/if}
  </div>
</main>

<Dialog.Root bind:open={deleteDialogOpen}>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Delete Entry</Dialog.Title>
      <Dialog.Description>
        Are you sure you want to delete "{deleteTitle}"? This action cannot be undone.
      </Dialog.Description>
    </Dialog.Header>
    <Dialog.Footer>
      <Button variant="outline" class="cursor-pointer" onclick={() => (deleteDialogOpen = false)}>
        Cancel
      </Button>
      <Button variant="destructive" class="cursor-pointer" onclick={handleDelete}>Delete</Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

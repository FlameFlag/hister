<script lang="ts">
  import { Clock, Trash2 } from 'lucide-svelte';
  import * as ContextMenu from '$lib/components/ui/context-menu';
  import { Button } from '$lib/components/ui/button';
  import { Badge } from '$lib/components/ui/badge';
  import { Card, CardContent } from '$lib/components/ui/card';
  import type { URLCount } from '$lib/api';

  let {
    historyItem,
    onDelete,
  }: {
    historyItem: URLCount;
    onDelete?: (url: string, e: Event) => void;
  } = $props();
</script>

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
                flex size-10 shrink-0 items-center justify-center rounded-lg
                bg-muted
              "
            >
              <Clock class="size-5 text-muted-foreground" />
            </div>
            <div class="flex flex-1 flex-col gap-1">
              <div class="flex items-start justify-between gap-2">
                <h2 class="text-base font-semibold text-foreground">
                  {historyItem.title}
                </h2>
                {#if onDelete}
                  <Button
                    variant="ghost"
                    size="icon"
                    onclick={(e) => onDelete(historyItem.url, e)}
                    class="
                      size-8 opacity-0
                      group-hover:opacity-100
                    "
                  >
                    <Trash2 class="size-4" />
                  </Button>
                {/if}
              </div>
              <p class="text-xs break-all text-muted-foreground">
                {historyItem.url}
              </p>
            </div>
          </div>
          <span class="flex items-center gap-2 text-xs text-muted-foreground">
            <Badge variant="secondary">History</Badge>
            <span>•</span>
            <span>{historyItem.count} visits</span>
          </span>
        </a>
      </CardContent>
    </Card>
  </ContextMenu.Trigger>
  <ContextMenu.Content>
    <ContextMenu.Item class="cursor-pointer" onclick={(e) => onDelete?.(historyItem.url, e)}>
      <Trash2 class="mr-2 size-4" />
      Delete from history
    </ContextMenu.Item>
  </ContextMenu.Content>
</ContextMenu.Root>

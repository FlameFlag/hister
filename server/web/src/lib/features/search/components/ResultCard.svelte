<script lang="ts">
  import { Clock, Trash2 } from 'lucide-svelte';
  import { sanitizeHtml } from '$lib/sanitize';
  import { Button } from '$lib/components/ui/button';
  import { Badge } from '$lib/components/ui/badge';
  import { Card, CardContent } from '$lib/components/ui/card';
  import type { Document } from '$lib/api';
  import { fromUnixTime, formatDistanceToNow } from 'date-fns';

  let {
    result,
    onDelete,
    showHistoryBadge = false,
  }: {
    result: Document;
    onDelete?: (url: string, title: string, e: Event) => void;
    showHistoryBadge?: boolean;
  } = $props();

  function getDomainInitials(domain: string): string {
    return domain.slice(0, 2).toUpperCase();
  }

  function formatTimestamp(timestamp: number): string {
    return formatDistanceToNow(fromUnixTime(timestamp), { addSuffix: true });
  }
</script>

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
      class="
      flex flex-col gap-3 p-5
    "
    >
      <div class="flex items-start gap-4">
        {#if result.favicon}
          <img
            src={result.favicon}
            alt=""
            class="size-10 shrink-0 rounded-lg bg-muted object-contain p-1"
          />
        {:else}
          <div
            class="
              flex size-10 shrink-0 items-center justify-center rounded-lg
              bg-muted
            "
          >
            <span class="text-xs font-bold text-muted-foreground">
              {getDomainInitials(result.domain)}
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
            {#if onDelete}
              <Button
                variant="ghost"
                size="icon"
                onclick={(e) => onDelete(result.url, result.title || 'Untitled', e)}
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
            {result.url}
          </p>

          {#if result.text}
            <p class="mt-1 line-clamp-2 text-sm text-secondary-foreground">
              {@html sanitizeHtml(result.text)}
            </p>
          {/if}
        </div>
      </div>

      <span class="flex items-center gap-2 text-xs text-muted-foreground">
        {#if showHistoryBadge}
          <Badge variant="secondary">History</Badge>
          <span>•</span>
        {/if}
        <Clock class="size-3.5" />
        <span class="font-medium">{formatTimestamp(result.added)}</span>
        <span>•</span>
        <span>{result.domain}</span>
      </span>
    </a>
  </CardContent>
</Card>

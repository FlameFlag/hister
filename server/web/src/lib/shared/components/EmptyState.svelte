<script lang="ts">
  import type { Snippet } from 'svelte';
  import { Card, CardContent } from '$lib/components/ui/card';

  let {
    icon,
    title,
    description,
    action,
    children,
  }: {
    icon?: Snippet;
    title: string;
    description?: string;
    action?: Snippet;
    children?: Snippet;
  } = $props();
</script>

<Card class="py-12">
  <CardContent class="flex flex-col items-center justify-center text-center">
    {#if icon}
      {@render icon()}
    {/if}

    <h3 class="mt-4 text-lg font-semibold">{title}</h3>

    {#if description}
      <p class="mt-2 text-sm text-muted-foreground">
        {description}
      </p>
    {/if}

    {#if action}
      <div class="mt-4">
        {@render action()}
      </div>
    {/if}

    {#if children}
      <div class="mt-4">
        {@render children()}
      </div>
    {/if}
  </CardContent>
</Card>

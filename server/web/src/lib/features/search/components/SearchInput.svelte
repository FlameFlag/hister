<script lang="ts">
  import { onMount } from 'svelte';
  import { Search, X } from 'lucide-svelte';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Card } from '$lib/components/ui/card';
  import type { Snippet } from 'svelte';

  let {
    value = $bindable(''),
    onValueChange,
    placeholder = 'Search your browsing history...',
    onClear,
    oninput,
    onkeydown,
    children,
    autoFocus = false,
  }: {
    value?: string;
    onValueChange?: (value: string) => void;
    placeholder?: string;
    onClear?: () => void;
    oninput?: () => void;
    onkeydown?: (e: KeyboardEvent) => void;
    children?: Snippet;
    autoFocus?: boolean;
  } = $props();

  let inputRef = $state<HTMLInputElement | null>(null);

  $effect(() => {
    onValueChange?.(value);
  });

  onMount(() => {
    if (autoFocus && inputRef) {
      inputRef.focus();
    }
  });
</script>

<Card
  class="
    flex h-15 flex-row items-center gap-4 border border-border bg-background
    px-6 shadow-none
  "
>
  <Search class="size-6 shrink-0 self-center text-muted-foreground" />
  <Input
    type="text"
    bind:value
    bind:ref={inputRef}
    {placeholder}
    {oninput}
    {onkeydown}
    class="
      h-full flex-1 self-center border-0! bg-transparent! py-0! text-lg
      shadow-none!
      focus-visible:ring-0! focus-visible:ring-offset-0!
    "
  />
  {#if value && onClear}
    <Button
      variant="ghost"
      size="icon"
      onclick={onClear}
      class="size-8 self-center"
      aria-label="Clear search"
    >
      <X class="size-4" />
    </Button>
  {/if}
  {#if children}
    {@render children()}
  {/if}
</Card>

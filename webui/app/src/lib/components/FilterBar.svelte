<script lang="ts">
    import { Search } from 'lucide-svelte';

    let {
      value = $bindable(''),
      placeholder = 'Filter...'
    }: {
      value?: string;
      placeholder?: string;
    } = $props();
</script>

<div class="flex items-center gap-3 h-11 px-4 bg-card-surface border-b-[2px] border-border-brand-muted">
    <Search class="size-4 text-text-brand-muted shrink-0" />
    <input
      type="text"
      bind:value
      {placeholder}
      class="flex-1 h-full bg-transparent font-inter text-sm font-medium text-text-brand placeholder:text-text-brand-muted outline-none border-0"
    />
</div>

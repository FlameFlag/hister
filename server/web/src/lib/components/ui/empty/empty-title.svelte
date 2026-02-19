<script lang="ts">
  import type { HTMLAttributes } from 'svelte/elements';
  import { cn, type WithElementRef } from '$lib/utils.js';

  let {
    ref = $bindable(null),
    class: className,
    children,
    ...restProps
  }: WithElementRef<HTMLAttributes<HTMLHeadingElement>> = $props();
</script>

<h3
  bind:this={ref}
  data-slot="empty-title"
  class={cn('text-lg font-semibold', className)}
  {...restProps}
>
  {@render children?.()}
</h3>

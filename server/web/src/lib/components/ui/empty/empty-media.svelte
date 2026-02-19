<script lang="ts">
  import type { HTMLAttributes } from 'svelte/elements';
  import { cn, type WithElementRef } from '$lib/utils.js';

  let {
    ref = $bindable(null),
    class: className,
    children,
    variant = 'icon',
    ...restProps
  }: WithElementRef<
    HTMLAttributes<HTMLDivElement> & {
      variant?: 'icon' | 'default';
    }
  > = $props();
</script>

<div
  bind:this={ref}
  data-slot="empty-media"
  class={cn(
    'flex size-16 shrink-0 items-center justify-center',
    {
      'rounded-lg bg-muted': variant === 'icon',
      'rounded-full': variant === 'default',
    },
    className
  )}
  {...restProps}
>
  {@render children?.()}
</div>

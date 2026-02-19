<script lang="ts">
  import { cn } from '$lib/utils';
  import type { Snippet } from 'svelte';

  type Variant = 'default' | 'secondary' | 'outline';

  interface Props {
    variant?: Variant;
    children?: Snippet;
    class?: string;
    onclick?: () => void;
  }

  let { variant = 'default', children, class: className, onclick }: Props = $props();

  const baseStyles =
    'inline-flex items-center rounded-full px-4 py-2 text-[13px] font-medium transition-colors';

  const variants: Record<Variant, string> = {
    default: 'bg-primary text-primary-foreground',
    secondary: 'border border-border text-secondary-foreground hover:bg-secondary',
    outline: 'border border-border text-muted-foreground hover:bg-secondary',
  };
</script>

<button type="button" class={cn(baseStyles, variants[variant], className)} {onclick}>
  {#if children}
    {@render children()}
  {/if}
</button>

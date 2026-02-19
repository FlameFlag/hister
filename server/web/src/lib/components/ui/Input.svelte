<script lang="ts">
  import { cn } from '$lib/utils';

  interface Props {
    class?: string;
    type?: string;
    placeholder?: string;
    value?: string;
    id?: string;
    required?: boolean;
    disabled?: boolean;
    onkeydown?: (e: KeyboardEvent) => void;
    oninput?: (e: Event) => void;
    autofocus?: boolean;
  }

  let {
    class: className,
    type = 'text',
    placeholder,
    value = $bindable(''),
    id,
    required,
    disabled,
    onkeydown,
    oninput,
    autofocus,
  }: Props = $props();
</script>

<!-- svelte-ignore a11y_autofocus -->
<input
  {id}
  {type}
  {placeholder}
  {required}
  {disabled}
  bind:value
  {onkeydown}
  {oninput}
  {autofocus}
  class={cn(
    `
      flex w-full rounded-md border-0 bg-transparent px-3 py-2 text-sm
      text-foreground
      placeholder:text-muted-foreground
      focus-visible:ring-0 focus-visible:outline-none
      disabled:cursor-not-allowed disabled:opacity-50
    `,
    className
  )}
/>

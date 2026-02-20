<script lang="ts">
  import * as Dialog from '$lib/components/ui/dialog';
  import { Button } from '$lib/components/ui/button';

  let {
    open,
    onOpenChange,
    title,
    description,
    onConfirm,
    confirmText = 'Delete',
    cancelText = 'Cancel',
    variant = 'destructive',
  }: {
    open: boolean;
    onOpenChange?: (value: boolean) => void;
    title: string;
    description: string;
    onConfirm: () => void | Promise<void>;
    confirmText?: string;
    cancelText?: string;
    variant?: 'destructive' | 'default';
  } = $props();

  let loading = $state(false);
  let error = $state('');

  async function handleConfirm() {
    try {
      loading = true;
      error = '';
      await onConfirm();
    } catch (err) {
      error = err instanceof Error ? err.message : 'An error occurred';
      console.error('Delete failed:', err);
    } finally {
      loading = false;
    }
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>{title}</Dialog.Title>
      <Dialog.Description>
        {description}
      </Dialog.Description>
    </Dialog.Header>

    {#if error}
      <p class="mb-4 text-sm text-destructive">{error}</p>
    {/if}

    <Dialog.Footer>
      <Button
        variant="outline"
        onclick={() => onOpenChange?.(false)}
        disabled={loading}
        class="cursor-pointer"
      >
        {cancelText}
      </Button>
      <Button
        {variant}
        onclick={handleConfirm}
        disabled={loading}
        class="
        cursor-pointer
      "
      >
        {#if loading}
          <span class="animate-pulse">Deleting...</span>
        {:else}
          {confirmText}
        {/if}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

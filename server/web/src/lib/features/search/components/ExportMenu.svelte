<script lang="ts">
  import { Download } from 'lucide-svelte';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import { Button } from '$lib/components/ui/button';

  let {
    onExport,
  }: {
    onExport: (format: string) => void;
  } = $props();

  const exportFormats = [
    { value: 'json', label: 'Export as JSON' },
    { value: 'csv', label: 'Export as CSV' },
    { value: 'html', label: 'Export as HTML' },
  ];
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <Button {...props} variant="outline" class="h-9 cursor-pointer gap-2">
        <Download class="size-4" />
        Export
      </Button>
    {/snippet}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="end">
    <DropdownMenu.Label>Export Format</DropdownMenu.Label>
    <DropdownMenu.Separator />
    {#each exportFormats as format (format.value)}
      <DropdownMenu.Item class="cursor-pointer" onclick={() => onExport(format.value)}>
        {format.label}
      </DropdownMenu.Item>
    {/each}
  </DropdownMenu.Content>
</DropdownMenu.Root>

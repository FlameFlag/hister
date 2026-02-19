<script lang="ts">
  import { Plus, Loader, Check, Link, Type, FileText } from 'lucide-svelte';
  import { addEntry } from '$lib/api';
  import { Card, CardContent } from '$lib/components/ui/card';
  import { Input } from '$lib/components/ui/input';
  import { Textarea } from '$lib/components/ui/textarea';
  import { Label } from '$lib/components/ui/label';
  import { Button } from '$lib/components/ui/button';
  import { Alert, AlertDescription } from '$lib/components/ui/alert';

  let url = $state('');
  let title = $state('');
  let text = $state('');
  let submitting = $state(false);
  let success = $state(false);
  let error = $state<string | null>(null);

  async function handleSubmit(e: Event) {
    e.preventDefault();

    if (!url.trim()) return;

    submitting = true;
    error = null;
    success = false;

    try {
      await addEntry({
        url: url.trim(),
        title: title.trim(),
        text: text.trim(),
      });

      success = true;

      setTimeout(() => {
        url = '';
        title = '';
        text = '';
        success = false;
      }, 2000);
    } catch (err) {
      console.error('Failed to add entry:', err);
      error = err instanceof Error ? err.message : 'Failed to add entry';
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>Add History Entry - Hister</title>
</svelte:head>

<main class="flex flex-1 flex-col items-center px-16 py-12">
  <h1
    class="
      mb-10 font-display text-[32px] font-bold tracking-tight text-foreground
    "
  >
    Add History Entry
  </h1>

  {#if success}
    <Alert
      class="
        mb-6 w-full max-w-150 border-green-500/50 bg-green-500/10 text-green-600
      "
    >
      <Check class="size-4" />
      <AlertDescription>Entry added successfully!</AlertDescription>
    </Alert>
  {/if}

  {#if error}
    <Alert variant="destructive" class="mb-6 w-full max-w-150">
      <AlertDescription>{error}</AlertDescription>
    </Alert>
  {/if}

  <Card class="w-full max-w-150">
    <CardContent class="pt-6">
      <form onsubmit={handleSubmit} class="flex flex-col gap-6">
        <div class="flex flex-col gap-2">
          <div class="flex items-center gap-2">
            <Link class="size-4 text-muted-foreground" />
            <Label for="url">URL</Label>
            <span class="text-destructive">*</span>
          </div>
          <Input
            id="url"
            type="url"
            placeholder="https://example.com/page"
            bind:value={url}
            required
            disabled={submitting}
            class="h-14 bg-background"
          />
        </div>

        <div class="flex flex-col gap-2">
          <div class="flex items-center gap-2">
            <Type class="size-4 text-muted-foreground" />
            <Label for="title">Title</Label>
          </div>
          <Input
            id="title"
            type="text"
            placeholder="Page title"
            bind:value={title}
            disabled={submitting}
            class="h-14 bg-background"
          />
        </div>

        <div class="flex flex-col gap-2">
          <div class="flex items-center gap-2">
            <FileText class="size-4 text-muted-foreground" />
            <Label for="text">Text</Label>
          </div>
          <Textarea
            id="text"
            placeholder="Additional text or notes..."
            bind:value={text}
            rows={4}
            disabled={submitting}
            class="resize-none"
          />
        </div>

        <Button
          type="submit"
          disabled={submitting || !url.trim()}
          class="mt-2 h-12 w-full gap-2"
          size="lg"
        >
          {#if submitting}
            <Loader class="size-5 animate-spin" />
            Adding Entry…
          {:else}
            <Plus class="size-5" />
            Add
          {/if}
        </Button>
      </form>
    </CardContent>
  </Card>
</main>

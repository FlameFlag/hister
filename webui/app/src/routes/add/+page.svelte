<script lang="ts">
  import { onMount } from 'svelte';
  import { fetchConfig, apiFetch } from '$lib/api';
  import { Input } from '@hister/components/ui/input';
  import { Label } from '@hister/components/ui/label';
  import { Save, AlertCircle, CheckCircle } from 'lucide-svelte';

  let url = $state('');
  let title = $state('');
  let text = $state('');
  let message = $state('');
  let isError = $state(false);
  let submitting = $state(false);

  onMount(async () => {
    await fetchConfig();
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (submitting) return;
    submitting = true;
    message = '';
    try {
      const res = await apiFetch('/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url, title, text })
      });
      if (res.status === 201) {
        message = 'Document added successfully.';
        isError = false;
        url = '';
        title = '';
        text = '';
      } else if (res.status === 406) {
        message = 'URL skipped (matches skip rules or is a local URL).';
        isError = false;
      } else {
        message = 'Failed to add document.';
        isError = true;
      }
    } catch (err) {
      message = String(err);
      isError = true;
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>Hister - Add</title>
</svelte:head>

<div class="flex-1 flex items-start justify-center pt-12 px-6 overflow-y-auto">
  <div class="w-full max-w-[640px] bg-card-surface border-[3px] border-hister-indigo shadow-[6px_6px_0px_var(--hister-indigo)] overflow-hidden">
    <div class="flex items-center justify-between px-7 py-5 bg-hister-indigo">
      <h1 class="font-outfit font-black text-[22px] text-white">Add Entry</h1>
      <span class="font-inter text-[13px] font-medium text-white/70">Manually index a page</span>
    </div>

    <div class="p-7 space-y-6">
      {#if message}
        <div class="flex items-center gap-2 px-4 py-3 border-[2px] font-inter text-sm {isError ? 'border-hister-rose bg-hister-rose/10 text-hister-rose' : 'border-hister-teal bg-hister-teal/10 text-hister-teal'}">
          {#if isError}
            <AlertCircle class="size-4 shrink-0" />
          {:else}
            <CheckCircle class="size-4 shrink-0" />
          {/if}
          {message}
        </div>
      {/if}

      <form onsubmit={handleSubmit} class="space-y-6">
        <div class="space-y-2">
          <Label class="font-outfit text-sm font-bold text-text-brand">URL</Label>
          <Input
            type="text"
            bind:value={url}
            placeholder="https://..."
            required
            class="w-full h-12 px-4 bg-page-bg border-[3px] border-hister-indigo font-fira text-sm text-text-brand placeholder:text-text-brand-muted shadow-none focus-visible:ring-0 focus-visible:border-hister-coral transition-colors"
          />
        </div>

        <div class="space-y-2">
          <Label class="font-outfit text-sm font-bold text-text-brand">Title</Label>
          <Input
            type="text"
            bind:value={title}
            placeholder="Page title..."
            class="w-full h-12 px-4 bg-page-bg border-[3px] border-hister-indigo font-inter text-sm text-text-brand placeholder:text-text-brand-muted shadow-none focus-visible:ring-0 focus-visible:border-hister-coral transition-colors"
          />
        </div>

        <div class="space-y-2">
          <Label class="font-outfit text-sm font-bold text-text-brand">Content</Label>
          <textarea
            bind:value={text}
            placeholder="Paste or type page content..."
            class="w-full min-h-[180px] p-4 bg-page-bg border-[3px] border-hister-indigo font-inter text-sm text-text-brand placeholder:text-text-brand-muted outline-none focus:border-hister-coral transition-colors resize-y"
          ></textarea>
        </div>

        <button
          type="submit"
          disabled={submitting}
          class="w-full h-[52px] flex items-center justify-center gap-2.5 bg-hister-coral shadow-[4px_4px_0px_var(--hister-coral)] text-white font-outfit text-base font-extrabold tracking-[1px] border-0 hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <Save class="size-5 shrink-0" />
          <span>{submitting ? 'Saving...' : 'Save Entry'}</span>
        </button>
      </form>
    </div>
  </div>
</div>

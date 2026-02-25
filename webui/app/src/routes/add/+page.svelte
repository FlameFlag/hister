<script lang="ts">
  import { onMount } from 'svelte';
  import { fetchConfig, apiFetch } from '$lib/api';

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

<div class="container">
  {#if message}
    <p class:success={!isError} class:error={isError}>{message}</p>
  {/if}
  <form onsubmit={handleSubmit}>
    <input type="text" placeholder="URL..." bind:value={url} class="full-width" required /><br />
    <input type="text" placeholder="Title..." bind:value={title} class="full-width" /><br />
    <textarea placeholder="Text..." bind:value={text} class="full-width"></textarea>
    <input type="submit" value={submitting ? 'Adding...' : 'Add'} disabled={submitting} />
  </form>
</div>

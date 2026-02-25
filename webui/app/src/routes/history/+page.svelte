<script lang="ts">
  import { onMount } from 'svelte';
  import { fetchConfig, apiFetch } from '$lib/api';

  interface HistoryItem {
    query: string;
    url: string;
    title: string;
  }

  let items: HistoryItem[] = $state([]);
  let loading = $state(true);
  let error = $state('');

  onMount(async () => {
    try {
      await fetchConfig();
      const res = await apiFetch('/history', {
        headers: { 'Accept': 'application/json' }
      });
      if (!res.ok) throw new Error('Failed to load history');
      items = await res.json();
    } catch (e) {
      error = String(e);
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>Hister - History</title>
</svelte:head>

<div class="container full-width">
  <h1>History</h1>
  {#if loading}
    <p>Loading...</p>
  {:else if error}
    <p class="error">{error}</p>
  {:else}
    <table>
      <thead>
        <tr><th>Query</th><th>Result</th></tr>
      </thead>
      <tbody>
        {#each items as item}
          <tr>
            <td><a href="/?q={encodeURIComponent(item.query)}"><span class="success">{item.query}</span></a></td>
            <td><a href={item.url}>{item.title || item.url}</a></td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

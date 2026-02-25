<script lang="ts">
  import { onMount } from 'svelte';
  import { fetchConfig, apiFetch } from '$lib/api';

  interface RulesData {
    skip: string[];
    priority: string[];
    aliases: Record<string, string>;
  }

  let rules: RulesData = $state({ skip: [], priority: [], aliases: {} });
  let loading = $state(true);
  let saving = $state(false);
  let message = $state('');
  let isError = $state(false);
  let newAliasKeyword = $state('');
  let newAliasValue = $state('');

  onMount(async () => {
    await fetchConfig();
    await loadRules();
  });

  async function loadRules() {
    loading = true;
    try {
      const res = await apiFetch('/rules', {
        headers: { 'Accept': 'application/json' }
      });
      if (!res.ok) throw new Error('Failed to load rules');
      rules = await res.json();
    } catch (e) {
      message = String(e);
      isError = true;
    } finally {
      loading = false;
    }
  }

  async function saveRules(e: Event) {
    e.preventDefault();
    if (saving) return;
    saving = true;
    message = '';
    try {
      const formData = new URLSearchParams();
      formData.set('skip', rules.skip.join('\n'));
      formData.set('priority', rules.priority.join('\n'));
      const res = await apiFetch('/rules', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formData.toString()
      });
      if (!res.ok) throw new Error('Failed to save rules');
      message = 'Rules saved';
      isError = false;
      await loadRules();
    } catch (e) {
      message = String(e);
      isError = true;
    } finally {
      saving = false;
    }
  }

  async function deleteAlias(keyword: string) {
    const formData = new URLSearchParams({ alias: keyword });
    const res = await apiFetch('/delete_alias', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: formData.toString()
    });
    if (res.ok) {
      await loadRules();
    }
  }

  async function addAlias(e: Event) {
    e.preventDefault();
    if (!newAliasKeyword || !newAliasValue) return;
    const formData = new URLSearchParams({
      'alias-keyword': newAliasKeyword,
      'alias-value': newAliasValue
    });
    const res = await apiFetch('/add_alias', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: formData.toString()
    });
    if (res.ok) {
      newAliasKeyword = '';
      newAliasValue = '';
      await loadRules();
    }
  }
</script>

<svelte:head>
  <title>Hister - Rules</title>
</svelte:head>

<div class="container full-width">
  {#if message}
    <div class="container box" class:success={!isError} class:error={isError}>
      <div class="header">{message}</div>
    </div>
  {/if}

  {#if loading}
    <p>Loading...</p>
  {:else}
    <form onsubmit={saveRules}>
      <h2>Skip Rules</h2>
      <p>Define regexps to forbid indexing matching URLs</p>
      <textarea
        placeholder="Text..."
        name="skip"
        class="full-width"
        bind:value={rules.skip}
        oninput={(e) => { rules.skip = (e.target as HTMLTextAreaElement).value.split('\n').filter(Boolean); }}
      >{rules.skip.join('\n')}</textarea>
      <br />
      <h2>Priority Rules</h2>
      <p>Define regexps to prioritize matching URLs</p>
      <textarea
        placeholder="Text..."
        name="priority"
        class="full-width"
        bind:value={rules.priority}
        oninput={(e) => { rules.priority = (e.target as HTMLTextAreaElement).value.split('\n').filter(Boolean); }}
      >{rules.priority.join('\n')}</textarea>
      <br />
      <input type="submit" value={saving ? 'Saving...' : 'Save'} disabled={saving} class="mt-1" />
    </form>

    <h2>Search Keyword Aliases</h2>
    <p>Define aliases to simplify queries. Alias strings in queries are automatically replaced with the provided value.</p>

    {#if Object.keys(rules.aliases).length > 0}
      <table class="mv-1">
        <thead>
          <tr><th>Keyword</th><th>Value</th><th>Delete</th></tr>
        </thead>
        <tbody>
          {#each Object.entries(rules.aliases) as [k, v]}
            <tr>
              <td>{k}</td>
              <td>{v}</td>
              <td>
                <button onclick={() => deleteAlias(k)}>Delete</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {:else}
      <h3>There are no aliases</h3>
    {/if}

    <details>
      <summary class="mv-1">Add new alias</summary>
      <form onsubmit={addAlias}>
        <input type="text" bind:value={newAliasKeyword} placeholder="Keyword..." class="full-width" />
        <input type="text" bind:value={newAliasValue} placeholder="Value..." class="full-width" />
        <br />
        <input type="submit" value="Save" class="mt-1" />
      </form>
    </details>
  {/if}
</div>

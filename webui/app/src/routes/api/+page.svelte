<script lang="ts">
  import { onMount } from 'svelte';
  import { apiFetch } from '$lib/api';

  interface EndpointArg {
    name: string;
    type: string;
    required: boolean;
    description: string;
  }

  interface APIEndpoint {
    name: string;
    path: string;
    method: string;
    csrf_required: boolean;
    description: string;
    args: EndpointArg[] | null;
  }

  let endpoints: APIEndpoint[] = $state([]);
  let loading = $state(true);

  function slugify(name: string, method: string): string {
    return name.toLowerCase().replace(/\s+/g, '_') + '_' + method.toLowerCase();
  }

  onMount(async () => {
    try {
      const res = await apiFetch('/api', {
        headers: { 'Accept': 'application/json' }
      });
      if (res.ok) endpoints = await res.json();
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>Hister - API</title>
</svelte:head>

<div class="section">
  <h1>API documentation</h1>
  {#if loading}
    <p>Loading...</p>
  {:else}
    <ul>
      {#each endpoints as ep}
        <li><a href="#{slugify(ep.name, ep.method)}">{ep.name}</a></li>
      {/each}
    </ul>
    {#each endpoints as ep}
      <div class="container" id={slugify(ep.name, ep.method)}>
        <h3>{ep.name}</h3>
        <h2 class="success"><code>{ep.method}</code> <code>{ep.path}</code>{#if ep.csrf_required}<span class="small grey"> CSRF</span>{/if}</h2>
        <p>{ep.description}</p>
        <hr />
        {#if ep.args && ep.args.length > 0}
          <h4>Arguments</h4>
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Type</th>
                <th>Required</th>
                <th>Description</th>
              </tr>
            </thead>
            <tbody>
              {#each ep.args as arg}
                <tr>
                  <td><code>{arg.name}</code></td>
                  <td><code>{arg.type}</code></td>
                  <td>{arg.required}</td>
                  <td>{arg.description}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        {:else}
          <h5>No arguments available for this endpoint</h5>
        {/if}
      </div>
    {/each}
  {/if}
</div>

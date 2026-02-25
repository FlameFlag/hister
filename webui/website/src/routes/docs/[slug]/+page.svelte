<script lang="ts">
  let { data } = $props();
</script>

<svelte:head>
  <title>{data.meta.title} | Hister - Documentation</title>
</svelte:head>

<main>
  <h1>{data.meta.title}</h1>
  <data.content />
</main>

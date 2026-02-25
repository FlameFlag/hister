<script lang="ts">
  let { data } = $props();
</script>

<svelte:head>
  <title>{data.meta.title} | Hister - Web History on Steroids</title>
  {#if data.meta.description}
    <meta name="description" content={data.meta.description as string} />
  {/if}
  <meta property="og:title" content={data.meta.title} />
  {#if data.meta.description}
    <meta property="og:description" content={data.meta.description as string} />
  {/if}
  <meta property="og:type" content="article" />
  <meta property="article:section" content="Post" />
</svelte:head>

<main>
  <time datetime={data.date}>{data.date}</time>
  <h1>{data.meta.title}</h1>
  <data.content />
</main>

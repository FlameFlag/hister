<script lang="ts">
  let { data } = $props();
</script>

<svelte:head>
  <title>Posts | Hister - Web History on Steroids</title>
</svelte:head>

<main>
  <h1>Posts</h1>

  {#each data.posts as post}
    <section>
      <h2><a href="/posts/{post.slug}/">{post.title}</a> <span class="grey small">({new Date(post.date).toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' })})</span></h2>
    </section>
  {/each}
</main>

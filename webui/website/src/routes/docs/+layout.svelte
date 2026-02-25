<script lang="ts">
  let { children, data } = $props();
</script>

<div class="container">
  <div class="sidebar">
    {#each data.docs as doc}
      <section>
        <a href="/docs/{doc.slug}">{doc.title}</a>
      </section>
    {/each}
  </div>

  <main>
    {@render children()}
  </main>
</div>

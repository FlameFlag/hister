<script lang="ts">
  import '@hister/components/style.css';
  let { children } = $props();
</script>

<nav>
  <a href="/">Home</a>
  <a href="/docs">Documentation</a>
  <a href="/posts">Posts</a>
</nav>

<div class="container">
  <main>
    {@render children()}
  </main>
</div>

<footer>
  <div>
    <a href="https://github.com/asciimoo/hister">GitHub</a> &bull; 
    <a href="https://codeberg.org/asciimoo/hister">Codeberg</a>
  </div>
  <div style="margin-top: 10px; font-size: 0.9rem;">AGPLv3</div>
</footer>

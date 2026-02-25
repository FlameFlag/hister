<script lang="ts">
  let { data } = $props();
</script>

<svelte:head>
  <title>Documentation | Hister - Web History on Steroids</title>
</svelte:head>

<main>
    <h1>Documentation</h1>
    <p>Welcome to Hister documentation. Hister is a web history management tool that provides blazing fast, content-based search for visited websites.</p>
    
    <h2 id="what-is-hister">What is Hister?</h2>
    <p>Unlike traditional browser history that only searches URLs and titles, Hister indexes full content of web pages you visit. Keep your data private and local while enjoying powerful search capabilities.</p>
    <p>Hister consists of two major components:</p>
    <ul>
      <li><strong>Binary executable</strong>: Providing indexer, web interface and command line tool</li>
      <li><strong>Browser extension</strong>: Sending visited websites and opened results to indexer</li>
    </ul>

    <h2 id="quick-links">Quick Links</h2>
    <ul>
      {#each data.docs as doc}
        <li><a href="/docs/{doc.slug}">{doc.title}</a></li>
      {/each}
      <li><a href="https://github.com/asciimoo/hister">GitHub Repository</a></li>
      <li><a href="https://github.com/asciimoo/hister/issues">Issue Tracker</a></li>
      <li><a href="https://codeberg.org/asciimoo/hister">Codeberg Repository</a></li>
    </ul>

    <h2 id="browser-extensions">Browser Extensions</h2>
    <p>Hister provides browser extensions to automatically index visited content:</p>
    <ul>
      <li><a href="https://chromewebstore.google.com/detail/hister/cciilamhchpmbdnniabclekddabkifhb">Chrome Extension</a></li>
      <li><a href="https://addons.mozilla.org/en-US/firefox/addon/hister/">Firefox Extension</a></li>
    </ul>

    <h2 id="getting-help">Getting Help</h2>
    <p>If you encounter bugs or have suggestions, please visit <a href="https://github.com/asciimoo/hister/issues">issue tracker</a> on GitHub.</p>

  <h2 id="license">License</h2>
  <p>Hister is licensed under AGPLv3 license.</p>
</main>

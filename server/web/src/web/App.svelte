<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import Home from './Home.svelte';
  import Search from './Search.svelte';
  import Rules from './Rules.svelte';
  import Add from './Add.svelte';
  import { ModeWatcher } from 'mode-watcher';
  import { initBackgroundStats, cleanupBackgroundStats } from '$lib/api';

  type Page = 'home' | 'search' | 'rules' | 'add';

  let currentPage: Page = $state('home');
  let searchQuery = $state('');

  function updatePageFromUrl() {
    const path = window.location.pathname;
    const params = new URLSearchParams(window.location.search);

    if (path === '/search') {
      currentPage = 'search';
      if (params.has('q')) {
        searchQuery = params.get('q') || '';
      }
    } else if (path === '/rules') {
      currentPage = 'rules';
    } else if (path === '/add') {
      currentPage = 'add';
    } else if (path === '/' || path === '') {
      currentPage = 'home';
    }
  }

  onMount(() => {
    // Initialize background stats loading
    initBackgroundStats();

    updatePageFromUrl();

    // Listen for popstate events (back/forward buttons)
    window.addEventListener('popstate', updatePageFromUrl);

    return () => {
      window.removeEventListener('popstate', updatePageFromUrl);
    };
  });

  onDestroy(() => {
    cleanupBackgroundStats();
  });

  // Update page when URL changes
  $effect(() => {
    updatePageFromUrl();
  });
</script>

<ModeWatcher />

{#if currentPage === 'home'}
  <Home />
{:else if currentPage === 'search'}
  <Search query={searchQuery} />
{:else if currentPage === 'rules'}
  <Rules />
{:else if currentPage === 'add'}
  <Add />
{/if}

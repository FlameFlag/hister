<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';

  let { children } = $props();

  onMount(() => {
    const theme = localStorage.getItem('theme') ||
      (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
    document.documentElement.setAttribute('data-theme', theme);
  });

  function toggleTheme() {
    const current = document.documentElement.getAttribute('data-theme');
    const next = current === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', next);
    localStorage.setItem('theme', next);
  }
</script>

<header>
  <h1 class="menu-item">
    <img src="/static/logo.png" alt="Hister logo" />
    <a href="/">Hister</a>
  </h1>
  <a class="menu-item" class:active={$page.url.pathname === '/history'} href="/history">History</a>
  <a class="menu-item" class:active={$page.url.pathname === '/rules'} href="/rules">Rules</a>
  <a class="menu-item" class:active={$page.url.pathname === '/add'} href="/add">Add</a>
  <button id="theme-toggle" class="theme-toggle float-right" title="Toggle theme" onclick={toggleTheme}>
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <circle cx="12" cy="12" r="5"/>
      <line x1="12" y1="1" x2="12" y2="3"/>
      <line x1="12" y1="21" x2="12" y2="23"/>
      <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
      <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
      <line x1="1" y1="12" x2="3" y2="12"/>
      <line x1="21" y1="12" x2="23" y2="12"/>
      <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
      <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
    </svg>
  </button>
</header>

<main>
  {@render children()}
</main>

<footer>
  <a href="/help">Help</a> |
  <a href="/about">About</a> |
  <a href="/api">API</a> |
  <a href="https://github.com/asciimoo/hister/">GitHub</a>
</footer>

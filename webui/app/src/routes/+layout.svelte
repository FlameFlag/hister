<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { Button } from '@hister/components/ui/button';
  import "../style.css";

  let { children } = $props();

  function applyTheme(theme: string) {
    document.documentElement.setAttribute('data-theme', theme);
    if (theme === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }

  onMount(() => {
    const theme = localStorage.getItem('theme') ||
      (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
    applyTheme(theme);
  });

  function toggleTheme() {
    const current = document.documentElement.getAttribute('data-theme');
    const next = current === 'dark' ? 'light' : 'dark';
    applyTheme(next);
    localStorage.setItem('theme', next);
  }
</script>

<header class="h-16 px-6 bg-page-bg border-b-[2px] border-border-brand flex items-center justify-between shadow-[4px_4px_0px_var(--hister-indigo)]">
  <h1 class="flex items-center gap-2">
    <img src="static/logo.png" alt="Hister logo" class="h-8 w-8" />
    <a data-sveltekit-reload href="./" class="font-outfit text-xl font-extrabold text-text-brand no-underline hover:underline">
      Hister
    </a>
  </h1>
  <nav class="flex items-center gap-6">
    <a
      class:underline={$page.url.pathname === new URL('history', $page.url).pathname}
      class="font-inter text-sm font-semibold text-text-brand-secondary hover:text-text-brand no-underline hover:underline"
      href="history"
    >
      History
    </a>
    <a
      class:underline={$page.url.pathname === new URL('rules', $page.url).pathname}
      class="font-inter text-sm font-semibold text-text-brand-secondary hover:text-text-brand no-underline hover:underline"
      href="rules"
    >
      Rules
    </a>
    <a
      class:underline={$page.url.pathname === new URL('add', $page.url).pathname}
      class="font-inter text-sm font-semibold text-text-brand-secondary hover:text-text-brand no-underline hover:underline"
      href="add"
    >
      Add
    </a>
  </nav>
  <button
    type="button"
    class="text-text-brand-muted hover:text-hister-indigo transition-all hover:scale-110 bg-transparent border-0 cursor-pointer p-1"
    title="Toggle theme"
    onclick={toggleTheme}
  >
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="size-6">
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

<main class="flex-1">
  {@render children()}
</main>

<footer class="h-12 px-6 bg-page-bg border-t-[2px] border-border-brand flex items-center justify-center gap-4 text-sm">
  <a href="help" class="font-inter text-text-brand-secondary hover:text-hister-indigo no-underline hover:underline">Help</a>
  <span class="text-text-brand-muted">|</span>
  <a href="about" class="font-inter text-text-brand-secondary hover:text-hister-indigo no-underline hover:underline">About</a>
  <span class="text-text-brand-muted">|</span>
  <a href="api" class="font-inter text-text-brand-secondary hover:text-hister-indigo no-underline hover:underline">API</a>
  <span class="text-text-brand-muted">|</span>
  <a href="https://github.com/asciimoo/hister/" class="font-inter text-text-brand-secondary hover:text-hister-indigo no-underline hover:underline" target="_blank" rel="noopener">GitHub</a>
</footer>

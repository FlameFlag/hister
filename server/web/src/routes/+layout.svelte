<script lang="ts">
  import { Moon, Sun } from 'lucide-svelte';
  import { toggleMode, ModeWatcher } from 'mode-watcher';
  import { page } from '$app/state';
  import { fly } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { Button } from '$lib/components/ui/button';
  import { Badge } from '$lib/components/ui/badge';
  import * as Tooltip from '$lib/components/ui/tooltip';
  import '../app.css';

  let { children } = $props();

  const navItems = [
    { href: '/', label: 'Home' },
    { href: '/history', label: 'History' },
    { href: '/rules', label: 'Rules' },
    { href: '/add', label: 'Add' },
  ];
</script>

<ModeWatcher />

<Tooltip.Provider>
  <div class="flex min-h-screen flex-col bg-background font-sans antialiased">
    <header
      class="
        flex h-18 items-center justify-between border-b border-border
        bg-background px-12
      "
    >
      <div class="flex items-center gap-2">
        <span
          class="
            font-display text-[28px] font-bold tracking-[-0.5px] text-primary
          ">Hister</span
        >
      </div>

      <nav class="flex items-center gap-2">
        {#each navItems as item (item.href)}
          {#if page.url.pathname === item.href}
            <Badge
              variant="default"
              class="px-4 py-2 text-[15px] font-semibold"
            >
              {item.label}
            </Badge>
          {:else}
            <Button
              variant="ghost"
              href={item.href}
              class="h-10 cursor-pointer text-[15px] font-medium"
            >
              {item.label}
            </Button>
          {/if}
        {/each}
      </nav>

      <Button
        variant="ghost"
        size="icon"
        onclick={toggleMode}
        class="size-10 cursor-pointer"
        aria-label="Toggle theme"
      >
        <div class="relative flex items-center justify-center">
          <!-- relative wrapper creates positioning context for absolutely stacked icons -->
          <Sun
            class="
              absolute size-5 scale-100 rotate-0 transition-all
              dark:scale-0 dark:-rotate-90
            "
          />
          <Moon
            class="
              absolute size-5 scale-0 rotate-90 transition-all
              dark:scale-100 dark:rotate-0
            "
          />
        </div>
      </Button>
    </header>

    <div class="flex flex-1 flex-col">
      {#key page.url.pathname}
        <div class="flex flex-1 flex-col" in:fly={{ y: 10, duration: 250, easing: cubicOut }}>
          {@render children?.()}
        </div>
      {/key}
    </div>
  </div>
</Tooltip.Provider>

<script lang="ts">
  import { Search, Settings, Menu } from 'lucide-svelte';
  import { Button, Input } from '$lib/components/ui';

  let searchQuery = $state('');
  let isMenuOpen = $state(false);

  function handleSearch(e: KeyboardEvent) {
    if (e.key === 'Enter' && searchQuery.trim()) {
      console.log('Searching:', searchQuery);
    }
  }

  function openSettings() {
    chrome.runtime.openOptionsPage();
  }
</script>

<div class="min-h-125 w-110 bg-background font-sans antialiased">
  <header class="flex h-16 items-center justify-end px-6">
    <nav class="flex items-center gap-3">
      <Button
        variant="ghost"
        size="sm"
        onclick={openSettings}
        class="
          text-muted-foreground
          hover:text-foreground
        "
      >
        <Settings class="size-4" />
        <span>Settings</span>
      </Button>
      <div class="relative">
        <Button
          variant="ghost"
          size="sm"
          onclick={() => (isMenuOpen = !isMenuOpen)}
          class="
            text-muted-foreground
            hover:text-foreground
          "
        >
          <Menu class="size-4" />
          <span>Menu</span>
        </Button>
        {#if isMenuOpen}
          <div
            class="
              absolute top-full right-0 z-50 mt-2 w-48 rounded-lg border
              border-border bg-popover p-2 shadow-lg
            "
          >
            <button
              type="button"
              class="
                flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm
                hover:bg-secondary
              "
            >
              <Settings class="size-4" />
              Preferences
            </button>
          </div>
        {/if}
      </div>
    </nav>
  </header>

  <main class="flex flex-col items-center px-6 py-8">
    <div class="mt-8 w-full max-w-md">
      <div
        class="
          flex items-center gap-4 rounded-full bg-card px-6 py-4 shadow-sm
          transition-shadow
          focus-within:shadow-md
        "
      >
        <Search class="size-6 shrink-0 text-muted-foreground" />
        <Input
          bind:value={searchQuery}
          placeholder="Search your browsing history..."
          onkeydown={handleSearch}
          autofocus
          class="text-lg"
        />
      </div>
    </div>
  </main>
</div>

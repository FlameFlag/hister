<script lang="ts">
  import {
    ArrowLeft,
    Menu,
    Settings,
    Moon,
    Sun,
    Trash2,
    Plus,
    Upload,
    Globe,
    History,
    ListTree,
    EllipsisVertical,
    ArrowUp,
    Download,
    ChevronDown,
  } from 'lucide-svelte';
  import { Button } from '$lib/components/ui/button';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import { ModeWatcher, toggleMode } from 'mode-watcher';

  type Tab = 'history' | 'rules' | 'add';
  let activeTab = $state<Tab>('history');

  // History stats
  const historyStats = {
    totalPages: '12,847',
    uniqueDomains: '342',
    daysTracked: '90',
  };

  // Rules data
  const rules = [
    {
      id: 1,
      name: 'Prioritize GitHub',
      description:
        'Boost results from github.com to the top of search results when searching for code-related queries',
      tag: 'domain:github.com',
    },
  ];

  function deleteRule(id: number) {
    console.log('Delete rule', id);
  }

  function prioritizeRule(id: number) {
    console.log('Prioritize rule', id);
  }

  function exportData(format: string) {
    console.log('Export as', format);
  }

  function clearAllHistory() {
    console.log('Clear all history');
  }

  function importFromBrowser(browser: string) {
    console.log('Import from', browser);
  }
</script>

<ModeWatcher />

<div class="min-h-screen bg-background font-sans antialiased">
  <!-- Header -->
  <header class="
    flex h-16 items-center justify-between border-b border-border px-6
  ">
    <div class="flex items-center gap-6">
      <button
        type="button"
        onclick={() => window.close()}
        class="
          flex items-center gap-2 text-muted-foreground
          hover:text-foreground
        "
      >
        <ArrowLeft class="size-4" />
        <span>Back</span>
      </button>

      <nav class="flex items-center gap-3">
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            {#snippet child({ props })}
              <Button
                variant="ghost"
                size="sm"
                {...props}
                class="
                  gap-2 text-muted-foreground
                  hover:text-foreground
                "
              >
                <Menu class="size-4" />
                <span>Menu</span>
              </Button>
            {/snippet}
          </DropdownMenu.Trigger>
          <DropdownMenu.Content>
            <DropdownMenu.Label>Navigation</DropdownMenu.Label>
            <DropdownMenu.Separator />
            <DropdownMenu.Item onclick={() => (window.location.href = '/')}>Home</DropdownMenu.Item>
            <DropdownMenu.Item onclick={() => (window.location.href = '/search')}>
              Search
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>

        <Button
          variant="ghost"
          size="sm"
          onclick={() => window.location.reload()}
          class="
            gap-2 text-muted-foreground
            hover:text-foreground
          "
        >
          <Settings class="size-4" />
          <span>Settings</span>
        </Button>
      </nav>
    </div>

    <!-- Single Icon Theme Toggle -->
    <Button
      variant="ghost"
      size="icon"
      onclick={toggleMode}
      class="
        relative size-9 cursor-pointer text-muted-foreground
        hover:text-foreground
      "
    >
      <Sun
        class="
          h-[1.2rem] w-[1.2rem] scale-100 rotate-0 transition-all
          dark:scale-0 dark:-rotate-90
        "
      />
      <Moon
        class="
          absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 transition-all
          dark:scale-100 dark:rotate-0
        "
      />
      <span class="sr-only">Toggle theme</span>
    </Button>
  </header>

  <!-- Main Content -->
  <main class="mx-auto max-w-4xl p-8">
    <!-- Page Title -->
    <div class="mb-8 space-y-2">
      <h1 class="font-display text-4xl font-bold tracking-tight">Manage Hister</h1>
      <p class="text-muted-foreground">Configure your history search settings and preferences</p>
    </div>

    <!-- Tabs -->
    <div class="mb-8 border-b border-border">
      <div class="flex gap-1">
        <button
          class="
            flex items-center gap-2 border-b-2 px-4 py-3 text-sm font-medium
            transition-colors
            {activeTab === 'history'
            ? 'border-foreground text-foreground'
            : `
              border-transparent text-muted-foreground
              hover:text-foreground
            `}"
          onclick={() => (activeTab = 'history')}
        >
          <History class="size-4" />
          History
        </button>
        <button
          class="
            flex items-center gap-2 border-b-2 px-4 py-3 text-sm font-medium
            transition-colors
            {activeTab === 'rules'
            ? 'border-foreground text-foreground'
            : `
              border-transparent text-muted-foreground
              hover:text-foreground
            `}"
          onclick={() => (activeTab = 'rules')}
        >
          <ListTree class="size-4" />
          Rules
        </button>
        <button
          class="
            flex items-center gap-2 border-b-2 px-4 py-3 text-sm font-medium
            transition-colors
            {activeTab === 'add'
            ? 'border-foreground text-foreground'
            : `
              border-transparent text-muted-foreground
              hover:text-foreground
            `}"
          onclick={() => (activeTab = 'add')}
        >
          <Plus class="size-4" />
          Add
        </button>
      </div>
    </div>

    <!-- History Tab Content -->
    {#if activeTab === 'history'}
      <div class="space-y-6">
        <!-- Header with Clear All -->
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-semibold">Browsing History</h2>
          <Button variant="destructive" size="sm" onclick={clearAllHistory} class="
            gap-2
          ">
            <Trash2 class="size-4" />
            Clear All
          </Button>
        </div>

        <!-- Stats -->
        <div
          class="
            grid gap-4 rounded-xl bg-muted p-6
            sm:grid-cols-3
          "
        >
          <div class="space-y-1">
            <div class="font-display text-3xl font-bold">{historyStats.totalPages}</div>
            <div class="text-sm text-muted-foreground">Total Pages</div>
          </div>
          <div class="space-y-1">
            <div class="font-display text-3xl font-bold">{historyStats.uniqueDomains}</div>
            <div class="text-sm text-muted-foreground">Unique Domains</div>
          </div>
          <div class="space-y-1">
            <div class="font-display text-3xl font-bold">{historyStats.daysTracked}</div>
            <div class="text-sm text-muted-foreground">Days Tracked</div>
          </div>
        </div>

        <!-- Export Section with Dropdown -->
        <div class="rounded-xl border border-border bg-card p-6">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="font-semibold">Export History</h3>
              <p class="text-sm text-muted-foreground">Download your browsing history data</p>
            </div>
            <DropdownMenu.Root>
              <DropdownMenu.Trigger>
                {#snippet child({ props })}
                  <Button variant="outline" {...props} class="gap-2">
                    <Download class="size-4" />
                    Export As
                    <ChevronDown class="size-3" />
                  </Button>
                {/snippet}
              </DropdownMenu.Trigger>
              <DropdownMenu.Content align="end">
                <DropdownMenu.Label>Choose Format</DropdownMenu.Label>
                <DropdownMenu.Separator />
                <DropdownMenu.Item onclick={() => exportData('json')}>
                  Export as JSON
                </DropdownMenu.Item>
                <DropdownMenu.Item onclick={() => exportData('csv')}>
                  Export as CSV
                </DropdownMenu.Item>
                <DropdownMenu.Item onclick={() => exportData('html')}>
                  Export as HTML
                </DropdownMenu.Item>
              </DropdownMenu.Content>
            </DropdownMenu.Root>
          </div>
        </div>
      </div>
    {/if}

    <!-- Rules Tab Content -->
    {#if activeTab === 'rules'}
      <div class="space-y-6">
        <!-- Header with Add Rule -->
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-semibold">Search Rules</h2>
          <Button size="sm" class="gap-2">
            <Plus class="size-4" />
            Add Rule
          </Button>
        </div>

        <!-- Rules List -->
        <div class="space-y-4">
          {#each rules as rule (rule.id)}
            <div class="rounded-xl border border-border bg-card p-6">
              <div class="flex items-start justify-between gap-4">
                <div class="flex-1 space-y-2">
                  <div class="flex items-center gap-3">
                    <h3 class="font-semibold">{rule.name}</h3>
                    <span
                      class="
                        rounded-full bg-secondary px-2.5 py-0.5 text-xs
                        text-secondary-foreground
                      "
                    >
                      {rule.tag}
                    </span>
                  </div>
                  <p class="text-sm text-muted-foreground">{rule.description}</p>
                </div>

                <!-- 3-Dot Action Menu -->
                <DropdownMenu.Root>
                  <DropdownMenu.Trigger>
                    {#snippet child({ props })}
                      <Button
                        {...props}
                        variant="ghost"
                        size="icon"
                        class="
                          size-8 shrink-0 text-muted-foreground
                          hover:text-foreground
                        "
                      >
                        <EllipsisVertical class="size-4" />
                        <span class="sr-only">Open menu</span>
                      </Button>
                    {/snippet}
                  </DropdownMenu.Trigger>
                  <DropdownMenu.Content align="end">
                    <DropdownMenu.Label>Actions</DropdownMenu.Label>
                    <DropdownMenu.Separator />
                    <DropdownMenu.Item onclick={() => prioritizeRule(rule.id)} class="
                      gap-2
                    ">
                      <ArrowUp class="size-4" />
                      Prioritize
                    </DropdownMenu.Item>
                    <DropdownMenu.Separator />
                    <DropdownMenu.Item
                      onclick={() => deleteRule(rule.id)}
                      class="
                        gap-2 text-destructive
                        focus:text-destructive
                      "
                    >
                      <Trash2 class="size-4" />
                      Delete
                    </DropdownMenu.Item>
                  </DropdownMenu.Content>
                </DropdownMenu.Root>
              </div>
            </div>
          {/each}
        </div>

        {#if rules.length === 0}
          <div
            class="
              rounded-xl border border-dashed border-border p-12 text-center
            "
          >
            <ListTree class="mx-auto size-12 text-muted-foreground/50" />
            <h3 class="mt-4 font-semibold">No rules yet</h3>
            <p class="mt-2 text-sm text-muted-foreground">
              Create rules to customize your search results
            </p>
            <Button class="mt-4 gap-2">
              <Plus class="size-4" />
              Add Rule
            </Button>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Add Tab Content -->
    {#if activeTab === 'add'}
      <div class="space-y-6">
        <div>
          <h2 class="text-xl font-semibold">Add New Content</h2>
          <p class="text-muted-foreground">Import your browsing history from supported sources</p>
        </div>

        <div
          class="
            grid gap-4
            sm:grid-cols-2
            lg:grid-cols-3
          "
        >
          <!-- Chrome -->
          <div class="rounded-xl border border-border bg-card p-6">
            <div
              class="
                mb-4 flex size-12 items-center justify-center rounded-xl
                bg-secondary
              "
            >
              <Globe class="size-6" />
            </div>
            <h3 class="font-semibold">Chrome</h3>
            <p class="mb-4 text-sm text-muted-foreground">
              Import from Google Chrome browser history
            </p>
            <Button onclick={() => importFromBrowser('chrome')} class="
              w-full gap-2
            ">
              <Upload class="size-4" />
              Import
            </Button>
          </div>

          <!-- Firefox -->
          <div class="rounded-xl border border-border bg-card p-6">
            <div
              class="
                mb-4 flex size-12 items-center justify-center rounded-xl
                bg-secondary
              "
            >
              <Globe class="size-6" />
            </div>
            <h3 class="font-semibold">Firefox</h3>
            <p class="mb-4 text-sm text-muted-foreground">Import from Firefox browser history</p>
            <Button onclick={() => importFromBrowser('firefox')} class="
              w-full gap-2
            ">
              <Upload class="size-4" />
              Import
            </Button>
          </div>

          <!-- Safari -->
          <div class="rounded-xl border border-border bg-card p-6">
            <div
              class="
                mb-4 flex size-12 items-center justify-center rounded-xl
                bg-secondary
              "
            >
              <Globe class="size-6" />
            </div>
            <h3 class="font-semibold">Safari</h3>
            <p class="mb-4 text-sm text-muted-foreground">Import from Safari browser history</p>
            <Button onclick={() => importFromBrowser('safari')} class="
              w-full gap-2
            ">
              <Upload class="size-4" />
              Import
            </Button>
          </div>
        </div>
      </div>
    {/if}
  </main>
</div>

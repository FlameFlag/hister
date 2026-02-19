<script lang="ts">
  import { onMount } from 'svelte';
  import { Plus, Trash2, Moon, Sun, Loader, Tag } from 'lucide-svelte';
  import { toggleMode } from 'mode-watcher';
  import { fetchRules, saveRules, addAlias, deleteAlias, type Rules } from '$lib/api';
  import Button from '$lib/components/ui/button/button.svelte';
  import { Input } from '$lib/components/ui/input';
  import { Textarea } from '$lib/components/ui/textarea';
  import { Badge } from '$lib/components/ui/badge';
  import * as Card from '$lib/components/ui/card';
  import * as Alert from '$lib/components/ui/alert';
  import * as Item from '$lib/components/ui/item';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import * as Empty from '$lib/components/ui/empty';
  import * as NavigationMenu from '$lib/components/ui/navigation-menu';
  import * as Field from '$lib/components/ui/field';

  let rules = $state<Rules>({
    skip: [],
    priority: [],
    aliases: {},
  });

  let loading = $state(true);
  let saving = $state(false);
  let saveMessage = $state('');

  // For the textarea inputs (space-separated patterns)
  let skipInput = $state('');
  let priorityInput = $state('');

  // For adding new aliases
  let newAliasKey = $state('');
  let newAliasValue = $state('');

  onMount(async () => {
    try {
      rules = await fetchRules();
      skipInput = rules.skip.join('\n');
      priorityInput = rules.priority.join('\n');
    } catch (err) {
      console.error('Failed to load rules:', err);
    } finally {
      loading = false;
    }
  });

  async function handleSaveRules() {
    saving = true;
    saveMessage = '';

    try {
      // Parse space-separated patterns from textarea
      const updatedRules: Rules = {
        skip: skipInput.split(/\s+/).filter((s) => s.trim() !== ''),
        priority: priorityInput.split(/\s+/).filter((s) => s.trim() !== ''),
        aliases: rules.aliases,
      };

      await saveRules(updatedRules);
      rules = updatedRules;
      saveMessage = 'Rules saved successfully!';

      setTimeout(() => {
        saveMessage = '';
      }, 3000);
    } catch (err) {
      console.error('Failed to save rules:', err);
      saveMessage = 'Failed to save rules. Please try again.';
    } finally {
      saving = false;
    }
  }

  async function handleAddAlias() {
    if (!newAliasKey.trim() || !newAliasValue.trim()) return;

    try {
      await addAlias(newAliasKey.trim(), newAliasValue.trim());
      rules.aliases[newAliasKey.trim()] = newAliasValue.trim();
      newAliasKey = '';
      newAliasValue = '';
      saveMessage = 'Alias added successfully!';

      setTimeout(() => {
        saveMessage = '';
      }, 3000);
    } catch (err) {
      console.error('Failed to add alias:', err);
      saveMessage = 'Failed to add alias. Please try again.';
    }
  }

  async function handleDeleteAlias(key: string) {
    try {
      await deleteAlias(key);
      const newAliases = { ...rules.aliases };
      delete newAliases[key];
      rules.aliases = newAliases;
      saveMessage = 'Alias deleted successfully!';

      setTimeout(() => {
        saveMessage = '';
      }, 3000);
    } catch (err) {
      console.error('Failed to delete alias:', err);
      saveMessage = 'Failed to delete alias. Please try again.';
    }
  }
</script>

<svelte:head>
  <title>History Rules - Hister</title>
</svelte:head>

<div class="flex min-h-screen flex-col bg-background font-sans antialiased">
  <!-- Header -->
  <header
    class="
      flex h-18 items-center justify-between border-b border-border
      bg-background px-12
    "
  >
    <!-- Logo Section -->
    <div class="flex items-center gap-2">
      <span
        class="
          font-display text-[28px] font-bold tracking-[-0.5px] text-primary
        ">Hister</span
      >
    </div>

    <!-- Navigation -->
    <nav class="flex items-center gap-2">
      <NavigationMenu.Root>
        <NavigationMenu.List>
          <NavigationMenu.Item>
            <NavigationMenu.Link onclick={() => (window.location.href = '/')}>
              History
            </NavigationMenu.Link>
          </NavigationMenu.Item>
          <NavigationMenu.Item>
            <Badge variant="default">Rules</Badge>
          </NavigationMenu.Item>
          <NavigationMenu.Item>
            <NavigationMenu.Link onclick={() => (window.location.href = '/add')}>
              Add
            </NavigationMenu.Link>
          </NavigationMenu.Item>
        </NavigationMenu.List>
      </NavigationMenu.Root>
    </nav>

    <!-- Theme Toggle -->
    <Button variant="ghost" size="icon" onclick={toggleMode} aria-label="Toggle theme">
      <Sun
        class="
          size-5
          dark:hidden
        "
      />
      <Moon
        class="
          hidden size-5
          dark:block
        "
      />
    </Button>
  </header>

  <!-- Main Content -->
  <main class="flex flex-1 flex-col items-center px-16 py-12">
    <h1 class="mb-10 font-display text-[32px] font-bold text-foreground">History Rules</h1>

    {#if saveMessage}
      <Alert.Root
        class="
          mb-6 w-full max-w-248 border-primary/50 bg-primary/10 text-primary
        "
      >
        <Alert.Description>{saveMessage}</Alert.Description>
      </Alert.Root>
    {/if}

    {#if loading}
      <div
        class="
          flex w-full max-w-248 flex-col gap-8
          md:flex-row
        "
      >
        <div class="flex-1 space-y-4">
          <Skeleton class="h-6 w-1/2" />
          <Skeleton class="h-32 w-full" />
          <Skeleton class="h-10 w-full" />
        </div>
        <div class="flex-1 space-y-4">
          <Skeleton class="h-6 w-1/2" />
          <Skeleton class="h-32 w-full" />
          <Skeleton class="h-10 w-full" />
        </div>
      </div>
    {:else}
      <div
        class="
          flex w-full max-w-248 flex-col gap-8
          md:flex-row
        "
      >
        <!-- Skip Rules Section -->
        <Card.Card class="min-w-0 flex-1 bg-[#FAFAFA]">
          <Card.CardHeader>
            <Card.CardTitle>Skip Rules</Card.CardTitle>
            <Card.CardDescription>
              Define regexps to forbid indexing matching URLs
            </Card.CardDescription>
          </Card.CardHeader>
          <Card.CardContent class="space-y-6">
            <Textarea
              bind:value={skipInput}
              placeholder="*.google.com/search*
*.facebook.com/*
localhost:*"
              rows={5}
              class="font-mono"
            />
            <Button onclick={handleSaveRules} disabled={saving} class="w-full">
              {#if saving}
                <Loader class="mr-2 size-4 animate-spin" />
                Saving…
              {:else}
                Save
              {/if}
            </Button>
          </Card.CardContent>
        </Card.Card>

        <!-- Priority Rules Section -->
        <Card.Card class="min-w-0 flex-1 bg-[#FAFAFA]">
          <Card.CardHeader>
            <Card.CardTitle>Priority Rules</Card.CardTitle>
            <Card.CardDescription>Define regexps to prioritize matching URLs</Card.CardDescription>
          </Card.CardHeader>
          <Card.CardContent class="space-y-6">
            <Textarea
              bind:value={priorityInput}
              placeholder="github.com/*
stackoverflow.com/*
docs.python.org/*"
              rows={5}
              class="font-mono"
            />
            <Button onclick={handleSaveRules} disabled={saving} class="w-full">
              {#if saving}
                <Loader class="mr-2 size-4 animate-spin" />
                Saving…
              {:else}
                Save
              {/if}
            </Button>
          </Card.CardContent>
        </Card.Card>
      </div>

      <!-- Aliases Section -->
      <Card.Card class="mt-8 w-full max-w-248 bg-[#FAFAFA]">
        <Card.CardHeader>
          <Card.CardTitle>Search Keyword Aliases</Card.CardTitle>
          <Card.CardDescription>
            Define aliases to simplify queries. Alias strings in queries are automatically replaced
            with the provided value.
          </Card.CardDescription>
        </Card.CardHeader>
        <Card.CardContent class="space-y-6">
          <!-- Existing Aliases -->
          {#if Object.keys(rules.aliases).length > 0}
            <Item.Group class="space-y-2">
              {#each Object.entries(rules.aliases) as [key, value] (key)}
                <Item.Root>
                  <Badge variant="secondary" class="min-w-30">{key}</Badge>
                  <Item.Content class="flex items-center gap-2">
                    <span class="text-[#71717A]">→</span>
                    <code class="flex-1 text-sm text-foreground">{value}</code>
                  </Item.Content>
                  <Item.Actions>
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      onclick={() => handleDeleteAlias(key)}
                      aria-label="Delete alias"
                    >
                      <Trash2 class="size-4" />
                    </Button>
                  </Item.Actions>
                </Item.Root>
              {/each}
            </Item.Group>
          {:else}
            <Empty.Root>
              <Empty.Header>
                <Empty.Media variant="icon">
                  <Tag class="size-6 text-muted-foreground" />
                </Empty.Media>
                <Empty.Title class="text-base">No Aliases Defined</Empty.Title>
                <Empty.Description class="text-xs">
                  Create aliases to simplify your search queries
                </Empty.Description>
              </Empty.Header>
            </Empty.Root>
          {/if}

          <!-- Add New Alias -->
          <Field.Label class="sr-only">Keyword</Field.Label>
          <Field.Label class="sr-only">Value</Field.Label>
          <div class="flex gap-2">
            <Input
              type="text"
              placeholder="Keyword (e.g., 'gh')"
              bind:value={newAliasKey}
              onkeydown={(e: KeyboardEvent) => e.key === 'Enter' && handleAddAlias()}
              class="bg-background"
            />
            <Input
              type="text"
              placeholder="Value (e.g., 'github.com')"
              bind:value={newAliasValue}
              onkeydown={(e: KeyboardEvent) => e.key === 'Enter' && handleAddAlias()}
              class="bg-background"
            />
            <Button
              onclick={handleAddAlias}
              disabled={!newAliasKey.trim() || !newAliasValue.trim()}
            >
              <Plus class="size-4" />
              Add
            </Button>
          </div>
        </Card.CardContent>
      </Card.Card>
    {/if}
  </main>
</div>

<script lang="ts">
  import { onMount } from 'svelte';
  import { Plus, Trash2, Loader } from 'lucide-svelte';
  import { fetchRules, saveRules, addAlias, deleteAlias, type Rules } from '$lib/api';
  import {
    Card,
    CardContent,
    CardHeader,
    CardTitle,
    CardDescription,
  } from '$lib/components/ui/card';
  import { Textarea } from '$lib/components/ui/textarea';
  import { Input } from '$lib/components/ui/input';
  import { Button } from '$lib/components/ui/button';
  import { Badge } from '$lib/components/ui/badge';
  import { Alert } from '$lib/components/ui/alert';

  let rules = $state<Rules>({
    skip: [],
    priority: [],
    aliases: {},
  });

  let loading = $state(true);
  let saving = $state(false);
  let saveMessage = $state('');

  let skipInput = $state('');
  let priorityInput = $state('');
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

<main class="flex flex-1 flex-col items-center px-16 py-12">
  <h1 class="mb-10 font-display text-[32px] font-bold text-foreground">History Rules</h1>

  {#if saveMessage}
    <Alert
      variant={saveMessage.includes('Failed') ? 'destructive' : 'default'}
      class="mb-6 w-full max-w-248 border-primary/50 bg-primary/10 text-primary {saveMessage.includes(
        'Failed'
      )
        ? 'border-destructive/50 bg-destructive/10 text-destructive'
        : ''}"
    >
      {saveMessage}
    </Alert>
  {/if}

  {#if loading}
    <div class="flex items-center gap-2 text-muted-foreground">
      <Loader class="size-5 animate-spin" />
      Loading rules…
    </div>
  {:else}
    <div
      class="
        flex w-full max-w-248 flex-col gap-8
        md:flex-row
      "
    >
      <Card class="flex-1">
        <CardHeader>
          <CardTitle>Skip Rules</CardTitle>
          <CardDescription>Define regexps to forbid indexing matching URLs</CardDescription>
        </CardHeader>
        <CardContent>
          <Textarea
            bind:value={skipInput}
            placeholder="*.google.com/search*
*.facebook.com/*
localhost:*"
            rows={5}
            class="mb-6 h-30 resize-none font-mono text-sm"
          />
          <Button onclick={handleSaveRules} disabled={saving} class="w-full">
            {#if saving}
              <Loader class="mr-2 size-4 animate-spin" />
              Saving…
            {:else}
              Save
            {/if}
          </Button>
        </CardContent>
      </Card>

      <Card class="flex-1">
        <CardHeader>
          <CardTitle>Priority Rules</CardTitle>
          <CardDescription>Define regexps to prioritize matching URLs</CardDescription>
        </CardHeader>
        <CardContent>
          <Textarea
            bind:value={priorityInput}
            placeholder="github.com/*
stackoverflow.com/*
docs.python.org/*"
            rows={5}
            class="mb-6 h-30 resize-none font-mono text-sm"
          />
          <Button onclick={handleSaveRules} disabled={saving} class="w-full">
            {#if saving}
              <Loader class="mr-2 size-4 animate-spin" />
              Saving…
            {:else}
              Save
            {/if}
          </Button>
        </CardContent>
      </Card>
    </div>

    <Card class="mt-8 w-full max-w-248">
      <CardHeader>
        <CardTitle>Search Keyword Aliases</CardTitle>
        <CardDescription>
          Define aliases to simplify queries. Alias strings in queries are automatically replaced
          with provided value.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {#if Object.keys(rules.aliases).length > 0}
          <div class="mb-6 space-y-2">
            {#each Object.entries(rules.aliases) as [key, value] (key)}
              <div class="flex items-center gap-3 rounded-lg bg-muted p-3">
                <Badge
                  variant="secondary"
                  class="
                  min-w-30 justify-center font-mono text-sm
                "
                >
                  {key}
                </Badge>
                <span class="text-muted-foreground">→</span>
                <code class="flex-1 text-sm text-foreground">{value}</code>
                <Button
                  variant="ghost"
                  size="icon"
                  class="size-8"
                  onclick={() => handleDeleteAlias(key)}
                >
                  <Trash2 class="size-4" />
                </Button>
              </div>
            {/each}
          </div>
        {:else}
          <p class="mb-6 text-sm text-muted-foreground italic">No aliases defined yet.</p>
        {/if}

        <form
          onsubmit={(e) => {
            e.preventDefault();
            handleAddAlias();
          }}
          class="flex gap-2"
        >
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
          <Button type="submit" disabled={!newAliasKey.trim() || !newAliasValue.trim()}>
            <Plus class="size-4" />
            Add
          </Button>
        </form>
      </CardContent>
    </Card>
  {/if}
</main>

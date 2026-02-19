<script lang="ts">
  import { Plus, Moon, Sun, Loader, Check, Link, Type, FileText } from 'lucide-svelte';
  import { toggleMode } from 'mode-watcher';
  import { addEntry } from '$lib/api';
  import Button from '$lib/components/ui/button/button.svelte';
  import { Input } from '$lib/components/ui/input';
  import { Textarea } from '$lib/components/ui/textarea';
  import { Badge } from '$lib/components/ui/badge';
  import * as Card from '$lib/components/ui/card';
  import * as Alert from '$lib/components/ui/alert';
  import * as NavigationMenu from '$lib/components/ui/navigation-menu';
  import * as Field from '$lib/components/ui/field';

  let url = $state('');
  let title = $state('');
  let text = $state('');
  let submitting = $state(false);
  let success = $state(false);
  let error = $state<string | null>(null);

  async function handleSubmit(e: Event) {
    e.preventDefault();

    if (!url.trim()) return;

    submitting = true;
    error = null;
    success = false;

    try {
      await addEntry({
        url: url.trim(),
        title: title.trim(),
        text: text.trim(),
      });

      success = true;

      // Reset form after successful submission
      setTimeout(() => {
        url = '';
        title = '';
        text = '';
        success = false;
      }, 2000);
    } catch (err) {
      console.error('Failed to add entry:', err);
      error = err instanceof Error ? err.message : 'Failed to add entry';
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>Add History Entry - Hister</title>
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
            <NavigationMenu.Link onclick={() => (window.location.href = '/rules')}>
              Rules
            </NavigationMenu.Link>
          </NavigationMenu.Item>
          <NavigationMenu.Item>
            <Badge variant="default">Add</Badge>
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
    <h1
      class="
        mb-10 font-display text-[32px] font-bold tracking-tight text-foreground
      "
    >
      Add History Entry
    </h1>

    <!-- Success Message -->
    {#if success}
      <Alert.Root
        class="
          mb-6 w-full max-w-150 border-green-500/50 bg-green-500/10
          text-green-600
        "
      >
        <Check class="size-4" />
        <Alert.Description>Entry added successfully!</Alert.Description>
      </Alert.Root>
    {/if}

    <!-- Error Message -->
    {#if error}
      <Alert.Root variant="destructive" class="mb-6 w-full max-w-150">
        <Alert.Description>{error}</Alert.Description>
      </Alert.Root>
    {/if}

    <!-- Add Form -->
    <Card.Card class="w-full max-w-248">
      <Card.CardContent class="p-10">
        <form onsubmit={handleSubmit} class="flex w-full flex-col gap-6">
          <!-- URL Field -->
          <Field.Set>
            <Field.Label class="flex items-center gap-2">
              <Link class="size-4 text-[#71717A]" />
              URL
              <span class="text-destructive">*</span>
            </Field.Label>
            <Input
              type="url"
              placeholder="https://example.com/page"
              bind:value={url}
              required
              disabled={submitting}
              class="w-full bg-background"
            />
          </Field.Set>

          <!-- Title Field -->
          <Field.Set>
            <Field.Label class="flex items-center gap-2">
              <Type class="size-4 text-[#71717A]" />
              Title
            </Field.Label>
            <Input
              type="text"
              placeholder="Page title"
              bind:value={title}
              disabled={submitting}
              class="w-full bg-background"
            />
          </Field.Set>

          <!-- Text Field -->
          <Field.Set>
            <Field.Label class="flex items-center gap-2">
              <FileText class="size-4 text-[#71717A]" />
              Text
            </Field.Label>
            <Textarea
              placeholder="Additional text or notes..."
              bind:value={text}
              rows={4}
              disabled={submitting}
              class="w-full"
            />
          </Field.Set>

          <!-- Submit Button -->
          <Button
            type="submit"
            disabled={submitting || !url.trim()}
            class="mt-2 w-full"
          >
            {#if submitting}
              <Loader class="mr-2 size-5 animate-spin" />
              Adding Entry…
            {:else}
              <Plus class="mr-2 size-5" />
              Add
            {/if}
          </Button>
        </form>
      </Card.CardContent>
    </Card.Card>
  </main>
</div>

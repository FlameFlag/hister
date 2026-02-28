<script lang="ts">
   import { onMount } from 'svelte';
   import { fetchConfig, apiFetch } from '$lib/api';
   import { Button } from '@hister/components/ui/button';
   import { Input } from '@hister/components/ui/input';
   import { Badge } from '@hister/components/ui/badge';
   import { Label } from '@hister/components/ui/label';
   import { Shield, Link2, Plus, Trash2 } from 'lucide-svelte';
   import FilterBar from '$lib/components/FilterBar.svelte';
   import StatusMessage from '$lib/components/StatusMessage.svelte';

  interface RulesData {
    skip: string[];
    priority: string[];
    aliases: Record<string, string>;
  }

  interface RuleRow {
    pattern: string;
    type: 'skip' | 'priority';
  }

  let rules: RulesData = $state({ skip: [], priority: [], aliases: {} });
  let loading = $state(true);
  let saving = $state(false);
  let message = $state('');
  let isError = $state(false);
  let newAliasKeyword = $state('');
  let newAliasValue = $state('');
  let newRulePattern = $state('');
  let newRuleType: 'skip' | 'priority' = $state('skip');

  const ruleRows = $derived.by(() => {
    const rows: RuleRow[] = [];
    for (const p of rules.skip) rows.push({ pattern: p, type: 'skip' });
    for (const p of rules.priority) rows.push({ pattern: p, type: 'priority' });
    return rows;
  });

  onMount(async () => {
    await fetchConfig();
    await loadRules();
  });

  async function loadRules() {
    loading = true;
    try {
      const res = await apiFetch('/rules', { headers: { 'Accept': 'application/json' } });
      if (!res.ok) throw new Error('Failed to load rules');
      rules = await res.json();
    } catch (e) {
      message = String(e);
      isError = true;
    } finally {
      loading = false;
    }
  }

  async function saveRules() {
    if (saving) return;
    saving = true;
    message = '';
    try {
      const formData = new URLSearchParams();
      formData.set('skip', rules.skip.join('\n'));
      formData.set('priority', rules.priority.join('\n'));
      const res = await apiFetch('/rules', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formData.toString()
      });
      if (!res.ok) throw new Error('Failed to save rules');
      message = 'Rules saved successfully';
      isError = false;
      await loadRules();
    } catch (e) {
      message = String(e);
      isError = true;
    } finally {
      saving = false;
    }
  }

  function removeRule(pattern: string, type: 'skip' | 'priority') {
    if (type === 'skip') {
      rules.skip = rules.skip.filter(p => p !== pattern);
    } else {
      rules.priority = rules.priority.filter(p => p !== pattern);
    }
    saveRules();
  }

  function addRule() {
    if (!newRulePattern.trim()) return;
    if (newRuleType === 'skip') {
      rules.skip = [...rules.skip, newRulePattern.trim()];
    } else {
      rules.priority = [...rules.priority, newRulePattern.trim()];
    }
    newRulePattern = '';
    saveRules();
  }

  async function deleteAlias(keyword: string) {
    const formData = new URLSearchParams({ alias: keyword });
    const res = await apiFetch('/delete_alias', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: formData.toString()
    });
    if (res.ok) await loadRules();
  }

  async function addAlias(e: Event) {
    e.preventDefault();
    if (!newAliasKeyword || !newAliasValue) return;
    const formData = new URLSearchParams({
      'alias-keyword': newAliasKeyword,
      'alias-value': newAliasValue
    });
    const res = await apiFetch('/add_alias', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: formData.toString()
    });
    if (res.ok) {
      newAliasKeyword = '';
      newAliasValue = '';
      await loadRules();
    }
  }
</script>

<svelte:head>
  <title>Hister Beta - Rules</title>
</svelte:head>

<div class="px-6 py-5 space-y-5 overflow-y-auto flex-1">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="space-y-1">
      <h1 class="font-outfit text-[22px] font-extrabold text-text-brand">Search Rules & Aliases</h1>
      <p class="font-inter text-sm text-text-brand-secondary">Configure how Hister indexes and searches your browsing history</p>
    </div>
    <div class="flex items-center gap-4">
      <Badge variant="outline" class="border-[2px] border-border-brand-muted bg-muted-surface text-text-brand-secondary font-inter text-xs font-semibold gap-1.5 px-3 py-1.5">
        <Shield class="size-3.5 text-hister-coral" />
        {ruleRows.length} rules
      </Badge>
      <Badge variant="outline" class="border-[2px] border-border-brand-muted bg-muted-surface text-text-brand-secondary font-inter text-xs font-semibold gap-1.5 px-3 py-1.5">
        <Link2 class="size-3.5 text-hister-indigo" />
        {Object.keys(rules.aliases).length} aliases
      </Badge>
    </div>
  </div>

  <StatusMessage
     type={isError ? 'error' : 'success'}
     message={message}
     show={!!message}
   />

  {#if loading}
    <StatusMessage type="loading" message="Loading rules..." />
  {:else}
    <!-- Search Aliases Card -->
    <div class="bg-card-surface border-[2px] border-border-brand-muted overflow-hidden">
      <!-- Indigo header -->
      <div class="flex items-center justify-between px-4 py-3 bg-hister-indigo">
        <span class="font-outfit text-lg font-extrabold text-white">Search Aliases</span>
        <span class="font-inter text-[13px] font-medium text-white/70">{Object.keys(rules.aliases).length} aliases</span>
      </div>

      <!-- Column headers -->
      <div class="flex items-center gap-4 px-4 py-2 bg-muted-surface border-b-[2px] border-border-brand-muted">
        <span class="font-inter text-xs font-bold text-text-brand-muted w-[120px] shrink-0">Keyword</span>
        <span class="font-inter text-xs font-bold text-text-brand-muted flex-1">Expands To</span>
        <span class="w-8"></span>
      </div>

      <!-- Alias rows -->
      {#each Object.entries(rules.aliases) as [keyword, value]}
        <div class="flex items-center gap-4 px-4 py-2.5 border-b-[2px] border-border-brand-muted">
          <span class="font-fira text-[13px] font-semibold text-text-brand w-[120px] shrink-0">{keyword}</span>
          <span class="font-fira text-[13px] text-text-brand-secondary flex-1 truncate">{value}</span>
          <Button
            variant="ghost"
            size="icon-sm"
            class="shrink-0 text-text-brand-muted hover:text-hister-rose"
            onclick={() => deleteAlias(keyword)}
          >
            <Trash2 class="size-4" />
          </Button>
        </div>
      {/each}

      <!-- Add alias row -->
      <form onsubmit={addAlias} class="flex items-center gap-3 px-4 py-2.5 bg-muted-surface">
        <Input
          type="text"
          bind:value={newAliasKeyword}
          placeholder="keyword..."
          class="w-[120px] h-9 px-3 bg-card-surface border-[2px] border-border-brand-muted font-fira text-xs text-text-brand shadow-none focus-visible:ring-0 focus-visible:border-hister-indigo"
        />
        <Input
          type="text"
          bind:value={newAliasValue}
          placeholder="expands to..."
          class="flex-1 h-9 px-3 bg-card-surface border-[2px] border-border-brand-muted font-fira text-xs text-text-brand shadow-none focus-visible:ring-0 focus-visible:border-hister-indigo"
        />
        <Button
          type="submit"
          size="sm"
          class="bg-hister-indigo text-white font-inter text-[13px] font-bold border-0 hover:bg-hister-indigo/90 shadow-none gap-1.5 leading-none"
        >
          <Plus class="size-3.5 shrink-0" />
          <span>Add</span>
        </Button>
      </form>
    </div>

    <!-- Indexing Rules Card -->
    <div class="bg-card-surface border-[2px] border-border-brand-muted overflow-hidden">
      <!-- Coral header -->
      <div class="flex items-center justify-between px-4 py-3 bg-hister-coral">
        <span class="font-outfit text-lg font-extrabold text-white">Indexing Rules</span>
        <span class="font-inter text-[13px] font-medium text-white/70">{ruleRows.length} rules</span>
      </div>

      <!-- Column headers -->
      <div class="flex items-center gap-4 px-4 py-2 bg-muted-surface border-b-[2px] border-border-brand-muted">
        <span class="font-inter text-xs font-bold text-text-brand-muted flex-1">Pattern</span>
        <span class="font-inter text-xs font-bold text-text-brand-muted">Type</span>
        <span class="w-8 shrink-0"></span>
      </div>

      <!-- Rule rows -->
      {#each ruleRows as row}
        <div class="flex items-center gap-4 px-4 py-2.5 border-b-[2px] border-border-brand-muted">
          <span class="font-fira text-[13px] text-text-brand flex-1 truncate">{row.pattern}</span>
          <Badge
            variant="default"
            class="text-[11px] font-bold px-2.5 py-0.5 border-0 {row.type === 'skip' ? 'bg-hister-rose text-white' : 'bg-hister-teal text-white'}"
          >
            {row.type === 'skip' ? 'Skip' : 'Priority'}
          </Badge>
          <Button
            variant="ghost"
            size="icon-sm"
            class="shrink-0 text-text-brand-muted hover:text-hister-rose"
            onclick={() => removeRule(row.pattern, row.type)}
          >
            <Trash2 class="size-4" />
          </Button>
        </div>
      {/each}

      {#if ruleRows.length === 0}
        <div class="px-4 py-4 text-center">
          <span class="font-inter text-sm text-text-brand-muted">No rules defined yet.</span>
        </div>
      {/if}

      <!-- Add rule row -->
      <div class="flex items-center gap-3 px-4 py-2.5 bg-muted-surface">
        <Input
          type="text"
          bind:value={newRulePattern}
          placeholder="Enter regex pattern..."
          class="flex-1 h-9 px-3 bg-card-surface border-[2px] border-border-brand-muted font-fira text-xs text-text-brand shadow-none focus-visible:ring-0 focus-visible:border-hister-coral"
        />
        <select
          bind:value={newRuleType}
          class="h-9 px-3 w-[100px] bg-card-surface border-[2px] border-border-brand-muted font-inter text-xs font-semibold text-text-brand outline-none cursor-pointer appearance-none text-center"
        >
          <option value="skip">Skip</option>
          <option value="priority">Priority</option>
        </select>
        <Button
          type="button"
          size="sm"
          onclick={addRule}
          class="bg-hister-coral text-white font-inter text-[13px] font-bold border-0 hover:bg-hister-coral/90 shadow-none gap-1.5 leading-none"
        >
          <Plus class="size-3.5 shrink-0" />
          <span>Add</span>
        </Button>
      </div>
    </div>
  {/if}
</div>

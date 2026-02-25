<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import {
    WebSocketManager,
    KeyHandler,
    apiRequest,
    getSearchUrl,
    exportJSON,
    exportCSV,
    exportRSS,
    formatTimestamp,
    formatRelativeTime,
    scrollTo,
    escapeHTML,
    buildSearchQuery,
    parseSearchResults,
    updateSearchURL,
    openURL
  } from '$lib/search';
  import { fetchConfig, getCsrf, setCsrf } from '$lib/api';
  import type { SearchResults } from '$lib/search';

  interface Config {
    wsUrl: string;
    csrf: string;
    searchUrl: string;
    openResultsOnNewTab: boolean;
    hotkeys: Record<string, string>;
    initialQuery: string;
  }

  let config: Config = $state({
    wsUrl: '',
    csrf: '',
    searchUrl: '',
    openResultsOnNewTab: false,
    hotkeys: {},
    initialQuery: ''
  });

  let wsManager: WebSocketManager | undefined;
  let keyHandler: KeyHandler | undefined;
  let inputEl: HTMLInputElement | undefined;
  let resultsEl: HTMLElement | undefined;

  let query = $state('');
  let autocomplete = $state('');
  let connected = $state(false);
  let lastResults: SearchResults | null = $state(null);
  let highlightIdx = $state(0);
  let currentSort = $state('');
  let dateFrom = $state('');
  let dateTo = $state('');
  let showHotkeyButton = $state(true);
  let showPopup = $state(false);
  let popupTitle = $state('');
  let popupContent = $state('');
  let showActionsForResult: string | null = $state(null);
  let actionsQuery = $state('');
  let actionsMessage: string | null = $state(null);
  let actionsError = $state(false);

  const hotkeyActions: Record<string, (e?: KeyboardEvent) => void> = {
    'open_result': openSelectedResult,
    'open_result_in_new_tab': (e) => openSelectedResult(e, true),
    'select_next_result': selectNextResult,
    'select_previous_result': selectPreviousResult,
    'open_query_in_search_engine': openQueryInSearchEngine,
    'focus_search_input': focusSearchInput,
    'view_result_popup': viewResultPopup,
    'autocomplete': autocompleteQuery,
    'show_hotkeys': showHotkeys
  };

  const hotkeyDescriptions: Record<string, string> = {
    'open_result': 'Open result',
    'open_result_in_new_tab': 'Open result in new tab',
    'select_next_result': 'Select next result',
    'select_previous_result': 'Select previous result',
    'open_query_in_search_engine': 'Open query in search engine',
    'focus_search_input': 'Focus search input',
    'view_result_popup': 'View result popup',
    'autocomplete': 'Autocomplete query',
    'show_hotkeys': 'Show Hotkeys'
  };

  const tips = [
    'Use <code>*</code> for partial match.<br />Prefixing word with <code>-</code> excludes matching documents.',
    'Click on the three dots near the result URL to specify priority queries for that result.',
    'Press <code>enter</code> to open the first result.',
    'Use <code>alt+k</code> and <code>alt+j</code> to navigate between results.',
    'Press <code>alt+o</code> to open current search query in your configured search engine.',
    'Use <code>url:</code> prefix to search only in the URL field. E.g.: <code>url:*github* hister</code>.',
    'Set hister to your default search engine in your browser to access it with ease.',
    'Start search query with <code>!!</code> to open the query in your configured search engine'
  ];

  const SORT_OPTIONS = [
    { id: '', label: 'Relevance' },
    { id: 'domain', label: 'Domain' }
  ];

  const emptyImg = 'data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw==';

  function connect() {
    wsManager = new WebSocketManager(config.wsUrl, {
      onOpen: () => {
        connected = true;
        if (query) sendQuery(query);
      },
      onMessage: renderResults,
      onClose: () => { connected = false; },
      onError: () => { connected = false; }
    });
    wsManager.connect();
  }

  function sendQuery(q: string) {
    const message = buildSearchQuery(q, currentSort, dateFrom, dateTo);
    wsManager?.send(JSON.stringify(message));
  }

  function updateURL() {
    updateSearchURL(window.location.pathname, query, dateFrom, dateTo);
  }

  function renderResults(event: MessageEvent) {
    const res = parseSearchResults(event.data);
    lastResults = res;
    autocomplete = (query && res.query_suggestion) || '';
    highlightIdx = 0;
  }

  function openUrl(u: string, newWindow?: boolean) {
    openURL(u, newWindow);
  }

  function openResult(url: string, title: string, newWindow = false) {
    if (config.openResultsOnNewTab) newWindow = true;
    saveHistoryItem(url, title, query, false, () => openUrl(url, newWindow));
  }

  function saveHistoryItem(url: string, title: string, queryStr: string, remove: boolean, callback?: (r: Response) => void) {
    apiRequest({
      url: '/history',
      params: {
        method: 'POST',
        headers: { 'Content-type': 'application/json; charset=UTF-8' },
        body: JSON.stringify({ url, title, query: queryStr, delete: remove })
      },
      csrfToken: config.csrf,
      csrfCallback: (tok) => { config.csrf = tok; setCsrf(tok); },
      callback
    });
  }

  function setSort(sortId: string) {
    if (currentSort === sortId) return;
    currentSort = sortId;
    if (query) sendQuery(query);
  }

  function deleteResult(url: string) {
    const data = new URLSearchParams({ url });
    apiRequest({
      url: '/delete',
      params: { method: 'POST', body: data },
      csrfToken: config.csrf,
      csrfCallback: (tok) => { config.csrf = tok; setCsrf(tok); },
      callback: () => {
        if (lastResults?.documents) {
          lastResults = {
            ...lastResults,
            documents: lastResults.documents.filter((d) => d.url !== url)
          };
        }
      }
    });
  }

  function updatePriorityResult(url: string, title: string, remove: boolean) {
    const q = actionsQuery || query;
    if (!q) return;
    saveHistoryItem(url, title, q, remove, (r) => {
      if (r.status === 200) {
        actionsMessage = `Priority result ${remove ? 'deleted' : 'added'}.`;
        actionsError = false;
      } else {
        actionsMessage = `Failed to ${remove ? 'delete' : 'add'} priority result.`;
        actionsError = true;
      }
    });
  }

  function openReadable(e: Event, url: string, title: string) {
    e.preventDefault();
    apiRequest({
      url: `/readable?url=${encodeURIComponent(url)}`,
      csrfToken: config.csrf,
      csrfCallback: (tok) => { config.csrf = tok; setCsrf(tok); },
      callback: (resp) => {
        resp.json().then((data) => {
          popupTitle = data.title;
          popupContent = data.content;
          showPopup = true;
        });
      }
    });
  }

  const historyLen = $derived(lastResults?.history?.length || 0);
  const docsLen = $derived(lastResults?.documents?.length || 0);
  const totalResults = $derived(historyLen + docsLen);

  function selectNthResult(n: number) {
    if (!totalResults) return;
    highlightIdx = (highlightIdx + n + totalResults) % totalResults;
    const results = document.querySelectorAll('.result');
    scrollTo(results[highlightIdx]);
  }

  function selectNextResult(e?: KeyboardEvent) {
    if (e) e.preventDefault();
    selectNthResult(1);
  }

  function selectPreviousResult(e?: KeyboardEvent) {
    if (e) e.preventDefault();
    selectNthResult(-1);
  }

  function openSelectedResult(e?: KeyboardEvent, newWindow = false) {
    if (e) e.preventDefault();
    if (query.startsWith('!!')) {
      openUrl(getSearchUrl(config.searchUrl, query.substring(2)), newWindow);
      return;
    }
    const res = document.querySelectorAll<HTMLAnchorElement>('.result .result-title a')[highlightIdx];
    if (res) {
      const url = res.getAttribute('href')!;
      const title = res.innerText;
      openResult(url, title, newWindow);
    }
  }

  function viewResultPopup(e?: KeyboardEvent) {
    if (e) e.preventDefault();
    closePopup();
    const readables = document.querySelectorAll('.result .readable');
    if (highlightIdx >= 0 && highlightIdx < readables.length && readables[highlightIdx]) {
      const readableEl = readables[highlightIdx];
      const result = readableEl.closest('.result')!;
      const link = result.querySelector<HTMLAnchorElement>('.result-title a')!;
      const url = link.getAttribute('href')!;
      const title = link.innerText;
      openReadable({ preventDefault: () => {} } as Event, url, title);
    }
  }

  function autocompleteQuery(e?: KeyboardEvent) {
    if (e) e.preventDefault();
    if (document.activeElement === inputEl && autocomplete && query !== autocomplete) {
      query = autocomplete;
      sendQuery(query);
    }
  }

  function openQueryInSearchEngine(e?: KeyboardEvent) {
    if (e) e.preventDefault();
    openUrl(getSearchUrl(config.searchUrl, query));
  }

  function focusSearchInput(e?: KeyboardEvent) {
    if (document.activeElement !== inputEl) {
      if (e) e.preventDefault();
      inputEl?.focus();
    }
  }

  function toggleHotkeyButton() {
    showHotkeyButton = !showHotkeyButton;
    localStorage.setItem('hideHotkeyButton', showHotkeyButton ? 'false' : 'true');
    closePopup();
  }

  function showHotkeys(e?: KeyboardEvent) {
    if (document.activeElement === inputEl) return;
    if (closePopup()) return;

    let hotkeysHtml = '<div class="hotkeys-list">';
    for (const k in config.hotkeys) {
      hotkeysHtml += `<div class="hotkey"><kbd>${escapeHTML(k)}</kbd><span>${hotkeyDescriptions[config.hotkeys[k]] || config.hotkeys[k]}</span></div>`;
    }
    hotkeysHtml += '</div>';

    const toggleSection = `<div class="hotkey-toggle-section"><p>The hotkey button can be toggled below. Press <kbd>?</kbd> to view this dialog.</p><button type="button" class="hotkey-toggle-btn" id="toggle-hotkey-btn">${showHotkeyButton ? 'Hide Hotkey Button' : 'Show Hotkey Button'}</button></div>`;

    popupTitle = 'Hotkeys';
    popupContent = hotkeysHtml + toggleSection;
    showPopup = true;

    // Wire up toggle button after DOM update
    setTimeout(() => {
      document.getElementById('toggle-hotkey-btn')?.addEventListener('click', toggleHotkeyButton);
    }, 0);
  }

  function closePopup(): boolean {
    if (showPopup) {
      showPopup = false;
      return true;
    }
    return false;
  }

  function handleKeydown(e: KeyboardEvent) {
    if (keyHandler?.handle(e)) {
      e.preventDefault();
      return;
    }
    if (e.key === 'Escape') {
      if (closePopup()) {
        e.preventDefault();
        return;
      }
    }
    showActionsForResult = null;
  }

  function handleButtonKeydown(e: KeyboardEvent, action: (e?: KeyboardEvent) => void) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      action(e);
    }
  }

  function getHighlightIdxForHistory(i: number) {
    return i === highlightIdx;
  }

  function getHighlightIdxForDocs(i: number) {
    return (historyLen + i) === highlightIdx;
  }

  $effect(() => {
    if (query && connected) {
      sendQuery(query);
      localStorage.setItem('lastQuery', query);
    }
  });

  $effect(() => {
    if (!query) {
      autocomplete = '';
      lastResults = null;
    }
  });

  $effect(() => {
    if (dateFrom || dateTo) sendQuery(query);
  });

  $effect(() => {
    updateURL();
  });

  $effect.pre(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const q = urlParams.get('q');
    const df = urlParams.get('date_from');
    const dt = urlParams.get('date_to');
    const lastQuery = localStorage.getItem('lastQuery');
    if (q) {
      query = q;
    }
    if (df) dateFrom = df;
    if (dt) dateTo = dt;

  });

  $effect(() => {
    inputEl?.focus();
  });

  onMount(async () => {
    const appConfig = await fetchConfig();
    config = {
      wsUrl: appConfig.wsUrl,
      csrf: getCsrf(),
      searchUrl: appConfig.searchUrl,
      openResultsOnNewTab: appConfig.openResultsOnNewTab,
      hotkeys: appConfig.hotkeys,
      initialQuery: ''
    };

    showHotkeyButton = !appConfig.hotkeys['show_hotkeys'] || localStorage.getItem('hideHotkeyButton') !== 'true';
    focusSearchInput();

    connect();
    keyHandler = new KeyHandler(config.hotkeys, hotkeyActions);

    return () => {
      wsManager?.close();
    };
  });
</script>

<svelte:window onkeydown={handleKeydown} />

{#if showPopup}
<div class="popup-wrapper" role="presentation" aria-hidden="true" onclick={(e) => { if (!(e.target as Element).closest('.popup')) closePopup(); }}>
  <div class="popup container">
    <div class="float-right">
      <!-- svelte-ignore a11y_missing_attribute -->
      <a class="popup-close" role="button" aria-label="Close" tabindex="0" onclick={closePopup} onkeydown={(e) => handleButtonKeydown(e, closePopup)}>x</a>
    </div>
    <div class="popup-header">{@html popupTitle}</div>
    <div class="popup-content">{@html popupContent}</div>
  </div>
</div>
{/if}

<div class="sticky">
  <div class="search text-center">
    <input type="text" id="search" bind:this={inputEl} bind:value={query} placeholder="Search..." />
    <input type="text" disabled id="autocomplete" value={autocomplete || ''} />
    <div id="ws-status" class="ws-status" class:connected={connected} title={connected ? 'Websocket connected' : 'Websocket disconnected'}></div>
  </div>
</div>

<details class="section" class:hidden={!lastResults?.documents?.length && !lastResults?.history?.length}>
  <summary>Actions</summary>
  <div class="container">
    <div class="sort-buttons small-grey">
      Sort by: <span class="sort-options-container">
        {#each SORT_OPTIONS as opt, i}
          <!-- svelte-ignore a11y_invalid_attribute -->
          <a class="sort-btn" class:active={currentSort === opt.id} onclick={(e) => { e.preventDefault(); setSort(opt.id); }} href="#" role="button" aria-pressed={currentSort === opt.id} tabindex="0">{opt.label}</a>{#if i < SORT_OPTIONS.length - 1}<span class="sort-separator"> | </span>{/if}
        {/each}
      </span>
    </div>
    <div class="export-buttons small-grey">
      <!-- svelte-ignore a11y_invalid_attribute -->
      Export: <a onclick={(e) => { e.preventDefault(); exportJSON(lastResults!); }} href="#" role="button" tabindex="0">JSON</a> |
      <!-- svelte-ignore a11y_invalid_attribute -->
      <a onclick={(e) => { e.preventDefault(); exportCSV(lastResults!, query); }} href="#" role="button" tabindex="0">CSV</a> |
      <!-- svelte-ignore a11y_invalid_attribute -->
      <a onclick={(e) => { e.preventDefault(); exportRSS(lastResults!, query); }} href="#" role="button" tabindex="0">RSS</a>
    </div>
    <div class="date-filters small-grey">
      Filter from: <input type="date" id="date-from" bind:value={dateFrom} placeholder="From date" title="From date" />
      Filter to: <input type="date" id="date-to" bind:value={dateTo} placeholder="To date" title="To date" />
    </div>
  </div>
</details>

{#if showHotkeyButton}
  <button type="button" id="hotkey-button" class="hotkeys-button" title="Hotkeys (?)" onclick={showHotkeys} aria-label="Show hotkeys">
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <rect x="2" y="4" width="20" height="16" rx="2"/>
      <path d="M6 8h.01M10 8h.01M14 8h.01M18 8h.01M8 12h.01M12 12h.01M16 12h.01M7 16h10"/>
    </svg>
  </button>
{/if}

<div class="container" bind:this={resultsEl} id="results">
  {#if !lastResults?.documents?.length && !lastResults?.history?.length}
    {#if !query}
      <div class="text-center">
        <h3>Tip</h3>
        <p>{@html tips[Math.floor(Math.random() * tips.length)]}</p>
      </div>
    {:else}
      <div class="result">
        <div class="result-title">
          <img src={emptyImg} alt="" />
          <a href={getSearchUrl(config.searchUrl, query)} class="error" onclick={(e) => { e.preventDefault(); openUrl(getSearchUrl(config.searchUrl, query), config.openResultsOnNewTab); }}>No results found - open query in web search engine</a>
        </div>
        <span class="result-url">{getSearchUrl(config.searchUrl, query)}</span>
      </div>
    {/if}
  {:else}
    {#if lastResults?.search_duration || lastResults?.total !== undefined}
      <div class="results-header">
        <div class="float-right">
          <div class="duration text-right">{lastResults.search_duration || ''}</div>
          <div class="search-engine-link">
            {#if query.trim()}
              <a id="external-search-link" href={getSearchUrl(config.searchUrl, query)}>Open in external search engine</a>
            {/if}
          </div>
        </div>
        <div>Total number of results: <b class="results-num">{lastResults.total || lastResults.documents?.length}</b></div>
        {#if lastResults.query && lastResults.query.text !== query}
          <div class="expanded-query">Expanded query: <code>"{escapeHTML(lastResults.query.text)}"</code></div>
        {/if}
      </div>
    {/if}

    {#if lastResults?.history?.length}
      {#each lastResults.history as r, i}
        <div class="result" class:highlight={getHighlightIdxForHistory(i)}>
          <div class="result-title">
            <img src={emptyImg} alt="" />
            <a href={r.url} class="success" onclick={(e) => { e.preventDefault(); openResult(r.url, r.title || '*title*'); }}>{@html r.title || '*title*'}</a>
          </div>
          <span class="result-url">{r.url}</span>
          <span class="action-button" role="button" tabindex="0" aria-label="Show actions"
            onclick={(e) => { e.stopPropagation(); showActionsForResult = showActionsForResult === 'history:' + r.url ? null : 'history:' + r.url; }}
            onkeydown={(e) => handleButtonKeydown(e, () => { showActionsForResult = showActionsForResult === 'history:' + r.url ? null : 'history:' + r.url; })}>
            <svg focusable="false" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
              <path fill="#95a5a6" d="M12 8c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm0 2c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2z"/>
            </svg>
          </span>
          <!-- svelte-ignore a11y_invalid_attribute -->
          <a class="readable" onclick={(e) => openReadable(e, r.url, r.title || '*title*')} href="#" role="button" tabindex="0">view</a>
          {#if showActionsForResult === 'history:' + r.url}
            <div class="actions bordered padded mt-1">
              <!-- svelte-ignore a11y_invalid_attribute -->
              <a class="close float-right" onclick={(e) => { e.stopPropagation(); showActionsForResult = null; }} href="#" role="button" tabindex="0">close</a>
              <button class="delete error" onclick={(e) => { e.stopPropagation(); updatePriorityResult(r.url, r.title || '*title*', true); }}>Delete this priority result</button>
              {#if actionsMessage}
                <p class:success={!actionsError} class:error={actionsError}>
                  <b>{actionsError ? 'Error!' : 'Success!'}</b> <span class="message">{actionsMessage}</span>
                </p>
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    {/if}

    {#if lastResults?.documents}
      {#each lastResults.documents as r, i}
        <div class="result" class:highlight={getHighlightIdxForDocs(i)}>
          <div class="result-title">
            <img src={r.favicon || emptyImg} alt="" />
            <a href={r.url} onclick={(e) => { e.preventDefault(); openResult(r.url, r.title || '*title*'); }}>{@html r.title || '*title*'}</a>
          </div>
          <span class="result-url">{r.url}</span>
          <span class="action-button" role="button" tabindex="0" aria-label="Show actions"
            onclick={(e) => { e.stopPropagation(); showActionsForResult = showActionsForResult === 'doc:' + r.url ? null : 'doc:' + r.url; }}
            onkeydown={(e) => handleButtonKeydown(e, () => { showActionsForResult = showActionsForResult === 'doc:' + r.url ? null : 'doc:' + r.url; })}>
            <svg focusable="false" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
              <path fill="#95a5a6" d="M12 8c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm0 2c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2z"/>
            </svg>
          </span>
          <span class="added" title={formatTimestamp(r.added || 0)}>{formatRelativeTime(r.added || 0)}</span>
          <!-- svelte-ignore a11y_invalid_attribute -->
          <a class="readable" onclick={(e) => openReadable(e, r.url, r.title || '*title*')} href="#" role="button" tabindex="0">view</a>
          <p class="result-content">{@html r.text || ''}</p>
          {#if showActionsForResult === 'doc:' + r.url}
            <div class="actions bordered padded mt-1">
              <!-- svelte-ignore a11y_invalid_attribute -->
              <a class="close float-right" onclick={(e) => { e.stopPropagation(); showActionsForResult = null; }} href="#" role="button" tabindex="0">close</a>
              Prioritize this result for the following query:<br />
              <input type="text" class="action-query" bind:value={actionsQuery} placeholder="Query.." />
              <button class="save" onclick={(e) => { e.stopPropagation(); updatePriorityResult(r.url, r.title || '*title*', false); }}>Save</button><br />
              <button class="delete error" onclick={(e) => { e.stopPropagation(); deleteResult(r.url); }}>Delete this result</button>
              {#if actionsMessage}
                <p class:success={!actionsError} class:error={actionsError}>
                  <b>{actionsError ? 'Error!' : 'Success!'}</b> <span class="message">{actionsMessage}</span>
                </p>
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    {/if}
  {/if}
</div>

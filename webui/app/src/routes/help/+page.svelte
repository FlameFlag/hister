<script lang="ts">
  import { Kbd } from '@hister/components/ui/kbd';
</script>

<svelte:head>
  <title>Hister - Help</title>
</svelte:head>

<div class="container">
  <h2>Search Shortcuts</h2>
  <p>Press <Kbd>enter</Kbd> to open the first result. Alternatively press <Kbd>alt+enter</Kbd> to open in new tab.</p>
  <p>Navigate in results with <Kbd>alt+j</Kbd> and <Kbd>alt+k</Kbd>. <Kbd>Enter</Kbd>/<Kbd>alt+enter</Kbd> opens the selected result.</p>
  <p>Press <Kbd>alt+o</Kbd> to open the search query in the configured search engine.</p>
  <h2>Search Syntax</h2>
  <p>Use <Kbd>quotes</Kbd> to match phrases.</p>
  <p>Use <Kbd>*</Kbd> for wildcard matches.</p>
  <p>Prefix words or phrases with <Kbd>-</Kbd> to exclude matching documents.</p>
  <p>Use <code>url:</code> prefix to search only in the URL field.</p>
  <h3>Examples</h3>
  <p><code>"free software" url:*wikipedia.org*</code>: Search for the phrase "free software" only in URLs containing wikipedia.org.</p>
  <p><code>golang template -url:*stackoverflow*</code>: Search sites containing both "golang" and "template" but the website's URL should not contain "stackoverflow".</p>
  <h2>Search Aliases</h2>
  <p>Queries can become long and complex quickly. Aliases can be defined in the <a href="rules">rules</a> page to shorten common query parts.</p>
  <h3>Examples</h3>
  <p><code>go</code> alias for <code>(go|golang)</code> matches either "go" or "golang" if you type "go".</p>
  <p><code>!so</code> alias for <code>url:*stackoverflow.com*</code> matches only sites where the URL contains "stackoverflow.com".</p>
</div>

{{define "main"}}
<div class="container">
<h2>Search Shortcuts</h2>
<p>Press <kbd>enter</kbd> to open the first result.</p>
<p>Navigate in results with <kbd>ctrl+j</kbd> and <kbd>ctrl+k</kbd>. <kbd>Enter</kbd> opens the selected result.</p>
<p>Press <kbd>ctrl+o</kbd> to open the search query in the configured search engine.</p>
<h2>Search Syntax</h2>
<p>Use <kbd>quotes</kbd> to match the whole phrase.</p>
<p>Use <kbd>*</kbd> for wildcard matches</p>
<p>Make phrases mandatory with <kbd>+</kbd> prefix.</p>
<p>Prefix words with <kbd>-</kbd> to exclude matching documents.</p>
<p>Use <code>url:</code> prefix to search only in the URL field</p>
<h3>Examples</h3>
<p><code>"free software" +url:*wikipedia.org*</code>: Search for the phrase "free software" only in URLs containing wikipedia.org</p>
<p><code>+golang +template -"stack overflow"</code>: Search sites containing both "golang" and "template" words but not the phrase "stack overflow"</p>
</div>
{{end}}

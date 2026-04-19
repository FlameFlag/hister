Delete documents matching a query. Same DSL as `search.text`. `dry_run` defaults to TRUE for safety: a dry run returns the count and a sample (up to 20 URLs) without deleting.

To actually delete, call again with `dry_run:false`. The natural flow is `search` (or `forget dry_run:true`) to preview -> `forget dry_run:false` to commit.

If the client supports elicitation, the server prompts the user to confirm the delete in-band when `dry_run:false` is called without `confirm_count`; `result.action` is `"deleted"` on accept, `"cancelled"` on decline.

`confirm_count` is optimistic concurrency for clients without elicitation: if the index shifts between your preview and your commit (e.g. a concurrent indexer adds or removes matching docs), passing `confirm_count` makes the delete abort when the count no longer matches, instead of silently deleting an unexpected number of documents. Pass the number you saw during preview. When set, it also skips the elicitation prompt.

Search the user's local browsing-history index ("what do I know about X?"). Returns ranked hits with snippets, plus facets (top domains, languages, date buckets) that summarise coverage of the result set.

The `text` argument accepts a rich query DSL:

```
cats                         plain terms; fields: url(w4) domain(w8) title(w12) text(w1)
"exact phrase"               phrase match
domain:example.com           field lookup
title:(cat|dog|bird)         alternation
url:*.pdf                    wildcard
-domain:ads.example.com      negation
language:en title:cat        combine with whitespace
```

Pass `page_key` from the previous response for pagination. Snippets are capped; use `fetch_document` when you need full text or raw HTML. Zero hits means the query didn't match; try broader terms or drop a field qualifier.

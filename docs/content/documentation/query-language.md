+++
date = '2026-02-25T14:40:00+01:00'
draft = false
title = 'Query Language Guide'
layout = 'documentation'
+++

Hister provides a query language that allows you to search through your content with precision. This guide explains the different query types and search techniques available.

## Basic Search

Simply type any word to search across all fields:

```
hister
```

This searches for "hister" in the title, text content, URL, and domain of all indexed pages.

## Quoted Phrases

Use double quotes to search for exact phrases:

```
"privacy policy"
```

This finds pages containing the exact phrase "privacy policy" (not just pages with both words separately).

**Examples:**
```
"self-hosted applications"
"data privacy regulations"
"end-to-end encryption"
```

## Field-Specific Searches

You can search within specific fields using the `field:value` syntax:

### Available Fields

- **title:** - Search in page titles only
- **text:** - Search in page content only
- **url:** - Search in URLs only
- **domain:** - Search in domain names only

**Examples:**

```
title:encryption
```
Finds pages with "encryption" in the title.

```
domain:github.com
```
Finds all pages from github.com.

```
url:*/security/*
```
Finds pages with "security" in the URL path.

```
text:"GDPR compliance"
```
Finds pages with "GDPR compliance" in the body text.

### Privacy & Security Examples

```
title:privacy domain:mozilla.org
title:"security audit" text:vulnerability
url:*/privacy-policy
domain:privacyguides.org text:encryption
```

## Wildcard Searches

Use asterisks (`*`) for wildcard matching:

```
secur*
```
Matches: security, secure, securing, etc.

```
*privacy*
```
Matches: privacy, myprivacy, privacytools, etc.

**Field-specific wildcards:**
```
domain:*.github.io
url:*/docs/*
title:*firewall*
```

## Negation

Prefix terms with a minus sign (`-`) to exclude results:

```
privacy -facebook
```
Finds pages about privacy but excludes results containing "facebook".

```
encryption title:-tutorial
```
Finds pages about encryption but not those with "tutorial" in the title.

**Field-specific negation:**
```
security domain:-facebook.com
title:hister url:-*/issues/*
privacy -"social media"
```

## Alternation Expressions

Use parentheses with pipe (`|`) to create OR conditions:

```
(security|privacy|encryption)
```
Finds pages containing any of these terms.

```
title:(firewall|vpn|proxy)
```
Finds pages with firewall, VPN, or proxy in the title.

### Combining with Other Queries

```
"data protection" (GDPR|CCPA|HIPAA)
```
Finds pages about data protection mentioning any of these regulations.

```
domain:(github.com|gitlab.com) title:security
```
Finds security-related pages from GitHub or GitLab.

## Combining Query Types

You can combine all query types for powerful searches:

```
title:encryption "end-to-end" domain:(signal.org|whatsapp.com) -deprecated
```

This finds pages where:
- "encryption" appears in the title
- Contains the phrase "end-to-end"
- From signal.org OR whatsapp.com domains
- Does NOT contain "deprecated"

### Real-World Examples

**Finding privacy tools:**
```
(privacy|security) tools "open source" -commercial
```

**Research on specific topics:**
```
"threat model" (encryption|authentication|authorization) -tutorial
```

**Documentation searches:**
```
title:(setup|installation|configuration) domain:(*.io|*.dev) hister
```

**Security vulnerabilities:**
```
(CVE|vulnerability|exploit) (2024|2025|2026) -"not affected"
```

**Self-hosting resources:**
```
"self-hosted" (docker|kubernetes|compose) title:(guide|tutorial)
```

## Tips

### 1. Case Insensitivity

All searches are case-insensitive:
```
Privacy = privacy = PRIVACY
```

### 2. Wildcards and Performance

- Leading wildcards (`*term`) are slower than trailing wildcards (`term*`)
- Starting query with `*` immedately tries to find every documentum, that can lead to performance issues
- Use field-specific wildcards when possible for better performance

### 3. Empty Alternations

Alternations must contain at least one option:
```
()           # Invalid
(a)          # Valid - single option
(a|b)        # Valid - multiple options
```

## Query Best Practices

### Start Broad, Then Narrow

```
# Start with:
encryption

# Refine to:
encryption privacy

# Further refine:
"end-to-end encryption" (signal|matrix) -deprecated
```

### Use Field Searches for Precision

Instead of:
```
github security issue
```

Try:
```
domain:github.com title:(security|vulnerability) -closed
```

### Combine Phrases with Alternations

```
"privacy policy" (updated|changed|revised) (2025|2026)
```

## Common Use Cases

### Finding Documentation
```
title:(docs|documentation|guide) domain:*.io hister
```

### Research Topic
```
"zero-knowledge" (encryption|proof|architecture) -marketing
```

### Tracking Updates
```
domain:mozilla.org (firefox|thunderbird) "release notes"
```

### Security News
```
(vulnerability|CVE|security) "disclosure" -duplicate
```

### Privacy Tools Comparison
```
"privacy" (comparison|vs|versus) (browser|vpn|email)
```

## Troubleshooting Queries

### No Results Found

- Remove field restrictions and try a broader search
- Check spelling and try wildcards
- Remove negations to see what's being excluded
- Simplify alternations

### Too Many Results

- Add field-specific searches
- Use quoted phrases for exact matches
- Add negations to filter out unwanted content
- Specify domains to narrow scope

### Unexpected Results

- Ensure quotes are properly closed
- Check that parentheses are balanced
- Verify field names are spelled correctly (title, text, url, domain)
- Remember searches are case-insensitive

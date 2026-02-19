export function sanitizeHtml(html: string): string {
  if (!html) return '';

  const allowedTags = ['mark', 'b', 'strong', 'em', 'i'];

  const temp = document.createElement('div');
  temp.textContent = html; // First escape everything
  const escaped = temp.innerHTML;

  let result = escaped;
  for (const tag of allowedTags) {
    const openRegex = new RegExp(`&lt;${tag}&gt;`, 'gi');
    const closeRegex = new RegExp(`&lt;/${tag}&gt;`, 'gi');
    result = result.replace(openRegex, `<${tag}>`);
    result = result.replace(closeRegex, `</${tag}>`);
  }

  return result;
}

export function stripHtml(html: string): string {
  if (!html) return '';
  const tmp = document.createElement('div');
  tmp.innerHTML = html;
  return tmp.textContent || tmp.innerText || '';
}

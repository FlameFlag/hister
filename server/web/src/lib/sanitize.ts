export function sanitizeHtml(html: string): string {
  if (!html) return '';

  const temp = document.createElement('div');
  temp.textContent = html;
  const escaped = temp.innerHTML;

  return ['mark', 'b', 'strong', 'em', 'i'].reduce((result, tag) => {
    const openRegex = new RegExp(`&lt;${tag}&gt;`, 'gi');
    const closeRegex = new RegExp(`&lt;/${tag}&gt;`, 'gi');
    return result.replace(closeRegex, `</${tag}>`).replace(openRegex, `<${tag}>`);
  }, escaped);
}

export function stripHtml(html: string): string {
  if (!html) return '';
  const tmp = document.createElement('div');
  tmp.innerHTML = html;
  return tmp.textContent || tmp.innerText || '';
}

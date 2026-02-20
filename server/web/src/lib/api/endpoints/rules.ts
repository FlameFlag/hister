import ky from 'ky';
import { fetchWithRetry, fetchWithCSRFAndFormData } from '../client';
import { parse } from 'valibot';
import { rulesSchema, type Rules } from './schemas';

export async function saveRules(rules: Rules): Promise<void> {
  const validated = parse(rulesSchema, rules);
  const csrf = await (await import('../client')).ensureCSRFToken();

  const response = await ky.post('/api/rules', {
    json: validated,
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrf,
    },
  });

  if (!response.ok) {
    throw new Error(`Failed to save rules: ${response.statusText}`);
  }

  (await import('../client')).updateCSRFToken(response.headers.get('X-CSRF-Token') || null);
}

export async function fetchRules(): Promise<Rules> {
  const response = await fetchWithRetry<Rules>('/api/rules');
  return parse(rulesSchema, response);
}

export async function addAlias(keyword: string, value: string): Promise<void> {
  const formData = new FormData();
  formData.set('alias-keyword', keyword);
  formData.set('alias-value', value);
  await fetchWithCSRFAndFormData('/add_alias', formData);
}

export async function deleteAlias(alias: string): Promise<void> {
  const formData = new FormData();
  formData.set('alias', alias);
  await fetchWithCSRFAndFormData('/add_alias', formData);
}

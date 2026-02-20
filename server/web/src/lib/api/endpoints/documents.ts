import { fetchWithCSRF, fetchWithCSRFAndFormData } from '../client';

export async function addEntry(url: string, title: string, text: string): Promise<void> {
  await fetchWithCSRF<void>('/add', {
    json: { url, title, text },
  });
}

export async function deleteDocument(url: string): Promise<void> {
  const formData = new FormData();
  formData.set('url', url);
  await fetchWithCSRFAndFormData('/delete', formData);
}

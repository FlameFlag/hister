import { error } from '@sveltejs/kit';

export const prerender = true;

const modules = import.meta.glob('../../../../docs/*.md', { eager: true });

const docs = Object.fromEntries(
  Object.entries(modules).map(([path, mod]) => {
    const slug = path.split('/').pop()?.replace('.md', '') ?? path;
    return [slug, mod];
  })
) as Record<string, { default: unknown; metadata?: Record<string, unknown> }>;

export const entries = () => Object.keys(docs).map(slug => ({ slug }));

export async function load({ params }: { params: { slug: string } }) {
  const post = docs[params.slug];
  if (!post) {
    error(404, `Documentation page "${params.slug}" not found`);
  }
  return {
    content: post.default,
    meta: post.metadata ?? {}
  };
}

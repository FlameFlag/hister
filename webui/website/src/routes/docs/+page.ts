const modules = import.meta.glob('../../../docs/*.md', { eager: true });

export async function load() {
  const docs = Object.entries(modules).map(([path, mod]) => {
    const slug = path.split('/').pop()?.replace('.md', '') ?? path;
    const { metadata } = mod as { metadata?: { title?: string } };
    return {
      slug,
      title: metadata?.title ?? slug.replace(/-/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
    };
  });

  return { docs };
}
const modules = import.meta.glob('../../content/docs/*.md', { eager: true });

export interface DocEntry {
  slug: string;
  title: string;
  order: number;
  category: string;
}

export interface DocCategory {
  name: string;
  docs: DocEntry[];
}

export async function load() {
  const docs: DocEntry[] = Object.entries(modules)
    .map(([path, mod]) => {
      const slug = path.split('/').pop()?.replace('.md', '') ?? path;
      const { metadata } = mod as { metadata?: Record<string, string | number> };
      return {
        slug,
        title: (metadata?.title as string) ?? slug.replace(/-/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase()),
        order: (metadata?.order as number) ?? 99,
        category: (metadata?.category as string) ?? 'Other'
      };
    })
    .sort((a, b) => a.order - b.order);

  const categoryOrder = ['Getting Started', 'Reference', 'Deployment'];
  const categoryMap = new Map<string, DocEntry[]>();
  for (const doc of docs) {
    if (!categoryMap.has(doc.category)) categoryMap.set(doc.category, []);
    categoryMap.get(doc.category)!.push(doc);
  }

  const categories: DocCategory[] = categoryOrder
    .filter((name) => categoryMap.has(name))
    .map((name) => ({ name, docs: categoryMap.get(name)! }));

  return { docs, categories };
}

import {
  object,
  string,
  array,
  number,
  record,
  optional,
  pipe,
  url as urlValidator,
  minLength,
  minValue,
  type InferOutput,
} from 'valibot';

export const documentSchema = object({
  url: pipe(string(), urlValidator()),
  domain: string(),
  title: string(),
  text: string(),
  favicon: string(),
  score: number(),
  added: number(),
});

export type Document = InferOutput<typeof documentSchema>;

export const urlCountSchema = object({
  url: string(),
  title: string(),
  count: pipe(number(), minValue(0)),
});

export type URLCount = InferOutput<typeof urlCountSchema>;

export const searchQuerySchema = object({
  text: pipe(string(), minLength(1)),
  highlight: optional(string()),
  fields: optional(array(string())),
  limit: optional(pipe(number(), minValue(1))),
  sort: optional(string()),
  date_from: optional(number()),
  date_to: optional(number()),
});

export type SearchQuery = InferOutput<typeof searchQuerySchema>;

export const searchResultsSchema = object({
  total: pipe(number(), minValue(0)),
  query: searchQuerySchema,
  documents: array(documentSchema),
  history: array(urlCountSchema),
  search_duration: string(),
  query_suggestion: string(),
});

export type SearchResults = InferOutput<typeof searchResultsSchema>;

export const rulesSchema = object({
  skip: array(string()),
  priority: array(string()),
  aliases: record(string(), string()),
});

export type Rules = InferOutput<typeof rulesSchema>;

export const historyItemSchema = object({
  query: string(),
  title: string(),
  url: string(),
  favicon: optional(string()),
});

export type HistoryItem = InferOutput<typeof historyItemSchema>;

export const addEntryRequestSchema = object({
  url: pipe(string(), urlValidator()),
  title: string(),
  text: string(),
});

export type AddEntryRequest = InferOutput<typeof addEntryRequestSchema>;

export const statsSchema = object({
  pagesIndexed: pipe(number(), minValue(0)),
  domains: pipe(number(), minValue(0)),
  dateRange: string(),
  minDate: optional(number()),
  maxDate: optional(number()),
});

export type Stats = InferOutput<typeof statsSchema>;

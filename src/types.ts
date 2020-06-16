import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface HarperDBQuery extends DataQuery {
  queryText?: string;
  constant: number;
  options?: HarperDBOptions;
}

export const defaultQuery: Partial<HarperDBQuery> = {
  constant: 6.5,
};

/**
 * These are options configured for each DataSource instance
 */
export interface HarperDBOptions extends DataSourceJsonData {
  url?: string;
  schema?: string;
  table?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface HarperDBJsonData {
  apiKey?: string;
}

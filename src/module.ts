import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './DataSource';
import { ConfigEditor } from './ConfigEditor';
import { QueryEditor } from './QueryEditor';
import { HarperDBQuery, HarperDBOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, HarperDBQuery, HarperDBOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);

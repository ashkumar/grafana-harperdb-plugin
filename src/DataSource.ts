import { DataSourceInstanceSettings, DataQueryRequest, DataQueryResponse } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { HarperDBOptions, HarperDBQuery } from './types';
import { Observable } from 'rxjs';

export class DataSource extends DataSourceWithBackend<HarperDBQuery, HarperDBOptions> {
  iSetting: DataSourceInstanceSettings<HarperDBOptions>;
  constructor(instanceSettings: DataSourceInstanceSettings<HarperDBOptions>) {
    super(instanceSettings);
    this.iSetting = instanceSettings;
  }

  query(request: DataQueryRequest<HarperDBQuery>): Observable<DataQueryResponse> {
    // Why? This is needed otherwise options never get passed into the query (ex: v.bucket)
    console.log(this.iSetting.jsonData);
    request.targets.forEach((target: HarperDBQuery, i: number) => {
      request.targets[i] = {
        ...target,
        options: this.iSetting.jsonData,
      };
    });

    return super.query.call(this, request);
  }
}

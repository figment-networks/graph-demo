// These functions hook into the Go V8 runtime
export function callGQL<Response>(graphName: Network,  query: string, arr: object): GraphQLResponse<Response>;
export function storeRecord(type: string,  record: object): any;
export function logInfo(msg: string);

export enum Network {
    COSMOS = 'cosmos',
}

export interface GraphQLResponse<T> {
    error: any;
    data: T;
}

export interface BlockResponse {
    id: string;
    height: number;
    time: Date;
}

export interface NewBlockEvent {
    height: number;
    network: Network;
}

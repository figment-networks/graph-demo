// These functions hook into the Go V8 runtime
export function callGQL(graphName: Network, query: string, args: object, version?: string): string;
export function storeRecord(type: string,  record: object): any;
export function printA(msg: string);

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

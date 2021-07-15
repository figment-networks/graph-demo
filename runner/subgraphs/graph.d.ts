// These functions hook into the jsRuntime. The names and params must be changed in both places.
export function call(identifier: GraphQLSourceIdentifier, query: string, variables: object, version?: string): GraphQLResponse;
export function storeRecord(type: string,  record: object): any;
export function printA(msg: string);

export enum Network {
    COSMOS = 'cosmos',
}

type SubgraphId = string;
type Self = 'Self';

export type GraphQLSourceIdentifier = Network | SubgraphId | Self;

export interface GraphQLResponse {
    error: Error;
    data: any;
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

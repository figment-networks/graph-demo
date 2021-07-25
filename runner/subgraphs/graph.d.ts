// These functions hook into the jsRuntime. The names and params must be changed in both places.
export declare namespace store {
    export function save(type: string,  record: object): any;
}

export declare namespace log {
    export function debug(msg: string);
}

export declare namespace graphql {
    export function call(identifier: GraphQLSourceIdentifier, query: string, variables: object, version?: string): GraphQLResponse;
}

export enum Network {
    COSMOS = 'cosmos',
}
// type SubgraphId = string;
// type Self = 'Self';
export type GraphQLSourceIdentifier = Network; // | SubgraphId | Self;

export interface GraphQLResponse {
    error: Error;
    data: any;
}

export interface BlockResponse {
    id: string;
    height: number;
    time: Date;
}

export interface BlockEvent {
    height: number;
}

export interface TransactionEvent {
    height: number;
}

export function callGQL<Response>(graphName: Network,  query: string, arr: object): GraphQLResponse<Response>;
export function storeRecord(type: string,  record: object): any;
export function printA(a: any);

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

export interface SubgraphNewBlock {
    height: number;
    network: Network;
}

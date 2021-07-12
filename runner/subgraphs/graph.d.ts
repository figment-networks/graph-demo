export function callGQL<ResponseType>(graphName: Network,  query: string, arr: object): GraphQLResponse<ResponseType>;
export function storeRecord(type: string,  record: object): any;
export function printA(a: any);

export enum Network {
    COSMOS = 'cosmos',
}

export interface GraphQLResponse<T> {
    error: any;
    data: T;
}

export interface Block {
    id: string;
    height: number;
    time: Date;
}

export interface SubgraphNewBlock {
    height: number;
    network: Network;
}

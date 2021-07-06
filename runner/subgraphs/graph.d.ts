export function callGQL(subgraph: string,  query: string, arr: object): any;
export function storeRecord(type: string,  record: object): any;
export function printA(a: any);



export interface GrapQLResponse {
    error: any;
    data: any;
}

export interface GetBlockResponse {
    id: string;
    height: number;
    time: Date;
}

export interface SubgraphNewBlock {
    height: number;
}



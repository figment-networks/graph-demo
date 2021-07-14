package ws

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/figment-networks/graph-demo/manager/client"
	"github.com/figment-networks/graph-demo/manager/conn"
	"github.com/figment-networks/graph-demo/manager/structs"
)

// Connector is main WS connector for manager
type Connector struct {
	cli client.ControllContractor
}

// NewConnector is  Connector constructor
func NewConnector(cli client.ControllContractor) *Connector {
	return &Connector{cli}
}

func (c *Connector) Handler(connMux conn.WsConn) {
	connMux.Attach("last_data", c.lastData)
	connMux.Attach("sync_range", c.syncRange)
}

func (c *Connector) lastData(ctx context.Context, req conn.Request, response conn.Response) {
	arg := req.Arguments()
	if len(arg) == 0 {
		response.Send(nil, errors.New("not enough arguments "))
		return
	}

	ldrM, ok := arg[0].(map[string]interface{})
	if !ok {
		response.Send(nil, errors.New("bad Request "))
		return
	}

	ldr := &structs.LatestDataRequest{}
	ldr.FromMapStringInterface(ldrM)
	ldResp, err := c.cli.LatestData(ctx, *ldr)
	if err != nil {
		response.Send(nil, err)
		return
	}
	enc := json.NewEncoder(response)

	if err := enc.Encode(ldResp); err != nil {
		response.Send(nil, err)
	}
}

func (c *Connector) syncRange(ctx context.Context, req conn.Request, response conn.Response) {
	arg := req.Arguments()
	if len(arg) == 0 {
		response.Send(nil, errors.New("not enough arguments "))
		return
	}

	ldrM, ok := arg[0].(map[string]interface{})
	if !ok {
		response.Send(nil, errors.New("bad Request "))
		return
	}

	sdr := &structs.SyncDataRequest{}
	sdr.FromMapStringInterface(ldrM)
	ldResp, err := c.cli.SyncData(ctx, *sdr)
	if err != nil {
		response.Send(nil, err)
		return
	}
	enc := json.NewEncoder(response)

	if err := enc.Encode(ldResp); err != nil {
		response.Send(nil, err)
	}
}

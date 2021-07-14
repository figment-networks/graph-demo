package grpc

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/figment-networks/graph-demo/manager/connectivity/structs"
	"github.com/figment-networks/graph-demo/manager/transport/grpc/indexer"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ConnectGRPCRunner  GRPC Runner interface
type ConnectGRPCRunner interface {
	Connect(ctx context.Context, stream *structs.StreamAccess) error
	Run(ctx context.Context, logger *zap.Logger, stream *structs.StreamAccess)
}

// AwaitingPing is basic iinformation about sent ping's
type AwaitingPing struct {
	t    time.Time
	resp chan structs.ClientResponse
	live bool
}

// Client is GRPC implementation for connectivity
type Client struct {
	GrpcMaxRecvSize int
	GrpcMaxSendSize int
}

// NewClient Client constructor
func NewClient(maxGrpcReceiveSize, maxGrpcSendSize int) *Client {
	return &Client{
		GrpcMaxRecvSize: maxGrpcReceiveSize,
		GrpcMaxSendSize: maxGrpcSendSize,
	}
}

// Type returns client type
func (c *Client) Type() string {
	return "grpc"
}

// Run runs grpc client, connecting to given address and maintaining connection
func (c *Client) Run(ctx context.Context, logger *zap.Logger, stream *structs.StreamAccess) {
	defer stream.Close()
	id, _ := uuid.NewRandom()

	logger.Info("[GRPC] Trying to connect with ", zap.Stringer("id", id), zap.Any("connection_info", stream.WorkerInfo.ConnectionInfo))
	var conn *grpc.ClientConn
	var errSet []error
	var dialErr error
	var connectedTo string
	// (lukanus): try every possibility
	for _, ci := range stream.WorkerInfo.ConnectionInfo {
		for _, address := range ci.Addresses {
			if address.Address != "" {
				connectedTo = address.Address
				conn, dialErr = grpc.Dial(address.Address, grpc.WithInsecure(),
					grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(c.GrpcMaxRecvSize)),
					grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(c.GrpcMaxSendSize)),
				)
			} else {
				connectedTo = address.IP.String()
				conn, dialErr = grpc.Dial(address.IP.String(), grpc.WithInsecure(),
					grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(c.GrpcMaxRecvSize)),
					grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(c.GrpcMaxSendSize)),
				)
			}

			if dialErr != nil {
				errSet = append(errSet, dialErr)
			}

			if dialErr == nil && conn != nil {
				break
			}
		}
	}

	if conn == nil || len(errSet) > 0 {
		logger.Error(fmt.Sprintf("Cannot connect to any address given by worker: %s ", id.String()), zap.Errors("errors", errSet))
		return
	}
	defer conn.Close()

	gIndexer := indexer.NewIndexerServiceClient(conn)
	taskStream, err := gIndexer.TaskRPC(ctx, grpc.WaitForReady(true))
	if err != nil {
		logger.Error(fmt.Sprintf("Error during connecting to worker: %s", id.String()), zap.Any("worker_info", stream.WorkerInfo))
		return
	}

	logger.Info("[GRPC] Successfully Dialed ", zap.Stringer("id", id), zap.String("address", connectedTo))

	receiverClosed := make(chan error)

	pingCh := make(chan structs.TaskResponse, 20)
	defer close(pingCh)

	pings := make(map[uuid.UUID]AwaitingPing)

	go Recv(id, logger, taskStream, stream, pingCh, receiverClosed)

ControlRPC:
	for {
		select {
		case cT := <-stream.ClientControl:
			if cT.Type == "PING" {
				u, _ := uuid.NewRandom()
				pings[u] = AwaitingPing{time.Now(), cT.Resp, true}
				taskStream.Send(&indexer.TaskRequest{
					Id:   u.String(),
					From: stream.ManagerID,
					Type: "PING",
				})
			} else if cT.Type == "CLOSE" {
				taskStream.CloseSend()
				cT.Resp <- structs.ClientResponse{OK: true}
				close(cT.Resp)
				return
			} else {
				logger.Error("[GRPC] Unknown ControlType", zap.Stringer("id", id), zap.Any("control_type", cT))
			}
		case resp := <-pingCh:
			p, ok := pings[resp.ID]
			if !ok {
				logger.Error("[GRPC] Unknown PONG message", zap.Stringer("id", id), zap.Any("message", p))
				continue
			}
			if p.resp != nil {
				logger.Debug("[GRPC] Sending ClientResponse PONG", zap.Duration("pong_duration", time.Since(p.t)))
				p.resp <- structs.ClientResponse{OK: true, Time: time.Since(p.t)}
			}
			delete(pings, resp.ID)
		case <-ctx.Done():
			logger.Debug("[GRPC] Context Done", zap.Stringer("id", id))
			break ControlRPC
		case <-receiverClosed:
			logger.Debug("[GRPC] ReceiverClosed", zap.Stringer("id", id))
			break ControlRPC
		case req := <-stream.RequestListener:
			if err := taskStream.Send(&indexer.TaskRequest{
				Id:      req.ID.String(),
				From:    stream.ManagerID,
				Type:    req.Type,
				Payload: req.Payload,
			}); err != nil {
				if err == io.EOF {
					logger.Debug("[GRPC] Stream io.EOF", zap.Stringer("id", id))
					break ControlRPC
				}

				logger.Error("[GRPC] Error sending TaskResponse", zap.Stringer("id", id), zap.Error(err))
			}
		}

	}

	if taskStream != nil {
		taskStream.CloseSend()
	}

	for _, p := range pings {
		close(p.resp)
	}

	logger.Info("[GRPC] Finished", zap.Stringer("id", id), zap.String("address", connectedTo))
}

// Recv listener for stream messages
func Recv(id uuid.UUID, logger *zap.Logger, stream indexer.IndexerService_TaskRPCClient, internalStream *structs.StreamAccess, pingCh chan<- structs.TaskResponse, finishCh chan error) {
	defer close(finishCh)
	for {
		in, err := stream.Recv()
		if err != nil {
			if err == io.EOF { // (lukanus): Done reading/end of stream
				logger.Debug("[GRPC] Receive: Finished - io.EOF", zap.Stringer("id", id))
				return
			}
			logger.Error("[GRPC] Receive: error", zap.Stringer("id", id), zap.Error(err))
			finishCh <- fmt.Errorf("Receive error %w", err)
			return
		}

		msgID, err := uuid.Parse(in.Id)
		if err != nil {
			logger.Error("[GRPC] Receive: cannot parse incomming uuid", zap.Stringer("id", id), zap.String("uuid", in.Id), zap.Error(err))
			continue
		}

		if in.Type == "PONG" {
			select { // (lukanus): Never stuck
			case pingCh <- structs.TaskResponse{ID: msgID, Type: in.Type}:
			default:
				logger.Error("[GRPC] Receive: PONG cannot send response", zap.Stringer("id", id), zap.Any("request", in))
			}
			continue
		}

		resp := &structs.TaskResponse{
			ID:      msgID,
			Order:   in.Order,
			Final:   in.Final,
			Type:    in.Type,
			Payload: in.Payload,
		}

		if in.Error != nil {
			resp.Error = structs.TaskError{
				Msg:  in.Error.Msg,
				Type: structs.TaskErrorType(in.Error.Type),
			}
		}

		if err = internalStream.Recv(resp); err != nil {
			logger.Error("[GRPC] Receive: Error passing response", zap.Stringer("id", id), zap.Any("request", in), zap.Error(err))
		}
	}

}

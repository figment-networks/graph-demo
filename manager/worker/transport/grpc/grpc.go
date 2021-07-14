// Package GRPC incorporate GRPC interface of worker allows it to accept new streams and forward messages
package grpc

import (
	"context"
	"io"

	"github.com/figment-networks/graph-demo/manager/worker/transport/grpc/indexer"
	"github.com/google/uuid"
	"go.uber.org/zap"

	cStructs "github.com/figment-networks/graph-demo/manager/worker/connectivity/structs"
)

// IndexerServer is implemantation of GRPC IndexerServiceServer
type IndexerServer struct {
	indexer.UnimplementedIndexerServiceServer
	ctx    context.Context
	client cStructs.IndexerClienter
	logger *zap.Logger
}

// NewIndexerServer if IndexerServer constructor
func NewIndexerServer(ctx context.Context, client cStructs.IndexerClienter, logger *zap.Logger) *IndexerServer {
	return &IndexerServer{
		client: client,
		logger: logger,
		ctx:    ctx,
	}
}

// TaskRPC is fullfilment of TaskRPC endpoint from grpc.
// it receives new stream requests and creates the second stream, afterwards just controls incoming messages
func (is *IndexerServer) TaskRPC(taskStream indexer.IndexerService_TaskRPCServer) error {
	ctx := taskStream.Context()
	receiverClosed := make(chan error)
	defer close(receiverClosed)
	defer is.logger.Sync()

	var after bool
	var stream *cStructs.StreamAccess
	var streamID uuid.UUID
	for {
		select {
		case <-ctx.Done():
			is.logger.Debug("[GRPC] TaskRPC CONTEXT DONE")
			is.client.CloseStream(ctx, streamID)
			return ctx.Err()
		default:
		}

		in, err := taskStream.Recv()
		if err == io.EOF { // (lukanus): Done reading/end of stream
			is.logger.Debug("[GRPC] Indexer stream io.EOF")
			is.client.CloseStream(ctx, streamID)
			return nil
		}

		if err != nil {
			receiverClosed <- err
			is.client.CloseStream(ctx, streamID)
			return err
		}

		if !after {
			stream = cStructs.NewStreamAccess()
			streamID = stream.StreamID
			defer stream.Close()

			errorStreamReg := is.client.RegisterStream(ctx, stream)
			if errorStreamReg != nil {
				is.logger.Error("[GRPC] Error sending Ping:", zap.Error(err))
				continue
			}

			go Send(is.ctx, taskStream, stream, is.logger, receiverClosed)
			after = true
		}

		uid, err := uuid.Parse(in.Id)

		// PING type support on top of the default implementation.
		if in.Type == "PING" {
			err := stream.Send(cStructs.TaskResponse{
				Id:   uid,
				Type: "PONG",
			})
			if err != nil {
				is.logger.Error("[GRPC] Error sending Ping:", zap.Error(err))
			}
			continue
		}

		stream.Req(cStructs.TaskRequest{
			Id:      uid,
			Type:    in.Type,
			Payload: in.Payload,
		})
	}

}

// Send pairs outgoing messages, with proper stream. It shoud run in separate goroutine than TaskRPC
func Send(workerContext context.Context, taskStream indexer.IndexerService_TaskRPCServer, accessCh *cStructs.StreamAccess, logger *zap.Logger, receiverClosed chan error) {
	logger.Debug("[GRPC] Send started ")
	defer logger.Debug("[GRPC] Send finished ")
	defer logger.Sync()

	streamCtx := taskStream.Context()
ControlRPC:
	for {
		select {
		case <-streamCtx.Done():
			break ControlRPC
		case <-workerContext.Done():
			break ControlRPC
		case <-receiverClosed:
			break ControlRPC
		case resp := <-accessCh.ResponseListener:
			tr := &indexer.TaskResponse{
				Version: resp.Version,
				Id:      resp.Id.String(),
				Type:    resp.Type,
				Payload: resp.Payload,
				Order:   int64(resp.Order),
				Final:   resp.Final,
			}

			if resp.Error.Msg != "" {
				tr.Error = &indexer.TaskError{
					Msg: resp.Error.Msg,
				}
			}

			err := taskStream.Send(tr)
			if err == nil {

				continue ControlRPC
			}
			if err == io.EOF {
				logger.Debug("[GRPC] io.EOF")
				break ControlRPC
			}
			logger.Error("[GRPC] Error sending TaskResponse", zap.Error(err))
		}
	}
}

package messagerouter

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.xubinbest.com/go-game-server/internal/gateway/msgfactory"
	"github.xubinbest.com/go-game-server/internal/gateway/protoconv"
	"github.xubinbest.com/go-game-server/internal/utils"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// HandleMessage 处理HTTP消息（带熔断器支持）
func HandleMessage(r *http.Request, conn *grpc.ClientConn, cbManager interface{}) (proto.Message, error) {
	vars := mux.Vars(r)
	msgType := vars["service"] + "." + vars["path"]
	serviceName := vars["service"]
	body := getRequestBody(r)
	utils.Info("Received message", zap.String("type", msgType), zap.String("body", string(body)))

	req, err := msgfactory.GetRequestMessageStruct(msgType)
	if err != nil {
		utils.Error("GetRequestMessageStruct failed", zap.Error(err))
		return nil, err
	}

	if err := protoconv.JSONToProto(req, body); err != nil {
		return nil, fmt.Errorf("failed to convert JSON to protobuf: %v", err)
	}

	utils.Info("Request", zap.Any("request", req))

	// 如果有熔断器管理器，使用熔断器保护gRPC调用
	if cbManager != nil {
		cb := cbManager.(interface {
			Execute(ctx context.Context, key string, fn func() error) error
		})
		var resp proto.Message
		err = cb.Execute(r.Context(), serviceName, func() error {
			var dispatchErr error
			resp, dispatchErr = DispatchGRPCRequest(r.Context(), conn, req)
			return dispatchErr
		})
		if err != nil {
			utils.Error("gRPC dispatch failed (with circuit breaker)", zap.Error(err))
			return nil, err
		}
		return resp.(proto.Message), nil
	}

	// 没有熔断器时，直接调用
	resp, err := DispatchGRPCRequest(r.Context(), conn, req)
	if err != nil {
		utils.Error("gRPC dispatch failed", zap.Error(err))
		return nil, err
	}

	return resp, nil
}

func getRequestBody(r *http.Request) []byte {
	if r.Body == nil {
		return nil
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil
	}
	return body
}

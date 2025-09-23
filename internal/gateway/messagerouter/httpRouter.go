package messagerouter

import (
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

func HandleMessage(r *http.Request, conn *grpc.ClientConn) (proto.Message, error) {
	vars := mux.Vars(r)
	msgType := vars["service"] + "." + vars["path"]
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
	// Dispatch gRPC call
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

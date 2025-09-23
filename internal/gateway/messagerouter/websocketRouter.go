package messagerouter

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/gateway/msgfactory"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// HandleWSMessage 处理WebSocket消息
func HandleWSMessage(ctx context.Context, msg []byte, conn *grpc.ClientConn) ([]byte, error) {
	// 检查消息长度
	if len(msg) < 4 {
		return nil, fmt.Errorf("message too short")
	}

	// 解析消息长度
	msgLength := binary.BigEndian.Uint32(msg[:4])
	if uint32(len(msg)-4) != msgLength {
		return nil, fmt.Errorf("message length mismatch: expected %d, got %d", msgLength, len(msg)-4)
	}

	// 解析WebSocket消息
	var wsMsg pb.WSMessage
	if err := proto.Unmarshal(msg[4:], &wsMsg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal websocket message: %v", err)
	}

	// 构造消息类型
	msgType := wsMsg.Service + "." + wsMsg.Method
	utils.Info("Received message type", zap.String("type", msgType))

	// 使用msgfactory包下的GetRequestMessageStruct函数获取消息结构体
	req, err := msgfactory.GetRequestMessageStruct(msgType)
	if err != nil {
		return nil, fmt.Errorf("failed to get request message struct: %v", err)
	}

	// 将protobuf二进制数据反序列化为消息
	if err := proto.Unmarshal(wsMsg.Payload, req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf message: %v", err)
	}

	// 通过gRPC调用相应的服务
	resp, err := DispatchGRPCRequest(ctx, conn, req)
	if err != nil {
		utils.Error("gRPC dispatch failed", zap.Error(err))
		return nil, err
	}

	// 将protobuf响应序列化为二进制
	respBytes, err := proto.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protobuf response: %v", err)
	}

	// 构造响应消息
	wsResp := &pb.WSMessage{
		Service: wsMsg.Service,
		Method:  wsMsg.Method,
		Payload: respBytes,
	}

	// 序列化响应消息
	respMsgBytes, err := proto.Marshal(wsResp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response message: %v", err)
	}

	msgLen := uint32(len(respMsgBytes))
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, msgLen)

	return append(header, respMsgBytes...), nil
}

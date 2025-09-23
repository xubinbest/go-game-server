package ws

import (
	"fmt"

	"github.xubinbest.com/go-game-server/client/core"
	"github.xubinbest.com/go-game-server/internal/gateway/msgfactory"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func (c *WSClient) HandleMessage(msg *pb.WSMessage) {
	msgType := msg.Service + "." + msg.Method

	req, err := msgfactory.GetResponseMessageStruct(msgType)
	if err != nil {
		utils.Error("GetResponseMessageStruct failed", zap.Error(err))
		return
	}

	if err := proto.Unmarshal(msg.Payload, req); err != nil {
		utils.Error("Unmarshal failed", zap.Error(err))
		return
	}

	switch msgType {
	case "user.login":
		loginResp := req.(*pb.LoginResponse)
		core.SetAuth(loginResp.Token, loginResp.UserId)
		c.loginChan <- true
	default:
		fmt.Printf("\nreceived message: %+v\n", req.ProtoReflect().Interface())
	}
}

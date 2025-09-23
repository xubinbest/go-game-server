package ws

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.xubinbest.com/go-game-server/client/core"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/utils"
)

type WSClient struct {
	conn      *websocket.Conn
	closeChan chan struct{}
	msgChan   chan *pb.WSMessage
	loginChan chan bool
}

var guildPage = 1
var guildPageSize = 10

const (
	WSGatewayAddr = "ws://ws.gateway.example.com:30393/ws" // 添加 /ws 路径
)

func Start() {
	client, err := NewWSClient()
	if err != nil {
		fmt.Printf("初始化WebSocket客户端失败: %v\n", err)
		return
	}

	defer client.Close()

	runAuthMenu(client)
}

func NewWSClient() (*WSClient, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
	}

	conn, resp, err := dialer.Dial(WSGatewayAddr, nil)
	if err != nil {
		if resp != nil {
			utils.Error("WebSocket 握手失败", zap.Int("statusCode", resp.StatusCode), zap.Any("headers", resp.Header))
		}
		return nil, fmt.Errorf("连接WebSocket服务器失败: %v", err)
	}

	client := &WSClient{
		conn:      conn,
		closeChan: make(chan struct{}),
		msgChan:   make(chan *pb.WSMessage, 100),
		loginChan: make(chan bool, 1),
	}

	// 启动消息监听
	go client.readMessages()
	go client.handleMessage()

	return client, nil
}

func (c *WSClient) readMessages() {
	defer c.conn.Close()

	for {
		select {
		case <-c.closeChan:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				fmt.Printf("读取WebSocket消息失败: %v\n", err)
				return
			}
			// 检查消息长度
			if len(message) < 4 {
				utils.Warn("消息太短", zap.Int("length", len(message)))
				continue
			}

			// 解析消息长度
			msgLength := binary.BigEndian.Uint32(message[:4])
			if uint32(len(message)-4) != msgLength {
				utils.Warn("消息长度不匹配", zap.Uint32("expected", msgLength), zap.Int("actual", len(message)-4))
				continue
			}

			// 解析消息体
			var wsMsg pb.WSMessage
			if err := proto.Unmarshal(message[4:], &wsMsg); err != nil {
				utils.Error("解析消息失败", zap.Error(err))
				continue
			}

			c.msgChan <- &wsMsg
		}
	}
}

func (c *WSClient) handleMessage() {
	for {
		select {
		case <-c.closeChan:
			return
		case msg := <-c.msgChan:
			c.HandleMessage(msg)
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *WSClient) sendMessage(service, method string, data proto.Message) error {
	payload, err := proto.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %v", err)
	}

	// 构造消息
	msg := &pb.WSMessage{
		Service: service,
		Method:  method,
		Payload: payload,
	}

	// 序列化消息
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	// 添加消息长度头
	msgLen := uint32(len(msgBytes))
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, msgLen)

	if err := c.conn.WriteMessage(websocket.BinaryMessage, append(header, msgBytes...)); err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}
	return nil
}

func (c *WSClient) Register(username, password, email string) error {
	req := &pb.RegisterRequest{
		Username: username,
		Password: password,
		Email:    email,
	}

	return c.sendMessage("user", "register", req)
}

func (c *WSClient) Login(username, password string) error {
	req := &pb.LoginRequest{
		Username: username,
		Password: password,
	}

	return c.sendMessage("user", "login", req)
}

func (c *WSClient) GetInventory() error {
	req := &pb.GetInventoryRequest{
		UserId: core.GetCurrentUserID(),
	}
	return c.sendMessage("user", "getInventory", req)
}

func (c *WSClient) AddItem(tplID int64, count int32) error {
	req := &pb.AddItemRequest{
		UserId: core.GetCurrentUserID(),
		TplId:  tplID,
		Count:  count,
	}

	return c.sendMessage("user", "addItem", req)
}

func (c *WSClient) RemoveItem(itemID int64, count int32) error {
	req := &pb.RemoveItemRequest{
		UserId: core.GetCurrentUserID(),
		ItemId: itemID,
		Count:  count,
	}

	return c.sendMessage("user", "removeItem", req)
}

func (c *WSClient) UseItem(itemID int64, count int32) error {
	req := &pb.UseItemRequest{
		UserId: core.GetCurrentUserID(),
		ItemId: itemID,
		Count:  count,
	}

	return c.sendMessage("user", "useItem", req)
}

// 聊天功能
func (c *WSClient) SendChatMessage(channel int32, content string, targetID int64, extra string) error {
	req := &pb.SendChatMessageRequest{
		Channel:   channel,
		Content:   content,
		TargetId:  targetID,
		ExtraData: extra,
	}

	return c.sendMessage("social", "sendChatMessage", req)
}

func (c *WSClient) GetChatHistory(channel int32, targetID int64, page, pageSize int32) error {
	req := &pb.GetChatMessagesRequest{
		Channel:  channel,
		TargetId: targetID,
		Page:     page,
		PageSize: pageSize,
	}

	return c.sendMessage("social", "getChatMessages", req)
}

func (c *WSClient) GetFriendList() error {
	req := &pb.GetFriendListRequest{UserId: core.GetCurrentUserID()}
	return c.sendMessage("social", "getFriendList", req)
}

func (c *WSClient) SendFriendRequest(targetUserID int64) error {
	req := &pb.SendFriendRequestRequest{
		FromUserId: core.GetCurrentUserID(),
		ToUserId:   targetUserID,
	}

	return c.sendMessage("social", "sendFriendRequest", req)
}

func (c *WSClient) GetFriendRequests() error {
	req := &pb.GetFriendListRequest{UserId: core.GetCurrentUserID()}
	return c.sendMessage("social", "getFriendList", req)
}

func (c *WSClient) HandleFriendRequest(requestID int64, action int32) error {
	req := &pb.HandleFriendRequestRequest{
		RequestId: requestID,
		Action:    pb.HandleFriendRequestRequest_Action(action),
	}

	return c.sendMessage("social", "handleFriendRequest", req)
}

func (c *WSClient) BatchHandleFriendRequest(requestIDs []int64, action int32) error {
	req := &pb.BatchHandleFriendRequestRequest{
		RequestIds: requestIDs,
		Action:     pb.BatchHandleFriendRequestRequest_Action(action),
	}
	return c.sendMessage("social", "batchHandleFriendRequest", req)
}

func (c *WSClient) DeleteFriend(friendID int64) error {
	req := &pb.DeleteFriendRequest{
		UserId:   core.GetCurrentUserID(),
		FriendId: friendID,
	}

	return c.sendMessage("social", "deleteFriend", req)
}

// 帮派功能
func (c *WSClient) CreateGuild(name, description, announcement string) error {
	req := &pb.CreateGuildRequest{
		CreatorId:    core.GetCurrentUserID(),
		Name:         name,
		Description:  description,
		Announcement: announcement,
	}

	return c.sendMessage("social", "createGuild", req)
}

func (c *WSClient) GetGuildInfo(guildID int64) error {
	req := &pb.GetGuildInfoRequest{
		GuildId: guildID,
	}

	return c.sendMessage("social", "getGuildInfo", req)
}

func (c *WSClient) GetGuildMembers(guildID int64) error {
	req := &pb.GetGuildMembersRequest{
		GuildId: guildID,
	}

	return c.sendMessage("social", "getGuildMembers", req)
}

func (c *WSClient) ApplyToGuild(guildID int64) error {
	req := &pb.ApplyToGuildRequest{
		UserId:  core.GetCurrentUserID(),
		GuildId: guildID,
	}

	return c.sendMessage("social", "applyToGuild", req)
}

func (c *WSClient) InviteToGuild(targetUserID, guildID int64) error {
	req := &pb.InviteToGuildRequest{
		InviterId: core.GetCurrentUserID(),
		InviteeId: targetUserID,
		GuildId:   guildID,
	}

	return c.sendMessage("social", "inviteToGuild", req)
}

func (c *WSClient) GetGuildApplications(guildID int64) error {
	req := &pb.GetGuildApplicationsRequest{
		GuildId: guildID,
	}

	return c.sendMessage("social", "getGuildApplications", req)
}

func (c *WSClient) HandleGuildApplication(applicationID int64, action int32) error {
	req := &pb.HandleGuildApplicationRequest{
		OperatorId:    core.GetCurrentUserID(),
		ApplicationId: applicationID,
		Action:        pb.HandleGuildApplicationRequest_Action(action),
	}

	return c.sendMessage("social", "handleGuildApplication", req)
}

func (c *WSClient) KickGuildMember(memberID, guildID int64) error {
	req := &pb.KickGuildMemberRequest{
		OperatorId: core.GetCurrentUserID(),
		MemberId:   memberID,
		GuildId:    guildID,
	}

	return c.sendMessage("social", "kickGuildMember", req)
}

func (c *WSClient) ChangeMemberRole(memberID, guildID int64, role int32) error {
	req := &pb.ChangeMemberRoleRequest{
		OperatorId: core.GetCurrentUserID(),
		MemberId:   memberID,
		GuildId:    guildID,
		NewRole:    pb.GuildRole(role),
	}

	return c.sendMessage("social", "changeMemberRole", req)
}

func (c *WSClient) TransferGuildMaster(newMasterID, guildID int64) error {
	req := &pb.TransferGuildMasterRequest{
		RoleId:      core.GetCurrentUserID(),
		NewMasterId: newMasterID,
	}

	return c.sendMessage("social", "transferGuildMaster", req)
}

func (c *WSClient) DisbandGuild(guildID int64) error {
	req := &pb.DisbandGuildRequest{
		RoleId: core.GetCurrentUserID(),
	}

	return c.sendMessage("social", "disbandGuild", req)
}

func (c *WSClient) LeaveGuild(guildID int64) error {
	req := &pb.LeaveGuildRequest{
		RoleId: core.GetCurrentUserID(),
	}

	return c.sendMessage("social", "leaveGuild", req)
}

func (c *WSClient) GetGuildList(page, pageSize int32) error {
	req := &pb.GetGuildListRequest{
		Page:     page,
		PageSize: pageSize,
	}
	return c.sendMessage("social", "getGuildList", req)
}

// 游戏功能
// func (c *WSClient) JoinGame(gameID string) error {
// 	req := &pb.JoinGameRequest{
// 		PlayerId: core.GetCurrentUserID(),
// 		GameId:   gameID,
// 	}

// 	if err := c.sendMessage("user", "joinGame", req); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (c *WSClient) LeaveGame(gameID string) error {
// 	req := &pb.LeaveGameRequest{
// 		GameId: gameID,
// 	}

// 	if err := c.sendMessage("user", "leaveGame", req); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (c *WSClient) GetGameStatus(gameID string) (*core.GameStatusResult, error) {
// 	req := &pb.GetGameStatusRequest{
// 		GameId: gameID,
// 	}

// 	if err := c.sendMessage("user", "getGameStatus", req); err != nil {
// 		return nil, err
// 	}

// 	// 解析响应数据
// 	responseData, err := json.Marshal(response.Data)
// 	response, err := c.sendMessage("user", "getGameStatus", req)
// 	if err != nil {
// 		return nil, fmt.Errorf("序列化响应数据失败: %v", err)
// 	}

// 	var result core.GameStatusResult
// 	if err := json.Unmarshal(responseData, &result); err != nil {
// 		return nil, fmt.Errorf("解析响应数据失败: %v", err)
// 	}

// 	return &result, nil
// }

// func (c *WSClient) PlayerAction(gameID, action string, data []byte) error {
// 	req := &pb.PlayerActionRequest{
// 		GameId: gameID,
// 		Action: action,
// 		Data:   data,
// 	}

// 	if err := c.sendMessage("user", "playerAction", req); err != nil {
// 		return err
// 	}

// 	return nil
// }

func (c *WSClient) Close() error {
	close(c.closeChan)
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

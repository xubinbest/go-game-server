package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.xubinbest.com/go-game-server/client/core"
)

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func Start() {
	client, err := NewHTTPClient()

	if err != nil {
		fmt.Printf("初始化HTTP客户端失败: %v\n", err)
		return
	}

	defer client.Close()

	runAuthMenu(client)

}

// 具体功能实现函数
func register(client HTTPClient) {
	username := core.GetStringInput("请输入用户名: ")
	password := core.GetStringInput("请输入密码: ")
	email := core.GetStringInput("请输入邮箱: ")

	result, err := client.Register(username, password, email)
	if err != nil {
		fmt.Printf("注册失败: %v\n", err)
		return
	}

	fmt.Printf("注册成功! 用户ID: %d\n", result.UserID)
}

func login(client HTTPClient) bool {
	username := core.GetStringInput("请输入用户名: ")
	password := core.GetStringInput("请输入密码: ")

	result, err := client.Login(username, password)
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		return false
	}

	core.SetAuth(result.Token, result.UserID)
	fmt.Println("登录成功!")
	return true
}

func getInventory(client HTTPClient) {
	result, err := client.GetInventory()
	if err != nil {
		fmt.Printf("获取背包失败: %v\n", err)
		return
	}

	fmt.Println("背包内容:")
	for _, item := range result.Items {
		fmt.Printf("- 物品ID: %d, 名称: %s, 数量: %d\n", item.ItemID, item.Name, item.Count)
	}
}

func addItem(client HTTPClient) {
	itemID := core.GetIntInput("请输入物品ID: ")
	itemName := core.GetStringInput("请输入物品名称: ")
	count := core.GetIntInput("请输入数量: ")

	err := client.AddItem(int64(itemID), itemName, int32(count))
	if err != nil {
		fmt.Printf("添加物品失败: %v\n", err)
		return
	}

	fmt.Println("添加物品成功")
}

func removeItem(client HTTPClient) {
	itemID := core.GetIntInput("请输入要移除的物品ID: ")
	count := core.GetIntInput("请输入移除数量: ")

	err := client.RemoveItem(int64(itemID), int32(count))
	if err != nil {
		fmt.Printf("移除物品失败: %v\n", err)
		return
	}

	fmt.Println("移除物品成功")
}

func useItem(client HTTPClient) {
	itemID := core.GetIntInput("请输入要使用的物品ID: ")
	count := core.GetIntInput("请输入使用数量: ")

	err := client.UseItem(int64(itemID), int32(count))
	if err != nil {
		fmt.Printf("使用物品失败: %v\n", err)
		return
	}

	fmt.Println("使用物品成功")
}

func getFriendList(client HTTPClient) {
	result, err := client.GetFriendList()
	if err != nil {
		fmt.Printf("获取好友列表失败: %v\n", err)
		return
	}

	fmt.Println("好友列表:")
	for _, friend := range result.Friends {
		fmt.Printf("- 用户ID: %d\n", friend.UserID)
	}
}

func sendFriendRequest(client HTTPClient) {
	targetUserID := core.GetIntInput("请输入目标用户ID: ")

	err := client.SendFriendRequest(int64(targetUserID))
	if err != nil {
		fmt.Printf("发送好友请求失败: %v\n", err)
		return
	}

	fmt.Println("好友请求发送成功!")
}

func getFriendRequests(client HTTPClient) {
	result, err := client.GetFriendRequests()
	if err != nil {
		fmt.Printf("获取好友请求失败: %v\n", err)
		return
	}

	fmt.Println("好友请求列表:")
	for _, req := range result.Requests {
		fmt.Printf("- 请求ID: %d, 来自用户: %d\n", req.RequestID, req.FromUserID)
	}
}

func handleFriendRequest(client HTTPClient) {
	requestID := core.GetIntInput("请输入请求ID: ")
	fmt.Println("请选择操作:")
	fmt.Println("1. 同意")
	fmt.Println("2. 拒绝")
	action := core.GetIntInput("请选择(1-2): ")

	err := client.HandleFriendRequest(int64(requestID), int32(action))
	if err != nil {
		fmt.Printf("处理好友请求失败: %v\n", err)
		return
	}

	if action == 1 {
		fmt.Println("已同意好友请求")
	} else {
		fmt.Println("已拒绝好友请求")
	}
}

func deleteFriend(client HTTPClient) {
	friendID := core.GetIntInput("请输入要删除的好友ID: ")

	err := client.DeleteFriend(int64(friendID))
	if err != nil {
		fmt.Printf("删除好友失败: %v\n", err)
		return
	}

	fmt.Println("删除好友成功")
}

func joinGame(client HTTPClient) {
	gameID := core.GetStringInput("请输入游戏ID: ")

	err := client.JoinGame(gameID)
	if err != nil {
		fmt.Printf("加入游戏失败: %v\n", err)
		return
	}

	fmt.Println("成功加入游戏")
}

func leaveGame(client HTTPClient) {
	gameID := core.GetStringInput("请输入游戏ID: ")

	err := client.LeaveGame(gameID)
	if err != nil {
		fmt.Printf("离开游戏失败: %v\n", err)
		return
	}

	fmt.Println("成功离开游戏")
}

func getGameStatus(client HTTPClient) {
	gameID := core.GetStringInput("请输入游戏ID: ")

	result, err := client.GetGameStatus(gameID)
	if err != nil {
		fmt.Printf("获取游戏状态失败: %v\n", err)
		return
	}

	fmt.Printf("游戏状态: %s\n", result.State)
	fmt.Println("玩家列表:")
	for _, player := range result.Players {
		fmt.Printf("- %s\n", player)
	}
}

func playerAction(client HTTPClient) {
	gameID := core.GetStringInput("请输入游戏ID: ")
	action := core.GetStringInput("请输入动作: ")

	err := client.PlayerAction(gameID, action, []byte{})
	if err != nil {
		fmt.Printf("执行玩家动作失败: %v\n", err)
		return
	}

	fmt.Println("成功执行动作")
}

// 帮派功能
func getGuildList(client HTTPClient) {
	result, err := client.GetGuildList()
	if err != nil {
		fmt.Printf("获取帮派列表失败: %v\n", err)
		return
	}

	fmt.Println("帮派列表:")
	for _, guild := range result.Guilds {
		fmt.Printf("- 帮派ID: %d, 名称: %s, 成员: %d/%d, 帮主: %s\n",
			guild.GuildID, guild.Name, guild.MemberCount, guild.MaxMembers, guild.MasterName)
	}
}

func createGuild(client HTTPClient) {
	name := core.GetStringInput("请输入帮派名称: ")
	description := core.GetStringInput("请输入帮派描述: ")
	announcement := core.GetStringInput("请输入帮派公告: ")

	result, err := client.CreateGuild(name, description, announcement)
	if err != nil {
		fmt.Printf("创建帮派失败: %v\n", err)
		return
	}

	fmt.Printf("创建帮派成功! 帮派ID: %d, 名称: %s\n", result.GuildID, result.Name)
}

func getGuildInfo(client HTTPClient) {
	guildID := core.GetIntInput("请输入帮派ID: ")

	result, err := client.GetGuildInfo(int64(guildID))
	if err != nil {
		fmt.Printf("获取帮派信息失败: %v\n", err)
		return
	}

	guild := result.Guild
	fmt.Printf("帮派信息:\n")
	fmt.Printf("- 帮派ID: %d\n", guild.GuildID)
	fmt.Printf("- 名称: %s\n", guild.Name)
	fmt.Printf("- 描述: %s\n", guild.Description)
	fmt.Printf("- 成员数: %d/%d\n", guild.MemberCount, guild.MaxMembers)
	fmt.Printf("- 帮主: %s\n", guild.MasterName)
	fmt.Printf("- 创建时间: %s\n", guild.CreatedAt)
}

func getGuildMembers(client HTTPClient) {
	guildID := core.GetIntInput("请输入帮派ID: ")

	result, err := client.GetGuildMembers(int64(guildID))
	if err != nil {
		fmt.Printf("获取帮派成员失败: %v\n", err)
		return
	}

	fmt.Println("帮派成员:")
	for _, member := range result.Members {
		fmt.Printf("- 用户ID: %d, 用户名: %s, 职位: %s, 加入时间: %s\n",
			member.UserID, member.Username, member.Role, member.JoinTime)
	}
}

func applyToGuild(client HTTPClient) {
	guildID := core.GetIntInput("请输入帮派ID: ")

	err := client.ApplyToGuild(int64(guildID))
	if err != nil {
		fmt.Printf("申请加入帮派失败: %v\n", err)
		return
	}

	fmt.Println("申请加入帮派成功!")
}

func inviteToGuild(client HTTPClient) {
	targetUserID := core.GetIntInput("请输入目标用户ID: ")
	guildID := core.GetIntInput("请输入帮派ID: ")

	err := client.InviteToGuild(int64(targetUserID), int64(guildID))
	if err != nil {
		fmt.Printf("邀请加入帮派失败: %v\n", err)
		return
	}

	fmt.Println("邀请加入帮派成功!")
}

func getGuildApplications(client HTTPClient) {
	guildID := core.GetIntInput("请输入帮派ID: ")

	result, err := client.GetGuildApplications(int64(guildID))
	if err != nil {
		fmt.Printf("获取帮派申请列表失败: %v\n", err)
		return
	}

	fmt.Println("帮派申请列表:")
	for _, app := range result.Applications {
		fmt.Printf("- 申请ID: %d, 用户ID: %d, 用户名: %s, 申请时间: %s, 留言: %s\n",
			app.ApplicationID, app.UserID, app.Username, app.ApplyTime, app.Message)
	}
}

func handleGuildApplication(client HTTPClient) {
	applicationID := core.GetIntInput("请输入申请ID: ")
	fmt.Println("请选择操作:")
	fmt.Println("1. 同意")
	fmt.Println("2. 拒绝")
	action := core.GetIntInput("请选择(1-2): ")

	err := client.HandleGuildApplication(int64(applicationID), int32(action))
	if err != nil {
		fmt.Printf("处理帮派申请失败: %v\n", err)
		return
	}

	if action == 1 {
		fmt.Println("已同意帮派申请")
	} else {
		fmt.Println("已拒绝帮派申请")
	}
}

func kickGuildMember(client HTTPClient) {
	memberID := core.GetIntInput("请输入要踢出的成员ID: ")
	guildID := core.GetIntInput("请输入帮派ID: ")

	err := client.KickGuildMember(int64(memberID), int64(guildID))
	if err != nil {
		fmt.Printf("踢出帮派成员失败: %v\n", err)
		return
	}

	fmt.Println("踢出帮派成员成功")
}

func changeMemberRole(client HTTPClient) {
	memberID := core.GetIntInput("请输入成员ID: ")
	guildID := core.GetIntInput("请输入帮派ID: ")
	role := core.GetStringInput("请输入新职位: ")

	err := client.ChangeMemberRole(int64(memberID), int64(guildID), role)
	if err != nil {
		fmt.Printf("修改成员职位失败: %v\n", err)
		return
	}

	fmt.Println("修改成员职位成功")
}

func transferGuildMaster(client HTTPClient) {
	newMasterID := core.GetIntInput("请输入新帮主ID: ")
	guildID := core.GetIntInput("请输入帮派ID: ")

	err := client.TransferGuildMaster(int64(newMasterID), int64(guildID))
	if err != nil {
		fmt.Printf("转让帮主失败: %v\n", err)
		return
	}

	fmt.Println("转让帮主成功")
}

func disbandGuild(client HTTPClient) {
	guildID := core.GetIntInput("请输入帮派ID: ")

	err := client.DisbandGuild(int64(guildID))
	if err != nil {
		fmt.Printf("解散帮派失败: %v\n", err)
		return
	}

	fmt.Println("解散帮派成功")
}

func leaveGuild(client HTTPClient) {
	guildID := core.GetIntInput("请输入帮派ID: ")

	err := client.LeaveGuild(int64(guildID))
	if err != nil {
		fmt.Printf("离开帮派失败: %v\n", err)
		return
	}

	fmt.Println("离开帮派成功")
}

// 聊天功能
func sendWorldChat(client HTTPClient) {
	content := core.GetStringInput("请输入聊天内容: ")

	err := client.SendChatMessage(1, content, 0, "") // 世界聊天 channel=1, targetID=0
	if err != nil {
		fmt.Printf("发送世界聊天失败: %v\n", err)
		return
	}

	fmt.Println("发送世界聊天成功")
}

func sendGuildChat(client HTTPClient) {
	content := core.GetStringInput("请输入聊天内容: ")

	err := client.SendChatMessage(2, content, 0, "") // 帮派聊天 channel=2, targetID=0
	if err != nil {
		fmt.Printf("发送帮派聊天失败: %v\n", err)
		return
	}

	fmt.Println("发送帮派聊天成功")
}

func sendPrivateChat(client HTTPClient) {
	targetID := core.GetIntInput("请输入目标用户ID: ")
	content := core.GetStringInput("请输入聊天内容: ")

	err := client.SendChatMessage(3, content, int64(targetID), "") // 私聊 channel=3
	if err != nil {
		fmt.Printf("发送私聊失败: %v\n", err)
		return
	}

	fmt.Println("发送私聊成功")
}

func getChatHistory(client HTTPClient) {
	fmt.Println("请选择聊天类型:")
	fmt.Println("1. 世界聊天")
	fmt.Println("2. 帮派聊天")
	fmt.Println("3. 私聊")
	chatType := core.GetIntInput("请选择(1-3): ")

	var channel int32
	var targetID int64

	switch chatType {
	case 1:
		channel = 1
		targetID = 0
	case 2:
		channel = 2
		targetID = 0
	case 3:
		channel = 3
		targetID = int64(core.GetIntInput("请输入目标用户ID: "))
	default:
		fmt.Println("无效选择")
		return
	}

	page := core.GetIntInput("请输入页码(从1开始): ")
	pageSize := core.GetIntInput("请输入每页数量: ")

	result, err := client.GetChatHistory(channel, targetID, int32(page), int32(pageSize))
	if err != nil {
		fmt.Printf("获取聊天历史失败: %v\n", err)
		return
	}

	fmt.Printf("聊天历史 (第%d页, 共%d条):\n", result.Page, result.Total)
	for _, msg := range result.Messages {
		fmt.Printf("[%d] %s: %s\n", msg.Timestamp, msg.SenderName, msg.Content)
	}
}

func NewHTTPClient() (HTTPClient, error) {
	return HTTPClient{
		baseURL: "http://gateway.example.com:30393",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// sendPOSTRequest 发送POST请求的通用方法
func (c *HTTPClient) sendPOSTRequest(endpoint string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %v", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	token := core.GetCurrentToken()
	if token != "" {
		req.Header.Set("Authorization", token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// sendPOSTRequestWithResponse 发送POST请求并返回响应数据的通用方法
func (c *HTTPClient) sendPOSTRequestWithResponse(endpoint string, data interface{}, result interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %v", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	token := core.GetCurrentToken()
	if token != "" {
		req.Header.Set("Authorization", token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	return nil
}

func (c *HTTPClient) Register(username, password, email string) (*core.AuthResult, error) {
	data := map[string]string{
		"username": username,
		"password": password,
		"email":    email,
	}

	var result core.AuthResult
	if err := c.sendPOSTRequestWithResponse("/api/user/register", data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *HTTPClient) Login(username, password string) (*core.AuthResult, error) {
	data := map[string]string{
		"username": username,
		"password": password,
	}

	var result core.AuthResult
	if err := c.sendPOSTRequestWithResponse("/api/user/login", data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *HTTPClient) GetInventory() (*core.InventoryResult, error) {
	data := map[string]interface{}{
		"user_id": core.GetCurrentUserID(),
	}
	var result core.InventoryResult
	if err := c.sendPOSTRequestWithResponse("/api/user/getInventory", data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *HTTPClient) AddItem(itemID int64, itemName string, count int32) error {
	data := map[string]interface{}{
		"item_id":   itemID,
		"item_name": itemName,
		"count":     count,
	}

	return c.sendPOSTRequest("/api/user/addItem", data)
}

func (c *HTTPClient) RemoveItem(itemID int64, count int32) error {
	data := map[string]interface{}{
		"item_id": itemID,
		"count":   count,
	}

	return c.sendPOSTRequest("/api/user/removeItem", data)
}

func (c *HTTPClient) UseItem(itemID int64, count int32) error {
	data := map[string]interface{}{
		"item_id": itemID,
		"count":   count,
	}

	return c.sendPOSTRequest("/api/user/useItem", data)
}

func (c *HTTPClient) GetFriendList() (*core.FriendListResult, error) {
	data := map[string]interface{}{
		"user_id": core.GetCurrentUserID(),
	}
	var result core.FriendListResult
	if err := c.sendPOSTRequestWithResponse("/api/social/getFriendList", data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *HTTPClient) SendFriendRequest(targetUserID int64) error {
	data := map[string]interface{}{
		"target_user_id": targetUserID,
	}

	return c.sendPOSTRequest("/api/social/sendFriendRequest", data)
}

func (c *HTTPClient) GetFriendRequests() (*core.FriendRequestsResult, error) {
	data := map[string]interface{}{
		"user_id": core.GetCurrentUserID(),
	}
	var result core.FriendRequestsResult
	if err := c.sendPOSTRequestWithResponse("/api/social/getFriendRequestList", data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *HTTPClient) HandleFriendRequest(requestID int64, action int32) error {
	data := map[string]interface{}{
		"request_id": requestID,
		"action":     action,
	}

	return c.sendPOSTRequest("/api/social/handleFriendRequest", data)
}

func (c *HTTPClient) DeleteFriend(friendID int64) error {
	data := map[string]interface{}{
		"user_id":   core.GetCurrentUserID(),
		"friend_id": friendID,
	}

	return c.sendPOSTRequest("/api/social/deleteFriend", data)
}

// 帮派功能
func (c *HTTPClient) CreateGuild(name, description, announcement string) (*core.GuildResult, error) {
	data := map[string]interface{}{
		"creator_id":   core.GetCurrentUserID(),
		"name":         name,
		"description":  description,
		"announcement": announcement,
	}

	var result core.GuildResult
	if err := c.sendPOSTRequestWithResponse("/api/social/createGuild", data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *HTTPClient) GetGuildList() (*core.GuildListResult, error) {
	data := map[string]interface{}{
		"page":      1,
		"page_size": 10,
	}

	var result core.GuildListResult
	if err := c.sendPOSTRequestWithResponse("/api/social/getGuildList", data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *HTTPClient) GetGuildInfo(guildID int64) (*core.GuildInfoResult, error) {
	data := map[string]interface{}{
		"guild_id": guildID,
	}

	var result core.GuildInfoResult
	if err := c.sendPOSTRequestWithResponse("/api/social/getGuildInfo", data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *HTTPClient) GetGuildMembers(guildID int64) (*core.GuildMembersResult, error) {
	data := map[string]interface{}{
		"guild_id": guildID,
	}

	var result core.GuildMembersResult
	if err := c.sendPOSTRequestWithResponse("/api/social/getGuildMembers", data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *HTTPClient) ApplyToGuild(guildID int64) error {
	data := map[string]interface{}{
		"user_id":  core.GetCurrentUserID(),
		"guild_id": guildID,
	}

	return c.sendPOSTRequest("/api/social/applyToGuild", data)
}

func (c *HTTPClient) InviteToGuild(targetUserID, guildID int64) error {
	data := map[string]interface{}{
		"user_id":        core.GetCurrentUserID(),
		"target_user_id": targetUserID,
		"guild_id":       guildID,
	}

	return c.sendPOSTRequest("/api/social/inviteToGuild", data)
}

func (c *HTTPClient) GetGuildApplications(guildID int64) (*core.GuildApplicationsResult, error) {
	data := map[string]interface{}{
		"guild_id": guildID,
	}

	var result core.GuildApplicationsResult
	if err := c.sendPOSTRequestWithResponse("/api/social/getGuildApplications", data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *HTTPClient) HandleGuildApplication(applicationID int64, action int32) error {
	data := map[string]interface{}{
		"operator_id":    core.GetCurrentUserID(),
		"application_id": applicationID,
		"action":         action,
	}

	return c.sendPOSTRequest("/api/social/handleGuildApplication", data)
}

func (c *HTTPClient) KickGuildMember(memberID, guildID int64) error {
	data := map[string]interface{}{
		"operator_id": core.GetCurrentUserID(),
		"member_id":   memberID,
		"guild_id":    guildID,
	}

	return c.sendPOSTRequest("/api/social/kickGuildMember", data)
}

func (c *HTTPClient) ChangeMemberRole(memberID, guildID int64, role string) error {
	data := map[string]interface{}{
		"operator_id": core.GetCurrentUserID(),
		"member_id":   memberID,
		"guild_id":    guildID,
		"new_role":    role,
	}

	return c.sendPOSTRequest("/api/social/changeMemberRole", data)
}

func (c *HTTPClient) TransferGuildMaster(newMasterID, guildID int64) error {
	data := map[string]interface{}{
		"current_master_id": core.GetCurrentUserID(),
		"new_master_id":     newMasterID,
		"guild_id":          guildID,
	}

	return c.sendPOSTRequest("/api/social/transferGuildMaster", data)
}

func (c *HTTPClient) DisbandGuild(guildID int64) error {
	data := map[string]interface{}{
		"master_id": core.GetCurrentUserID(),
		"guild_id":  guildID,
	}

	return c.sendPOSTRequest("/api/social/disbandGuild", data)
}

func (c *HTTPClient) LeaveGuild(guildID int64) error {
	data := map[string]interface{}{
		"member_id": core.GetCurrentUserID(),
		"guild_id":  guildID,
	}

	return c.sendPOSTRequest("/api/social/leaveGuild", data)
}

// 聊天功能
func (c *HTTPClient) SendChatMessage(channel int32, content string, targetID int64, extra string) error {
	data := map[string]interface{}{
		"sender_id": core.GetCurrentUserID(),
		"channel":   channel,
		"content":   content,
		"target_id": targetID,
		"extra":     extra,
	}

	return c.sendPOSTRequest("/api/social/sendChatMessage", data)
}

func (c *HTTPClient) GetChatHistory(channel int32, targetID int64, page, pageSize int32) (*core.ChatHistoryResult, error) {
	data := map[string]interface{}{
		"channel":   channel,
		"target_id": targetID,
		"page":      page,
		"page_size": pageSize,
	}

	var result core.ChatHistoryResult
	if err := c.sendPOSTRequestWithResponse("/api/social/getChatMessages", data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *HTTPClient) JoinGame(gameID string) error {
	data := map[string]interface{}{
		"game_id": gameID,
	}

	return c.sendPOSTRequest("/api/game/joinGame", data)
}

func (c *HTTPClient) LeaveGame(gameID string) error {
	data := map[string]interface{}{
		"game_id": gameID,
	}

	return c.sendPOSTRequest("/api/game/leaveGame", data)
}

func (c *HTTPClient) GetGameStatus(gameID string) (*core.GameStatusResult, error) {
	data := map[string]interface{}{
		"game_id": gameID,
	}

	var result core.GameStatusResult
	if err := c.sendPOSTRequestWithResponse("/api/game/getGameStatus", data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *HTTPClient) PlayerAction(gameID, action string, data []byte) error {
	requestData := map[string]interface{}{
		"game_id": gameID,
		"action":  action,
		"data":    data,
	}

	return c.sendPOSTRequest("/api/game/playerAction", requestData)
}

func (c *HTTPClient) Close() error {
	// HTTP客户端不需要特殊关闭
	return nil
}

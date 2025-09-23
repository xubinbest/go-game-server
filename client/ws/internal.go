package ws

import (
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/client/core"
	"github.xubinbest.com/go-game-server/client/menu"
)

func runAuthMenu(client *WSClient) {
	for {
		menu.ShowAuthMenu()
		choice := core.GetIntInput("请选择(0-2): ")
		fmt.Println("choice: ", choice)
		switch choice {
		case 1:
			register(client)
		case 2:
			login(client)
		case 0:
			return
		default:
			fmt.Println("无效选择")
		}
	}
}

func runGameMainMenu(client *WSClient) {
	for {
		menu.ShowGameMainMenu()
		choice := core.GetIntInput("请选择(0-3): ")

		switch choice {
		case 0:
			fmt.Println("退出程序")
			return
		case 1:
			handlePlayerMenu(client)
		case 2:
			handleSocialMenu(client)
		case 3:
			handleGameMenu(client)
		default:
			fmt.Println("无效选择")
		}
	}
}

// 玩家菜单
func handlePlayerMenu(client *WSClient) {
	for {
		menu.ShowPlayerMenu()
		choice := core.GetIntInput("请选择(0-4): ")

		switch choice {
		case 0:
			return
		case 1:
			getInventory(client)
		case 2:
			addItem(client)
		case 3:
			removeItem(client)
		case 4:
			useItem(client)
		default:
			fmt.Println("无效选择")
		}
	}
}

// 社交菜单
func handleSocialMenu(client *WSClient) {
	for {
		menu.ShowSocialMainMenu()
		choice := core.GetIntInput("请选择(0-3): ")

		switch choice {
		case 0:
			return
		case 1:
			handleFriendMenu(client)
		case 2:
			handleGuildMenu(client)
		case 3:
			handleChatMenu(client)
		default:
			fmt.Println("无效选择")
		}
	}
}

// 好友菜单
func handleFriendMenu(client *WSClient) {
	for {
		menu.ShowFriendMenu()
		choice := core.GetIntInput("请选择(0-5): ")

		switch choice {
		case 0:
			return
		case 1:
			getFriendList(client)
		case 2:
			sendFriendRequest(client)
		case 3:
			getFriendRequests(client)
		case 4:
			handleFriendRequest(client)
		case 5:
			deleteFriend(client)
		default:
			fmt.Println("无效选择")
		}
	}
}

// 帮派菜单
func handleGuildMenu(client *WSClient) {
	for {
		menu.ShowGuildMenu()
		choice := core.GetIntInput("请选择(0-13): ")

		switch choice {
		case 0:
			return
		case 1:
			getGuildList(client)
		case 2:
			createGuild(client)
		case 3:
			getGuildInfo(client)
		case 4:
			getGuildMembers(client)
		case 5:
			applyToGuild(client)
		case 6:
			inviteToGuild(client)
		case 7:
			getGuildApplications(client)
		case 8:
			handleGuildApplication(client)
		case 9:
			kickGuildMember(client)
		case 10:
			changeMemberRole(client)
		case 11:
			transferGuildMaster(client)
		case 12:
			disbandGuild(client)
		case 13:
			leaveGuild(client)
		default:
			fmt.Println("无效选择")
		}
	}
}

// 聊天菜单
func handleChatMenu(client *WSClient) {
	for {
		menu.ShowChatMenu()
		choice := core.GetIntInput("请选择(0-4): ")

		switch choice {
		case 0:
			return
		case 1:
			sendWorldChat(client)
		case 2:
			sendGuildChat(client)
		case 3:
			sendPrivateChat(client)
		case 4:
			getChatHistory(client)
		default:
			fmt.Println("无效选择")
		}
	}
}

// 游戏菜单
func handleGameMenu(client *WSClient) {
	fmt.Println("游戏菜单开发中...")
	// for {
	// 	menu.ShowGameMenu()
	// 	choice := core.GetIntInput("请选择(0-4): ")

	// 	switch choice {
	// 	case 0:
	// 		return
	// 	case 1:
	// 		joinGame(client)
	// 	case 2:
	// 		leaveGame(client)
	// 	case 3:
	// 		getGameStatus(client)
	// 	case 4:
	// 		playerAction(client)
	// 	default:
	// 		fmt.Println("无效选择")
	// 	}
	// }
}

func register(client *WSClient) {
	username := core.GetStringInput("请输入用户名: ")
	password := core.GetStringInput("请输入密码: ")
	email := core.GetStringInput("请输入邮箱: ")

	err := client.Register(username, password, email)
	if err != nil {
		fmt.Printf("注册失败: %v\n", err)
	}
}

func login(client *WSClient) {
	username := core.GetStringInput("请输入用户名: ")
	password := core.GetStringInput("请输入密码: ")
	err := client.Login(username, password)
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
	}

	// 增加超时机制
	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()

	select {
	case <-client.closeChan:
		return
	case <-timeout.C:
		fmt.Println("登录超时")
		return
	case ok := <-client.loginChan:
		if ok {
			runGameMainMenu(client)
		} else {
			fmt.Println("登录失败")
		}
	}
}

func getInventory(client *WSClient) {
	err := client.GetInventory()
	if err != nil {
		fmt.Printf("获取背包失败: %v\n", err)
	}
}

func addItem(client *WSClient) {
	tplID := core.GetIntInput("请输入物品ID: ")
	count := core.GetIntInput("请输入数量: ")
	err := client.AddItem(int64(tplID), int32(count))
	if err != nil {
		fmt.Printf("添加物品失败: %v\n", err)
	}
}

func removeItem(client *WSClient) {
	itemID := core.GetIntInput("请输入物品ID: ")
	count := core.GetIntInput("请输入数量: ")
	err := client.RemoveItem(int64(itemID), int32(count))
	if err != nil {
		fmt.Printf("移除物品失败: %v\n", err)
	}
}

func useItem(client *WSClient) {
	itemID := core.GetIntInput("请输入物品ID: ")
	count := core.GetIntInput("请输入数量: ")
	err := client.UseItem(int64(itemID), int32(count))
	if err != nil {
		fmt.Printf("使用物品失败: %v\n", err)
	}
}

func getFriendList(client *WSClient) {
	err := client.GetFriendList()
	if err != nil {
		fmt.Printf("获取好友列表失败: %v\n", err)
	}
}

func sendFriendRequest(client *WSClient) {
	targetUserID := core.GetIntInput("请输入目标用户ID: ")
	err := client.SendFriendRequest(int64(targetUserID))
	if err != nil {
		fmt.Printf("发送好友请求失败: %v\n", err)
	}
}

func getFriendRequests(client *WSClient) {
	err := client.GetFriendRequests()
	if err != nil {
		fmt.Printf("获取好友请求失败: %v\n", err)
	}
}

func handleFriendRequest(client *WSClient) {
	requestID := core.GetIntInput("请输入请求ID: ")
	action := core.GetIntInput("请选择(1-2): ")
	err := client.HandleFriendRequest(int64(requestID), int32(action))
	if err != nil {
		fmt.Printf("处理好友请求失败: %v\n", err)
	}
}

func deleteFriend(client *WSClient) {
	friendID := core.GetIntInput("请输入好友ID: ")
	err := client.DeleteFriend(int64(friendID))
	if err != nil {
		fmt.Printf("删除好友失败: %v\n", err)
	}
}

func getGuildList(client *WSClient) {
	err := client.GetGuildList(int32(guildPage), int32(guildPageSize))
	if err != nil {
		fmt.Printf("获取帮派列表失败: %v\n", err)
	}
}

func createGuild(client *WSClient) {
	name := core.GetStringInput("请输入帮派名称: ")
	description := core.GetStringInput("请输入帮派描述: ")
	announcement := core.GetStringInput("请输入帮派公告: ")
	err := client.CreateGuild(name, description, announcement)
	if err != nil {
		fmt.Printf("创建帮派失败: %v\n", err)
	}
}

func getGuildInfo(client *WSClient) {
	guildID := core.GetIntInput("请输入帮派ID: ")
	err := client.GetGuildInfo(int64(guildID))
	if err != nil {
		fmt.Printf("获取帮派信息失败: %v\n", err)
	}
}

func getGuildMembers(client *WSClient) {
	guildID := core.GetIntInput("请输入帮派ID: ")
	err := client.GetGuildMembers(int64(guildID))
	if err != nil {
		fmt.Printf("获取帮派成员失败: %v\n", err)
	}
}

func applyToGuild(client *WSClient) {
	guildID := core.GetIntInput("请输入帮派ID: ")
	err := client.ApplyToGuild(int64(guildID))
	if err != nil {
		fmt.Printf("申请加入帮派失败: %v\n", err)
	}
}

func inviteToGuild(client *WSClient) {
	targetUserID := core.GetIntInput("请输入目标用户ID: ")
	guildID := core.GetIntInput("请输入帮派ID: ")
	err := client.InviteToGuild(int64(targetUserID), int64(guildID))
	if err != nil {
		fmt.Printf("邀请加入帮派失败: %v\n", err)
	}
}

func getGuildApplications(client *WSClient) {
	guildID := core.GetIntInput("请输入帮派ID: ")
	err := client.GetGuildApplications(int64(guildID))
	if err != nil {
		fmt.Printf("获取帮派申请失败: %v\n", err)
	}
}

func handleGuildApplication(client *WSClient) {
	applicationID := core.GetIntInput("请输入申请ID: ")
	action := core.GetIntInput("请选择(1-2): ")
	err := client.HandleGuildApplication(int64(applicationID), int32(action))
	if err != nil {
		fmt.Printf("处理帮派申请失败: %v\n", err)
	}
}

func kickGuildMember(client *WSClient) {
	memberID := core.GetIntInput("请输入成员ID: ")
	guildID := core.GetIntInput("请输入帮派ID: ")
	err := client.KickGuildMember(int64(memberID), int64(guildID))
	if err != nil {
		fmt.Printf("踢出帮派成员失败: %v\n", err)
	}
}

func changeMemberRole(client *WSClient) {
	memberID := core.GetIntInput("请输入成员ID: ")
	guildID := core.GetIntInput("请输入帮派ID: ")
	role := core.GetIntInput("请输入新职位: ")
	err := client.ChangeMemberRole(int64(memberID), int64(guildID), int32(role))
	if err != nil {
		fmt.Printf("修改成员职位失败: %v\n", err)
	}
}

func transferGuildMaster(client *WSClient) {
	newMasterID := core.GetIntInput("请输入新帮主ID: ")
	guildID := core.GetIntInput("请输入帮派ID: ")
	err := client.TransferGuildMaster(int64(newMasterID), int64(guildID))
	if err != nil {
		fmt.Printf("转让帮主失败: %v\n", err)
	}
}

func disbandGuild(client *WSClient) {
	guildID := core.GetIntInput("请输入帮派ID: ")
	err := client.DisbandGuild(int64(guildID))
	if err != nil {
		fmt.Printf("解散帮派失败: %v\n", err)
	}
}

func leaveGuild(client *WSClient) {
	guildID := core.GetIntInput("请输入帮派ID: ")
	err := client.LeaveGuild(int64(guildID))
	if err != nil {
		fmt.Printf("离开帮派失败: %v\n", err)
	}
}

func sendWorldChat(client *WSClient) {
	content := core.GetStringInput("请输入聊天内容: ")
	err := client.SendChatMessage(1, content, 0, "")
	if err != nil {
		fmt.Printf("发送世界聊天失败: %v\n", err)
	}
}

func sendGuildChat(client *WSClient) {
	content := core.GetStringInput("请输入聊天内容: ")
	err := client.SendChatMessage(2, content, 0, "")
	if err != nil {
		fmt.Printf("发送帮派聊天失败: %v\n", err)
	}
}

func sendPrivateChat(client *WSClient) {
	content := core.GetStringInput("请输入聊天内容: ")
	targetUserID := core.GetIntInput("请输入目标用户ID: ")
	err := client.SendChatMessage(3, content, int64(targetUserID), "")
	if err != nil {
		fmt.Printf("发送私聊失败: %v\n", err)
	}
}

func getChatHistory(client *WSClient) {
	channel := core.GetIntInput("请输入频道(1-3): ")
	targetID := core.GetIntInput("请输入目标ID: ")
	page := core.GetIntInput("请输入页码: ")
	pageSize := core.GetIntInput("请输入每页数量: ")
	err := client.GetChatHistory(int32(channel), int64(targetID), int32(page), int32(pageSize))
	if err != nil {
		fmt.Printf("获取聊天历史失败: %v\n", err)
	}
}

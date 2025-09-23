package http

import (
	"fmt"

	"github.xubinbest.com/go-game-server/client/core"
	"github.xubinbest.com/go-game-server/client/menu"
)

// 认证菜单
func runAuthMenu(client HTTPClient) {
	for {
		menu.ShowAuthMenu()
		choice := core.GetIntInput("请选择(0-2): ")

		switch choice {
		case 0:
			fmt.Println("退出程序")
			return
		case 1:
			register(client)
		case 2:
			if login(client) {
				// 登录成功后进入游戏主界面
				runGameMainMenu(client)
			}
		default:
			fmt.Println("无效选择")
		}
	}
}

// 游戏主界面
func runGameMainMenu(client HTTPClient) {
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
func handlePlayerMenu(client HTTPClient) {
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
func handleSocialMenu(client HTTPClient) {
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
func handleFriendMenu(client HTTPClient) {
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
func handleGuildMenu(client HTTPClient) {
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
func handleChatMenu(client HTTPClient) {
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
func handleGameMenu(client HTTPClient) {
	for {
		menu.ShowGameMenu()
		choice := core.GetIntInput("请选择(0-4): ")

		switch choice {
		case 0:
			return
		case 1:
			joinGame(client)
		case 2:
			leaveGame(client)
		case 3:
			getGameStatus(client)
		case 4:
			playerAction(client)
		default:
			fmt.Println("无效选择")
		}
	}
}

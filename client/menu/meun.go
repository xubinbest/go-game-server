package menu

import "fmt"

// 菜单显示函数
func ShowAuthMenu() {
	fmt.Println("\n=== 认证菜单 ===")
	fmt.Println("1. 注册")
	fmt.Println("2. 登录")
	fmt.Println("0. 退出")
}

func ShowGameMainMenu() {
	fmt.Println("\n=== 游戏主界面 ===")
	fmt.Println("1. 玩家")
	fmt.Println("2. 社交")
	fmt.Println("3. 游戏")
	fmt.Println("0. 退出")
}

func ShowPlayerMenu() {
	fmt.Println("\n=== 玩家功能 ===")
	fmt.Println("1. 获取背包")
	fmt.Println("2. 添加物品")
	fmt.Println("3. 移除物品")
	fmt.Println("4. 使用物品")
	fmt.Println("0. 返回")
}

func ShowSocialMainMenu() {
	fmt.Println("\n=== 社交功能 ===")
	fmt.Println("1. 好友")
	fmt.Println("2. 帮派")
	fmt.Println("3. 聊天")
	fmt.Println("0. 返回")
}

func ShowFriendMenu() {
	fmt.Println("\n=== 好友功能 ===")
	fmt.Println("1. 获取好友列表")
	fmt.Println("2. 发送好友请求")
	fmt.Println("3. 获取好友请求")
	fmt.Println("4. 处理好友请求")
	fmt.Println("5. 删除好友")
	fmt.Println("0. 返回")
}

func ShowGuildMenu() {
	fmt.Println("\n=== 帮派功能 ===")
	fmt.Println("1. 帮派列表")
	fmt.Println("2. 创建帮派")
	fmt.Println("3. 获取帮派信息")
	fmt.Println("4. 获取帮派成员")
	fmt.Println("5. 申请加入帮派")
	fmt.Println("6. 邀请加入帮派")
	fmt.Println("7. 获取帮派申请列表")
	fmt.Println("8. 处理帮派申请")
	fmt.Println("9. 踢出帮派成员")
	fmt.Println("10. 修改成员职位")
	fmt.Println("11. 转让帮主")
	fmt.Println("12. 解散帮派")
	fmt.Println("13. 离开帮派")
	fmt.Println("0. 返回")
}

func ShowChatMenu() {
	fmt.Println("\n=== 聊天功能 ===")
	fmt.Println("1. 世界聊天")
	fmt.Println("2. 帮派聊天")
	fmt.Println("3. 私聊")
	fmt.Println("4. 获取聊天历史")
	fmt.Println("0. 返回")
}

func ShowGameMenu() {
	fmt.Println("\n=== 游戏功能 ===")
	fmt.Println("1. 加入游戏")
	fmt.Println("2. 离开游戏")
	fmt.Println("3. 获取游戏状态")
	fmt.Println("4. 执行玩家动作")
	fmt.Println("0. 返回")
}

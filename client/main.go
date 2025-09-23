package main

import (
	"fmt"

	"github.xubinbest.com/go-game-server/client/http"
	"github.xubinbest.com/go-game-server/client/ws"
)

func main() {
	fmt.Println("=== 游戏服务器客户端 ===")
	fmt.Println("请选择连接方式:")
	fmt.Println("1. HTTP API")
	fmt.Println("2. WebSocket")
	fmt.Println("0. 退出")

	var choice int
	fmt.Print("请选择(0-2): ")
	fmt.Scanln(&choice)

	switch choice {
	case 0:
		fmt.Println("退出程序")
		return
	case 1:
		fmt.Println("使用HTTP API模式")
		http.Start()
	case 2:
		fmt.Println("使用WebSocket模式")
		ws.Start()
	default:
		fmt.Println("无效选择")
		return
	}
}

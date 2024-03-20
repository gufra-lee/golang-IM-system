package main

func main() {
	// 初始化Server结构体
	server := NewServer("127.0.0.1", 8888)
	// 函数入口
	server.Start()
}

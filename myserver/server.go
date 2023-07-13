package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	IP   string
	Port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 创建一个Server的接口（相当于类的）
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听message广播消息channel的goroutine，一旦有消息就发送给全部在线的User
func (this *Server) ListenMessage() {
	for {
		msg := <- this.Message

		// 将msg全部发给全部的在线user
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {  // 返回key-value，value是*user
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
}

// 当前应用的业务函数
func (this *Server) Handler(conn net.Conn) {

	//fmt.Println("连接成功")

	// 创建用户对象user
	user := NewUser(conn, this)

	// 用户上线
	user.Online()

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接受客户端发来的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)  // 这里接User类中的write()方法
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 提取用户的消息（去掉/n）
			msg := string(buf[:n-1])

			// 用户根据msg进行消息处理
			user.DoMessage(msg)

			// 用户的任意消息，代表用户当前是活跃的
			isLive <- true
		}
	}()

	// 当前handler的阻塞
	for {
		select {  // select会阻塞等待任意case发声
			case <- isLive:  // (写在最上面)
				//当前用户是活跃的，重置定时器（不做任何操作，只为了 更新下面的定时器）
			case <- time.After(time.Second * 300):  // 定时器10s,本质是个channel
				// 已经超时
				// 将当前的User强制的关闭
				user.SendMsg("因超时，你被踢了")

				// 销毁资源
				close(user.C)
				conn.Close()

				// 推出当前handle
				return
		}
	}
}

// 启动服务器的方法
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.IP, this.Port))
	if err != nil {
		fmt.Println("net.Listen err", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// 启动监听message服务器
	go this.ListenMessage()

	for {
		// accept
		conn, err := listener.Accept() // 阻塞监听
		if err != nil {
			fmt.Println("listener.accept err", err)
			continue
		}

		// do handler
		go this.Handler(conn) // 创建一个携程去处理业务
	}
}

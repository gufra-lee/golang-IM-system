package main

import (
	"net"
	"strings"
)

// 创建用户结构体
type User struct {
	Name   string
	Addr   string
	C      chan string // 每个user都包含一个携程
	conn   net.Conn    // 每个用户维护一个连接池
	server *Server     // 指定用户是属于哪个server的
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String() // 返回的是远程连接的地址

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server, // 当前user是属于哪个server的
	}

	// 只要当前User结构体里的消息队列C有消息，就用网络conn发出去
	go user.ListenMessage()

	return user
}

// 用户上线的业务
// go语言中没有this关键字，这里是定义了this
func (this *User) Online() {
	// 用户上线，将用户加入到Service.OnlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播用户上线消息(发到service自己的消息队列message中)
	this.server.BroadCast(this, "已上线")
}

// 用户下线的业务
func (this *User) Offline() {
	// 用户下线，将用户加入到onlineMap中
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播用户下线消息
	this.server.BroadCast(this, "已下线")
}

// 将当前user对应的客户端发消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			OnlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			this.SendMsg(OnlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式 "rename|张三" 更改用户姓名
		newName := strings.Split(msg, "|")[1]

		// 判断name是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("当前用户名已被使用\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("用户名已更新：" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式（to|用户名|白菜）
		// 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if (remoteName) == "" {
			this.SendMsg("消息格式不正确，请使用 \"to|张三|你好啊\"格式。\n")
			return
		}

		// 检查用户名对方的对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("该用户名不存在\n")
			return
		}

		//获取消息内容，通过对象的user发送消息
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("无消息记录，请重发\n")
			return
		}
		remoteUser.SendMsg(this.Name + "对您说" + content)
	} else {
		this.server.BroadCast(this, msg)
	}
}

// 监听当前User channel的方法，方法一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}

package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	IP               string
	Port             int
	OnlineMap        map[string]*User
	BoardcastMessage chan string
	mapLock          sync.RWMutex
}

//结构体初始化函数
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:               ip,
		Port:             port,
		OnlineMap:        make(map[string]*User),
		BoardcastMessage: make(chan string),
	}
	return server
}

//地址和端口拼接到一起
func (server *Server) toString() string {
	return fmt.Sprintf("%s:%d", server.IP, server.Port)
}

//处理单个业务
func (server *Server) handler(connection net.Conn) {
	fmt.Println("连接建立成功！")
	fmt.Println("当前连接客户端的地址为:", connection.RemoteAddr().String())
	user := NewUser(connection, server)
	user.Online()
	isLive := make(chan bool)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := connection.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("套接字读取错误:", err)
				return
			}
			//去除换行符
			msg := string(buf)
			msg = strings.TrimSpace(msg)
			fmt.Println(msg)
			user.DoMessage(msg)
			isLive <- true
		}
	}()
	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 99):
			user.connection.Write([]byte("长时间未活跃，已自动下线。\n"))
			close(user.Channel)
			delete(user.server.OnlineMap, user.Name)
			connection.Close()
			return
		}
	}
}
func (server *Server) Boardcast(user *User, msg string) {
	sendMSG := "[" + user.Address + "]" + user.Name + "说：" + msg
	server.BoardcastMessage <- sendMSG
}

//监听message广播消息channel的goroutine，一旦有消息就发送给全部的在线用户
func (server *Server) ListenMessage() {
	for {
		msg := <-server.BoardcastMessage
		server.mapLock.Lock()
		for _, client := range server.OnlineMap {
			client.Channel <- msg
		}
		server.mapLock.Unlock()
	}
}

//服务器运行
func (server *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", server.toString())
	if err != nil {
		fmt.Println("net.Listen函数调用出错:", err)
		return
	}
	defer listener.Close()
	//启动消息监听协程
	go server.ListenMessage()
	for {
		//socket accept
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("接受套接字时发生错误:", err)
			continue
		}
		//handler
		go server.handler(connection)
	}
}

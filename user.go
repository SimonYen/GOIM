package main

import (
	"net"
	"strings"
)

type User struct {
	Name       string
	Address    string
	Channel    chan string
	connection net.Conn
	server     *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	user := &User{
		Name:       conn.RemoteAddr().String(),
		Address:    conn.RemoteAddr().String(),
		Channel:    make(chan string),
		connection: conn,
		server:     server,
	}
	go user.ListenMessage()
	return user
}

//监听管道，一有回复马上发送
func (user *User) ListenMessage() {
	for {
		msg := <-user.Channel
		user.connection.Write([]byte(msg + "\n"))
	}
}

func (user *User) Online() {
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()
	user.server.Boardcast(user, "已上线")
}
func (user *User) Offline() {
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()
	user.server.Boardcast(user, "已下线")
}

func (user *User) DoMessage(msg string) {
	if msg == "#$%!WHO" {
		user.server.mapLock.Lock()
		for _, user := range user.server.OnlineMap {
			msg := "[" + user.Address + "]:" + user.Name + " 在线。"
			user.connection.Write([]byte(msg))
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "#$%!RN|" {
		newName := strings.Split(msg, "|")[1]
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.connection.Write([]byte("当前用户名已被占用！"))
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()
			user.Name = newName
			user.connection.Write([]byte("已更新新的用户名：" + user.Name + "\n"))
		}
	} else {
		user.server.Boardcast(user, msg)
	}
}

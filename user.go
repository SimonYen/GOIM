package main

import (
	"fmt"
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
	user.server.Boardcast(user, "已上线\n")
}
func (user *User) Offline() {
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()
	user.server.Boardcast(user, "已下线\n")
}

func (user *User) DoMessage(msg string) {
	if strings.HasPrefix(msg, "!") {
		userNames := "在线的用户有：\n"
		user.server.mapLock.Lock()
		for k := range user.server.OnlineMap {
			userNames = userNames + k + "\t"
		}
		user.server.mapLock.Unlock()
		user.Channel <- userNames
	} else if strings.HasPrefix(msg, "@") {
		newName := strings.Split(msg, "|")[1]
		user.server.mapLock.Lock()
		_, ok := user.server.OnlineMap[newName]
		user.server.mapLock.Unlock()
		if ok {
			m := newName + "已存在！"
			user.Channel <- m
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			newName = strings.Trim(newName, "0")
			fmt.Println(len([]rune(newName)))
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()
			user.Name = newName
			fmt.Println(len([]rune(user.Name)))
			m := "已成功更新用户名，当前你的用户名为：" + user.Name
			user.Channel <- m
		}
	} else if strings.HasPrefix(msg, "#") {
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			user.Channel <- "用户名不能为空！"
			return
		}
		user.server.mapLock.Lock()
		remoteUser, ok := user.server.OnlineMap[remoteName]
		user.server.mapLock.Unlock()
		if !ok {
			fmt.Println(len([]rune(remoteName)))
			user.server.mapLock.Lock()
			for k := range user.server.OnlineMap {
				fmt.Println(k)
				fmt.Println(len([]rune(k)))
			}
			user.server.mapLock.Unlock()
			m := remoteName + "不存在！"
			user.Channel <- m
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.Channel <- "消息不能为空！"
			return
		}
		m := user.Name + "对你说： " + content
		remoteUser.Channel <- m
	} else {
		user.server.Boardcast(user, msg)
	}
}

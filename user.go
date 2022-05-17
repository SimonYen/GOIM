package main

import (
	"net"
	"strconv"
	"strings"
)

type User struct {
	ID         int
	Name       string
	Address    string
	Channel    chan string
	connection net.Conn
	server     *Server
}

func NewUser(conn net.Conn, server *Server, ID int) *User {
	user := &User{
		Name:       conn.RemoteAddr().String(),
		Address:    conn.RemoteAddr().String(),
		Channel:    make(chan string),
		connection: conn,
		server:     server,
		ID:         ID,
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
	user.server.OnlineMap[user.ID] = user
	user.server.mapLock.Unlock()
	user.server.Boardcast(user, "已上线\n")
}
func (user *User) Offline() {
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.ID)
	user.server.mapLock.Unlock()
	user.server.Boardcast(user, "已下线\n")
}

func (user *User) DoMessage(msg string) {
	if strings.HasPrefix(msg, "!") {
		userNames := "在线的用户有：\n用户ID\t用户名\tIP地址\n"
		user.server.mapLock.Lock()
		for ID, u := range user.server.OnlineMap {
			userNames += strconv.Itoa(ID) + "\t" + u.Name + "\t" + u.Address + "\n"
		}
		user.server.mapLock.Unlock()
		user.Channel <- userNames
	} else if strings.HasPrefix(msg, "@") {
		newName := strings.Split(msg, "|")[1]
		user.Name = newName
		user.Channel <- "已更新用户名，现在你的用户名为：" + user.Name
	} else if strings.HasPrefix(msg, "#") {
		remoteID := strings.Split(msg, "|")[1]
		if remoteID == "" {
			user.Channel <- "用户ID不能为空！"
			return
		}
		ID, err := strconv.Atoi(remoteID)
		if err != nil {
			user.Channel <- "用户ID必须是整数！"
		}
		user.server.mapLock.Lock()
		remoteUser, ok := user.server.OnlineMap[ID]
		user.server.mapLock.Unlock()
		if !ok {
			m := remoteID + "不存在！"
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

package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {

	Cha chan string
	Name string //名字
	Addr string //地址
	Conn net.Conn // 和服务端连接的 "句柄"
	Server *Server // 服务端
}

//创建用户

func NewUser(conn net.Conn, server *Server) *User {

	userAddr := conn.RemoteAddr().String()

	user := &User{
		Cha: make(chan string),
		Name: userAddr,
		Addr: userAddr,
		Conn: conn,
		Server: server,
	}

	//开启监听自己的信箱
	go user.ListenMessage()

	return user
	
}

//监听当前user的channel,有消息就发送给客户端
func (this *User) ListenMessage() {

	for {
		msg := <- this.Cha
		this.Conn.Write([]byte(msg + "\n"))
	}

}

//上线功能

func (this *User) Online() {

	//上锁
	this.Server.mapLock.Lock()
	this.Server.OnlineMap[this.Name] = this
	this.Server.mapLock.Unlock()

	//boardcast msg
	this.Server.BroadCast(this,"he online !")
	
}

//下线功能
func (this *User) Offline() {

	//上锁
	this.Server.mapLock.Lock()
	delete(this.Server.OnlineMap,this.Name)
	this.Server.mapLock.Unlock()

	//boardcast msg
	this.Server.BroadCast(this,"he offline ! \n")

}

//给当前用户发送消息
func (this *User) sendMsg(msg string) {

	this.Conn.Write([] byte(msg))

}


//用户处理消息的业务
func (this *User) DoMsg(msg string) {

	fmt.Println(msg)

	if msg == "who"{
		//查询在线用户
		for _,cli := range this.Server.OnlineMap{
			onLineMsg := "[" + cli.Addr + "]" + cli.Name + "on line...\n"
			this.sendMsg(onLineMsg)
		}

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式
		newName := strings.Split(msg,"|")[1]

		//判断name是否已经存在

		_,ok := this.Server.OnlineMap[newName]
		if ok{
			//当前name已经存在
			this.sendMsg("this name already exist! \n")
		}else {
			this.Server.mapLock.Lock()
			delete(this.Server.OnlineMap,this.Name)
			this.Server.OnlineMap[newName] = this
			this.Server.mapLock.Unlock()
			this.Name = newName
			this.sendMsg("update ok :" + newName + "\n" )
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//进入私聊模式
		msgArray := strings.Split(msg,"|")
		toName := msgArray[1]
		toMsg:= msgArray[2]
		//获取名字是否存在
		if user,ok := this.Server.OnlineMap[toName] ; ok{
			//名字存在可发送
			user.sendMsg(toMsg + "\n")

		}else {
			//名字不存在
			this.sendMsg(toName + "is not exist! please try again ! Or use command 'who' to see online user ")
		}

	} else {
		//将消息广播
		this.Server.BroadCast(this,msg)
	}

}


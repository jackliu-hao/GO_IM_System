package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip string
	Port int
	//Online_User表
	OnlineMap map[string] *User
	//锁
	mapLock sync.RWMutex
	//广播管道
	Message chan string
}

//创建Server
func NewServer(ip string ,port int) *Server{

	server := &Server{
		Ip: ip,
		Port: port,
		OnlineMap: make(map[string] *User),
		Message: make(chan string),
	}
	return server
}

func (this *Server) Handler(conn net.Conn) {
	//当前连接的业务
	fmt.Println("连接建立成功")
	//用户上线
	//创建用户

	user := NewUser(conn,this)

	//addOnlineMap
	user.Online()

	//用于判断此用户是否存活
	isLive := make(chan bool)

	//接收客户端的消息
	go func() {
		buf := make([]byte,4096)
		for{
			n ,err := conn.Read(buf)
			if n == 0{
				//下线
				user.Offline()
				return
			}
			if err != nil && err != io.EOF{
				fmt.Println("Conn Read err : " ,err)
				return
			}
			//提取消息
			msg := string(buf[:n-1])
			//广播消息
			user.DoMsg(msg)
			isLive <- true
		}
	}()

	//当前handler阻塞
	//select {}
	//实现超时踢人
	for{
		select {
		case <- isLive :
			//当前用户活跃，重置定时器
		case <-time.After(time.Second * 300):
			//超时，直接踢走
			user.sendMsg("you are out ! \n")
			user.Offline()
			close(user.Cha)
			conn.Close()
			//退出当前handler
			return
		}
	}

}

//监听message广播消息
func (this *Server) ListenMessage() {
	for{
		msg := <- this.Message

		//发送msg给在线用户
		this.mapLock.Lock()
		for _,cli := range this.OnlineMap{
			cli.Cha <- msg
		}
		this.mapLock.Unlock()

	}
}


//服务端进行广播
/**
	user  由哪个用户发起
	msg	  发送的信息

 */
func (this *Server) BroadCast(user * User,msg string) {

	sendMsg := "[" + user.Addr + "]" + user.Name + " : " + msg

	this.Message <- sendMsg
}


//启动服务器的接口
func (this *Server) Start() {
	//socket listen
	listener ,error := net.Listen("tcp",fmt.Sprintf("%s:%d",this.Ip, this.Port))

	if error !=nil {
		fmt.Println("error" ,error)
		return
	}
	//close listen socket
	defer listener.Close()

	//启动监听msg的goroutine
	go this.ListenMessage()

	for {
		//accept
		connect ,err := listener.Accept()
		if err != nil{
			fmt.Println("listener accept  err  " , err)
			continue
		}
		//do handler
		go this.Handler(connect)
	}

}








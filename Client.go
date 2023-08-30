package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp string
	ServerPort int
	Name string //客户端名字
	conn net.Conn //生成的连接
	flag int //当前状态，是否需要退出
}

func NewClient(serverIP string  ,serverPort int ) *Client  {

	client := &Client{
		ServerIp: serverIP,
		ServerPort: serverPort,
		flag: 999,
	}
	//连接Server

	conn, err := net.Dial("tcp",
		fmt.Sprintf("%s:%d",client.ServerIp,client.ServerPort))
	if err != nil{
		fmt.Println("连接失败，请重试....")
		return nil
	}
	client.conn = conn

	return client
}

//初始化
var serverIp string
var serverPort int

func init() {
	//client.exe -ip 127.0.0.1
	flag.StringVar(&serverIp,"ip","127.0.0.1",
		"设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort,"port",8080,
		"设置服务器端口(默认是8080)")
}

func (client *Client) menu() bool {

	var flag int

	fmt.Println("1、公聊模式")
	fmt.Println("2、私聊模式")
	fmt.Println("3、更新用户名")
	fmt.Println("-1、退出")

	fmt.Scanln(&flag)

	if flag >= -1 && flag != 0  && flag <= 3{

		client.flag = flag

		return true
	}else {
		fmt.Scanln(">>>输入合法范围内的数字<<<")
		return false
	}

}

func (client *Client) Run() {
	for client.flag != -1{
		for client.menu() != true {

		}
		//根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			//公聊模式
			//fmt.Println("公聊模式")
			client.PublicChat()
			break
		case 2:
			//私聊模式
			//fmt.Println("私聊模式")
			client.privateChat()
			break
		case 3:
			//fmt.Println("更新用户名")
			//更新用户名
			client.updateName()
			break
		}
	}
}

//更新用户名
func (client *Client) updateName() bool {
	fmt.Println(">>>>输入用户名<<<<")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	if _,err := client.conn.Write([]byte(sendMsg)) ; err != nil{
		fmt.Println("conn.Write error " ,err)
		return false
	}
	return true
}

//用于接收服务段发送过来的消息
func (client *Client) DealResponse() {

	//for {
	//	buf := make()
	//	client.conn.Read(buf)
	//}
	//永久阻塞，一但conn有数据，就把数据拷贝到标准输出上
	io.Copy(os.Stdout,client.conn)

}
//公聊模式
func (client *Client) PublicChat() {

	//用户输入的消息
	var chatMsg  string
	fmt.Println(">>输入信息，exit退出")

	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//发送服务器	//消息不为空
		if len(chatMsg) != 0{
			sendMsg := chatMsg + "\n"
			if _ , err := client.conn.Write([]byte(sendMsg)) ;err != nil{
				fmt.Println("数据发送失败！")
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>输入信息，exit退出")

		fmt.Scanln(&chatMsg)

	}

}

//查询当前在线用户
func (client *Client) SelectUsers()  {
	sendMsg := "who\n"
	if _,err := client.conn.Write([]byte(sendMsg));err != nil{
		fmt.Println("查询用户失败",err)
		return
	}

}

//私聊模式
func (client *Client) privateChat() {
	//查询当前用户
	client.SelectUsers()

	//发送的用户名
	var remoteName string
	//发送的信息
	var chatMsg string
	fmt.Println(">>>输入聊天对象的用户名,exit 退出")
	fmt.Scanln(&remoteName)
	//查询当前用户名
	for remoteName != "exit"{
		fmt.Println("输入聊天信息  exit退出")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit"{
			//发送服务器	//消息不为空
			if len(chatMsg) != 0{

				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				if _ , err := client.conn.Write([]byte(sendMsg)) ;err != nil{
					fmt.Println("数据发送失败！")
					break
				}
				chatMsg = ""
				fmt.Println("输入聊天信息  exit退出")
				fmt.Scanln(&chatMsg)
			}
		}
	}

}

func main() {
	//命令行解析
	flag.Parse()

	client := NewClient(serverIp,serverPort)
	if client == nil{
		fmt.Println(">>>>>连接服务器失败<<<<")
		return
	}
	fmt.Println(">>>>连接服务器成功<<<<")

	//开启一个goroutine 处理服务器发来的消息
	go client.DealResponse()
	//启动客户端业务
	client.Run()
}
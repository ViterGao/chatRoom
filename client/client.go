package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

var m_pConn *net.TCPConn

func main() {
	// 启动客户端, 首先需要连接某个服务器
	Connect("127.0.0.1:10086")

	for {
		log.Println("请输入要发送的内容:")
		// 在这里对服务器进行循环发包
		strCommand := ""
		fmt.Scanln(&strCommand)
		SendText(m_pConn, strCommand)
	}
}

func Connect(strAddress string) {
	log.Println("客户端准备连接", strAddress)

	// 端口转换
	tcpAddr, err := net.ResolveTCPAddr("tcp", strAddress)
	if err != nil {
		log.Fatal("错误转换IP地址错误 (ResolveTCPAddr) ", err)
		return
	}
	log.Println("地址转换成功", tcpAddr)

	pConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal("连接TCP失败 (DialTCP)", err)
		return
	}

	szIp := pConn.RemoteAddr().String()
	log.Println("服务端连接成功! IP地址:", szIp)

	m_pConn = pConn
	go func() {
		// 先定义了这个
		defer func() {
			m_pConn = nil
			pConn.Close()
			log.Println("服务器", szIp, "已经断开了连接")
		}()

		for {
			nLen, buf := pack(pConn)
			if nLen < 0 {
				break
			}
			log.Println(buf)
		}
	}()
}

// ================================================
// 下面是公共部分
// 粘包拆包
func pack(pConn *net.TCPConn) (int, []byte) {
	// 前面两个是头
	bufHead := make([]byte, 2)

	// 获取包头长度
	nHeadLen, err := pConn.Read(bufHead)

	if err != nil {
		if err.Error() == "EOF" {
			log.Println("发现EOF异常 (Read) ", err)
		} else {
			log.Println("发现读包异常 (Read) ", err)
		}
		return -1, nil
	}

	if nHeadLen != 2 {
		log.Fatal("发现长度异常 (Read) ", err, "包长长度 = ", nHeadLen)
		return -1, nil
	}

	// 包长
	nPackageLen := BytesToInt16(bufHead)

	// 特殊情况下的空包
	if nPackageLen == 0 {
		return 0, nil
	}

	// 包内容
	buf := make([]byte, nPackageLen)

	// 实际包长
	nRealLen, err := pConn.Read(buf)
	if err != nil {
		log.Fatal("读包发生错误", err)
		return -1, nil
	}

	if nRealLen != nPackageLen {
		log.Fatal("需要粘包. 这里暂时不处理", nRealLen, nPackageLen)
		return -1, nil
	}

	log.Println("收到了包长", nPackageLen, "\n内容是", buf, "\n文字版", string(buf))
	return nPackageLen, buf
}

//字节转换成整形16
func BytesToInt16(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int16
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int(tmp)
}

//
// 发送 BUF
func SendBuf(pConn *net.TCPConn, buf []byte) {
	// 因为实际的BUF 需要补上包长
	realbuf := bytes.NewBuffer([]byte{})
	// 先写上包长
	binary.Write(realbuf, binary.BigEndian, int16(len(buf)))
	// 在写上实际的BUF
	realbuf.Write(buf)
	// 进行发送
	nLen, err := pConn.Write(realbuf.Bytes())
	//
	log.Println("发包成功", realbuf.Bytes(), nLen, err)
}

// 发送文字
func SendText(pConn *net.TCPConn, strText string) {
	SendBuf(pConn, []byte(strText))
}

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

var m_mapConn map[string]*net.TCPConn
var strAddress string

func main() {
	// 首先要设置一个端口, 启动服务器
	Open(":10086")

	// 用来处理命令
	strCommand := ""
	fmt.Scanln(&strCommand)
}

// 打开socket
func Open(strAddress string) {
	log.Println("TBaseTcpServer::open() 服务器准备启动", strAddress)

	// 新建连接队列
	m_mapConn = make(map[string]*net.TCPConn)

	// 端口转换
	tcpAddr, err := net.ResolveTCPAddr("tcp", strAddress)
	if err != nil {
		log.Fatal("TBaseTcpServer::open() 错误转换IP地址错误 (ResolveTCPAddr) ", err)
		return
	}
	log.Println("TBaseTcpServer::open() 地址转换成功", tcpAddr)

	//
	pListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal("TBaseTcpServer::open() 启动TCP端口错误 (ListenTCP) ", err)
		return
	}
	log.Println("TBaseTcpServer::open() 端口启动成功", pListener)

	// 当前第几个连接进来的
	nCount := 0
	for {
		pConn, err := pListener.AcceptTCP()
		if err != nil {
			log.Fatal("TBaseTcpServer::open() 客户端连接异常 (AcceptTCP) ", err)
			continue
		}
		nCount++
		szIp := pConn.RemoteAddr().String()
		log.Println("TBaseTcpServer::open() 客户端连接成功[", nCount, "]IP地址:", szIp)

		go func() {
			// 保存这个连接
			m_mapConn[szIp] = pConn
			defer func() {
				// 回收
				pConn.Close()
				delete(m_mapConn, szIp)

				log.Println("TBaseTcpServer::open() 客户端", szIp, "已经断开了连接")
			}()

			for {
				nLen, buf := pack(pConn)
				if nLen < 0 {
					break
				}

				// 服务器收到了包以后, 要广播给所有客户端
				for _, v := range m_mapConn {
					SendBuf(v, buf)
				}
			}
		}()
	}
}

// ================================================
// 下面两个是公共部分
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

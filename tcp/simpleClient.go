package tcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io"
	"net"
	"strconv"
	"time"
)

/**
调用方式
reqParams := map[string]string{"valid":   "76ed8a8c4ebeab4882414aa519e3850e"}
cli := tcp.NewTcpClient("192.168.204.204", 8912, 15)
cli.SetCall("App\\Service", "ViolationQuery::commonPrefix").SendRequest(15, reqParams)
_, _, _, _, data, err := cli.ReadResponse(15)
if err != nil {
	fmt.Println(err.Error())
}else {
	fmt.Println(string(data))
}
*/

type service struct {
	port        int
	host        string
	namespace   string
	funcName    string
	conn        net.Conn
	dialTimeout time.Duration
	data        []byte
	env         map[string]interface{}
}

func NewTcpClient(host string, port int, timeout int64) *service {
	return &service{
		host:        host,
		port:        port,
		dialTimeout: time.Duration(timeout),
		env:         make(map[string]interface{}),
	}
}

func (s *service) ResetConn(host string, port int) *service {
	s.host = host
	s.port = port
	return s
}

func (s *service) SetCall(namespace, funcName string) *service {
	s.namespace = namespace
	s.funcName = funcName
	return s
}

func (s *service) SetEnv(env map[string]interface{}) *service {
	for k, v := range env {
		s.env[k] = v
	}
	return s
}

func (s *service) PutEnv(key string, value string) *service {
	s.env[key] = value
	return s
}

/*
将数据转成二进制文件,使用大端写法（高位字节放在内存的低地址端，低位字节放在内存的高地址端）
rpctype		2:json,4:swoole,1(或其他):php   framework，RPCServer.php:202
*/
func (s *service) Pack(params interface{}, rpctype uint32, uid uint32, serid uint32) (err error) {
	data := make(map[string]interface{})
	data["call"] = s.namespace + "\\" + s.funcName
	data["env"] = s.env
	data["params"] = params
	dataJson, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(data)
	if err != nil {
		return
	}

	dataBuff := bytes.NewBuffer([]byte{})

	dataLen := uint32(len(dataJson))

	if err = binary.Write(dataBuff, binary.BigEndian, dataLen); err != nil {
		return
	}
	if err = binary.Write(dataBuff, binary.BigEndian, rpctype); err != nil {
		return
	}
	if err = binary.Write(dataBuff, binary.BigEndian, uid); err != nil {
		return
	}
	if err = binary.Write(dataBuff, binary.BigEndian, serid); err != nil {
		return
	}

	head := dataBuff.Bytes()
	s.data = append(head, dataJson...)

	return
}

/*
根据配置建立Dial连接Conn
*/
func (s *service) buildConn() (err error) {
	address := fmt.Sprintf("%s:%s", s.host, strconv.Itoa(s.port))

	s.conn, err = net.DialTimeout("tcp", address, s.dialTimeout*time.Second)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) SendRequest(timeout int64, params ...interface{}) (err error) {
	if err = s.Pack(params, 2, 0, 0); err != nil {
		return
	}

	if err = s.buildConn(); err != nil {
		return err
	}

	_, err = s.conn.Write(s.data)
	return err
}

func (s *service) ReadResponse(timeout int64) (respLen uint32, respType uint32, respUid uint32, respSerid uint32, data []byte, err error) {
	defer func() {
		if err := s.conn.Close();err != nil {
			err = fmt.Errorf("close fails(Error:%s,Data:%s,Env:%s)", err.Error(), s.data, s.env)
		}
	}()

	//设置读取的deadline时间,防止阻塞
	if err = s.conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second));err != nil{
		return 0, 0, 0, 0, nil, err
	}

	read := func(size int, buffData *uint32) {
		if err != nil {
			return
		}
		buff := make([]byte, size)
		_, err = io.ReadFull(s.conn, buff)
		*buffData = binary.BigEndian.Uint32(buff)
	}

	read(4, &respLen)
	read(4, &respType)
	read(4, &respUid)
	read(4, &respSerid)

	if err == nil {
		data = make([]byte, respLen)
		_, err = io.ReadFull(s.conn, data)
	}

	if err != nil {
		switch err.(type) {
		case *net.OpError:
			if err.(*net.OpError).Err.Error() == "i/o timeout" {
				//不排除读丢包导致的阻塞，引起的timeout，这时候部分是有值的
				return
			}
		}
		return 0, 0, 0, 0, nil, err
	}

	//因为服务的外层封装了响应状态码,所以返回的想要数据不能原路返回,需要再加工,拿到内层的data再重新以byte的格式返回,方便调用初解析结构
	var tcpRes struct {
		Errno int                    `json:"errno,omitempty"`
		Data  map[string]interface{} `json:"data"`
	}
	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(data, &tcpRes)
	if err != nil {
		return 0, 0, 0, 0, nil, err
	}

	if tcpRes.Errno != 0 {
		return 0, 0, 0, 0, nil, err
	}
	//回传内层数据
	data, _ = jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(tcpRes.Data)

	return
}
package relay

import (
	"errors"
	"github.com/chenyalyg/ByteBuf"
	"github.com/tarm/serial"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	RequestHeader  = 0x55 //发送帧数据头
	ResponseHeader = 0x22 //接受帧数据头

	//功能码
	RequestReadStatus = 0x10 //读取状态
	RequestCloseOne   = 0x11 //断开某路
	RequestOpenOne    = 0x12 //吸合某路
	RequestRunCMd     = 0x13 //命令执行
	RequestCloseGroup = 0x14 //组断开
	RequestOpenGroup  = 0x15 //组吸合
	RequestFlipGroup  = 0x16 //组翻转
	RequestFlipOne    = 0x20 //翻转某路

	RequestPointOpen  = 0x21 //点动闭合
	RequestPointClose = 0x21 //点动断开

	RequestFlipOneNoReturn  = 0x30 //翻转某路 下位机不返回数据，指令可以连续发送
	RequestCloseOneNoReturn = 0x31 //断开某路
	RequestOpenOneNoReturn  = 0x32 //吸合某路
	RequestRunCMdNoReturn   = 0x33 //命令执行

	RequestCloseGroupNoReturn = 0x34 //组断开
	RequestOpenGroupNoReturn  = 0x35 //组吸合
	RequestFlipGroupNoReturn  = 0x36 //组翻转

	RequestPointOpenNoReturn  = 0x37 //点动闭合
	RequestPointCloseNoReturn = 0x38 //点动断开

	RequestReadAddress  = 0x40 //读地址
	RequestWriteAddress = 0x41 //写地址

	RequestReadVariable  = 0x70 //读变量
	RequestWriteVariable = 0x71 //写变量

	ResponseReadStatus         = 0x10 //读取状态
	ResponseCloseOne           = 0x11 //关闭某一路
	ResponseOpenOne            = 0x12 //打开某一路
	ResponseRunCMd             = 0x13 //命令执行
	ResponseCloseGroup         = 0x14 //组断开
	ResponseOpenGroup          = 0x15 //组吸合
	ResponseFlipGroup          = 0x16 //组翻转
	ResponseModelAddress       = 0x40 //返回模块地址
	ResponseReadInnerVariable  = 0x70 //读内部变量
	ResponseWriteInnerVariable = 0x71 //写内部变量
)

var (
	ErrDisconnected = errors.New("继电器断开连接")
)

var (
	Request  = [8]byte{RequestHeader, 0}
	Response = [8]byte{ResponseHeader, 0}
)

//继电器
type Relay struct {
	conn        *serial.Port
	isConnected bool
	Config      *Config
	Result      chan []byte
	waitExit    *sync.WaitGroup
	Cache       *bytebuf.ByteBuffer
}

//继电器配置
type Config struct {
	Port          int
	Baud          int
	ReadTimeout   time.Duration
	CircuitNumber byte
}

//继电器连接
func (r *Relay) Connect() (*Relay, error) {
	c := &serial.Config{
		Name:        "COM" + strconv.Itoa(r.Config.Port),
		Baud:        r.Config.Baud,
		ReadTimeout: r.Config.ReadTimeout,
	}
	conn, err := serial.OpenPort(c)
	if err != nil {
		log.Println("打开继电器失败：COM", r.Config.Port, ",", err)
		r.isConnected = false
		return nil, err
	} else {
		log.Println("打开继电器成功：COM", r.Config.Port)
		r.isConnected = true
		r.conn = conn
		r.waitExit = &sync.WaitGroup{}
		r.Result = make(chan []byte, 0)
		go r.receive()
		return r, err
	}
}

//数据校验位赋值
func (r *Relay) checkSum(data []byte) []byte {
	sum := byte(0)
	for i := byte(0); i < 7; i += 1 {
		sum += data[i]
	}
	data[7] = 0xff & sum
	return data
}

//发送数据
func (r *Relay) send(data []byte) {
	i, err := r.conn.Write(data)
	if err != nil {
		log.Println("发送数据失败,i=", i, err)
	}
}

//接收数据
func (r *Relay) receive() {
	r.Cache = bytebuf.New(true)
	r.waitExit.Add(1)
	defer r.waitExit.Done()
	buf := make([]byte, 1024)
	for {
		size, err := r.conn.Read(buf)
		if err != nil {
			log.Println("读取继电器数据失败：COM", r.Config.Port)
			r.isConnected = false
			continue
		}
		r.Cache.WriteBytes(buf[0:size])
		for r.Cache.Len() >= int(r.Config.CircuitNumber) {
			buf := make([]byte, r.Config.CircuitNumber)
			r.Cache.ReadBytes(buf)
			sign := buf[7]
			if sign == r.checkSum(buf)[7] {
				r.Result <- buf
			} else {
				log.Println("响应数据校验失败")
			}
		}
		time.Sleep(time.Millisecond * 200)
	}
}

//关闭所有继电器路数
func (r *Relay) CloseAllNoReturn(address byte) {
	if r.isConnected {
		Request[1] = address
		Request[2] = 0x33
		Request[3] = 0x0
		Request[4] = 0x0
		Request[5] = 0x0
		Request[6] = 0x0
		Request[7] = 0x89
		r.send(Request[:])
	}
}

//打开所有继电器路数
func (r *Relay) OpenAllNoReturn(address byte) {
	if r.isConnected {
		Request[1] = address
		Request[2] = 0x33
		Request[3] = 0xff
		Request[4] = 0xff
		Request[5] = 0xff
		Request[6] = 0xff
		r.send(r.checkSum(Request[:]))
		r.send(Request[:])
	}
}

//断开某一路
func (r *Relay) CloseOne(address, index byte) (bool, error) {
	if !r.isConnected {
		return false, ErrDisconnected
	}
	if address < 1 {
		return false, errors.New("继电器地址不能小于1")
	}
	if index < 1 || index > r.Config.CircuitNumber {
		return false, errors.New("继电器路数不能小于1或大于最大路数")
	}
	Request[1] = address
	Request[2] = RequestCloseOne
	Request[3] = 0x00
	Request[4] = 0x00
	Request[5] = 0x00
	Request[6] = index
	r.send(r.checkSum(Request[:]))
	data := <-r.Result
	log.Println(data, data[6], index)
	log.Println(data[6] >> index)
	return data[6]>>index == 0, nil
}

//打开继电器某一路
func (r *Relay) OpenOne(address, index byte) (bool, error) {
	if !r.isConnected {
		return false, ErrDisconnected
	}
	if address < 1 {
		return false, errors.New("继电器地址不能小于1")
	}
	if index < 1 || index > r.Config.CircuitNumber {
		return false, errors.New("继电器路数不能小于1或大于最大路数")
	}
	Request[1] = address
	Request[2] = RequestOpenOne
	Request[3] = 0x00
	Request[4] = 0x00
	Request[5] = 0x00
	Request[6] = index
	r.send(r.checkSum(Request[:]))
	data := <-r.Result
	return data[6]&(1<<(index-1)) == 1<<(index-1), nil
}

//读取继电器路数状态
func (r *Relay) ReadStatus(address byte) []byte {
	if r.isConnected {
		Request[1] = address
		Request[2] = RequestReadStatus
		Request[3] = 0x00
		Request[4] = 0x00
		Request[5] = 0x00
		Request[6] = 0x00
		r.send(r.checkSum(Request[:]))
		data := <-r.Result
		return ByteBinary(data[6])
	}
	return nil
}

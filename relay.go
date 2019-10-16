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
	DataLength     = 0x8
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
	ErrDisconnected  = errors.New("继电器断开连接")
	ErrResponseCode  = errors.New("发送请求与响应数据不匹配")
	ErrAddress       = errors.New("继电器地址错误")
	ErrCircuitNumber = errors.New("继电器路数错误,必须是8的倍数")
)

var (
	request = [DataLength]byte{RequestHeader, 0}
	//response = [DataLength]byte{ResponseHeader, 0}
)

//继电器
type Relay struct {
	conn        *serial.Port
	isConnected bool
	Config      *Config
	result      chan []byte
	waitExit    *sync.WaitGroup
	cache       *bytebuf.ByteBuffer
	address     byte
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
	if r.address < 1 {
		return nil, ErrAddress
	}

	if r.Config.CircuitNumber%8 != 0 {
		return nil, ErrCircuitNumber
	}
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
		r.result = make(chan []byte, 0)
		go r.receive()
		return r, err
	}
}

//设置继电器地址
func (r *Relay) SetAddress(address byte) *Relay {
	if address < 1 {
		panic(ErrAddress)
	}
	r.address = address
	return r
}

//数据校验位赋值
func (r *Relay) checkSum(data []byte) []byte {
	sum := byte(0)
	for i := byte(0); i < DataLength-1; i += 1 {
		sum += data[i]
	}
	data[DataLength-1] = 0xff & sum
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
	r.cache = bytebuf.New(true)
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
		r.cache.WriteBytes(buf[0:size])
		for r.cache.Len() >= DataLength {
			buf := make([]byte, DataLength)
			r.cache.ReadBytes(buf)
			sign := buf[DataLength-1]
			if sign == r.checkSum(buf)[DataLength-1] {
				r.result <- buf
			} else {
				log.Println("响应数据校验失败")
			}
		}
		time.Sleep(time.Millisecond * 200)
	}
}

//关闭所有继电器路数
func (r *Relay) CloseAllNoReturn() error {
	if !r.isConnected {
		return ErrDisconnected
	}
	request[1] = r.address
	request[2] = 0x33
	request[3] = 0x0
	request[4] = 0x0
	request[5] = 0x0
	request[6] = 0x0
	r.send(r.checkSum(request[:]))
	return nil
}

//打开所有继电器路数
func (r *Relay) OpenAllNoReturn() error {
	if !r.isConnected {
		return ErrDisconnected
	}
	request[1] = r.address
	request[2] = 0x33
	request[3] = 0xff
	request[4] = 0xff
	request[5] = 0xff
	request[6] = 0xff
	r.send(r.checkSum(request[:]))
	return nil
}

//断开某一路
func (r *Relay) CloseOne(index byte) (bool, error) {
	if !r.isConnected {
		return false, ErrDisconnected
	}
	if index < 1 || index > r.Config.CircuitNumber {
		return false, errors.New("继电器路数不能小于1或大于最大路数")
	}
	request[1] = r.address
	request[2] = RequestCloseOne
	request[3] = 0x00
	request[4] = 0x00
	request[5] = 0x00
	request[6] = index
	r.send(r.checkSum(request[:]))
	data := <-r.result
	if data[2] == ResponseCloseOne {
		return data[6]>>index == 0, nil
	}
	return false, ErrResponseCode
}

//打开继电器某一路
func (r *Relay) OpenOne(index byte) (bool, error) {
	if !r.isConnected {
		return false, ErrDisconnected
	}
	if index < 1 || index > r.Config.CircuitNumber {
		return false, errors.New("继电器路数不能小于1或大于最大路数")
	}
	request[1] = r.address
	request[2] = RequestOpenOne
	request[3] = 0x00
	request[4] = 0x00
	request[5] = 0x00
	request[6] = index
	r.send(r.checkSum(request[:]))
	data := <-r.result
	if data[2] == ResponseOpenOne {
		return data[6]&(1<<(index-1)) == 1<<(index-1), nil
	}
	return false, ErrResponseCode
}

//以字节数组命令运行
func (r *Relay) RunCMd(circuits []byte) ([]byte, error) {
	if !r.isConnected {
		return nil, ErrDisconnected
	}
	if len(circuits) != int(r.Config.CircuitNumber) {
		return nil, errors.New("参数长度必须等于继电器路数")
	}
	request[1] = r.address
	request[2] = RequestRunCMd
	request[3] = 0x00
	request[4] = 0x00
	request[5] = 0x00
	switch r.Config.CircuitNumber {
	case 32:
		request[3] = BinaryByte(circuits[0:8])
		request[4] = BinaryByte(circuits[8:16])
		request[5] = BinaryByte(circuits[16:24])
		request[6] = BinaryByte(circuits[24:32])
	case 24:
		request[4] = BinaryByte(circuits[0:8])
		request[5] = BinaryByte(circuits[8:16])
		request[6] = BinaryByte(circuits[16:24])
	case 16:
		request[5] = BinaryByte(circuits[0:8])
		request[6] = BinaryByte(circuits[8:16])
	case 8:
		request[6] = BinaryByte(circuits[0:8])
	}
	r.send(r.checkSum(request[:]))
	data := <-r.result
	if data[2] == ResponseRunCMd {
		result := make([]byte, r.Config.CircuitNumber)
		switch r.Config.CircuitNumber {
		case 32:
			result = append(result, ByteBinary(data[3])...)
			result = append(result, ByteBinary(data[4])...)
			result = append(result, ByteBinary(data[5])...)
			result = append(result, ByteBinary(data[6])...)
		case 24:
			result = append(result, ByteBinary(data[4])...)
			result = append(result, ByteBinary(data[5])...)
			result = append(result, ByteBinary(data[6])...)
		case 16:
			result = append(result, ByteBinary(data[5])...)
			result = append(result, ByteBinary(data[6])...)
		case 8:
			result = append(result, ByteBinary(data[6])...)
		}
		return result, nil
	}
	return nil, ErrResponseCode
}

//读取继电器路数状态
func (r *Relay) ReadStatus() ([]byte, error) {
	if !r.isConnected {
		return nil, ErrDisconnected
	}
	request[1] = r.address
	request[2] = RequestReadStatus
	request[3] = 0x00
	request[4] = 0x00
	request[5] = 0x00
	request[6] = 0x00
	r.send(r.checkSum(request[:]))
	data := <-r.result
	if data[2] == ResponseReadStatus {
		var result []byte
		switch r.Config.CircuitNumber {
		case 32:
			result = append(result, ByteBinary(data[3])...)
			result = append(result, ByteBinary(data[4])...)
			result = append(result, ByteBinary(data[5])...)
			result = append(result, ByteBinary(data[6])...)
		case 24:
			result = append(result, ByteBinary(data[4])...)
			result = append(result, ByteBinary(data[5])...)
			result = append(result, ByteBinary(data[6])...)
		case 16:
			result = append(result, ByteBinary(data[5])...)
			result = append(result, ByteBinary(data[6])...)
		case 8:
			result = append(result, ByteBinary(data[6])...)
		}
		return result, nil
	}
	return nil, ErrResponseCode
}

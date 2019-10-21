package relay_test

import (
	"github.com/zhangrxiang/relay-xk"
	"log"
	"testing"
)

func conn() relay.Relay {
	r := relay.Relay{
		Config: &relay.Config{
			Port:          3,
			Baud:          9600,
			ReadTimeout:   10,
			CircuitNumber: 16,
		},
	}
	_, err := r.SetAddress(1).Connect()
	if err != nil {
		panic("connect error")
	}
	return r
}
func TestRelay_Connect(t *testing.T) {
	conn()
}
func TestRelay_CloseAllNoReturn(t *testing.T) {
	r := conn()
	_ = r.CloseAllNoReturn()
}

func openOne(r relay.Relay) {
	a1, _ := r.OpenOne(1)
	a2, _ := r.OpenOne(2)
	a3, _ := r.OpenOne(3)
	a4, _ := r.OpenOne(4)
	a5, _ := r.OpenOne(5)
	log.Println(a1, a2, a3, a4, a5)
}

func TestRelay_OpenOne(t *testing.T) {
	openOne(conn())
}

func closeOne(r relay.Relay) {
	a1, _ := r.CloseOne(1)
	a2, _ := r.CloseOne(2)
	a3, _ := r.CloseOne(3)
	a4, _ := r.CloseOne(4)
	a5, _ := r.CloseOne(5)
	log.Println(a1, a2, a3, a4, a5)
}

func TestRelay_CloseOne(t *testing.T) {
	r := conn()
	openOne(r)
	closeOne(r)
	openOne(r)
	closeOne(r)
}

func readStatus(r relay.Relay) {
	openOne(r)
	data, _ := r.ReadStatus()
	log.Println(data)
	closeOne(r)
	data, _ = r.ReadStatus()
	log.Println(data)
}

func TestRelay_ReadStatus(t *testing.T) {
	r := conn()
	//readStatus(4)
	data, _ := r.ReadStatus()
	log.Println(data)
}

func runCMd(r relay.Relay) {
	data, _ := r.RunCMd([]byte{0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1})
	log.Println(data)
	data, _ = r.RunCMd([]byte{1, 1, 0, 0, 0, 1, 1, 0})
	log.Println(data)
	data, _ = r.RunCMd([]byte{0, 0, 0, 0, 0, 0, 0, 1})
	log.Println(data)
}

func TestRelay_RunCMd(t *testing.T) {
	runCMd(conn())
}

func TestByteBinary(t *testing.T) {
	log.Println(relay.ByteBinary(6))
}

func TestBinaryByte(t *testing.T) {
	log.Println(relay.BinaryByte([]byte{0, 0, 0, 0, 0, 0, 0, 0}))
	log.Println(relay.BinaryByte([]byte{0, 0, 0, 0, 0, 0, 0, 1}))
	log.Println(relay.BinaryByte([]byte{0, 0, 0, 0, 1, 0, 0, 0}))
	log.Println(relay.BinaryByte([]byte{0, 0, 0, 0, 0, 0, 1, 1}))
	log.Println(relay.BinaryByte([]byte{1, 0, 0, 0, 0, 0, 0, 0}))
}

func TestRelay_FlipOne(t *testing.T) {
	r := conn()
	//r.CloseAllNoReturn()
	status, _ := r.FlipOne(7)
	log.Println(status)
}

func TestName(t *testing.T) {
	log.Println(1 << 7)
	log.Println(7&(1<<(25-25)) == 1<<(25-25))
	log.Println(3>>1 == 1)
	log.Println(3&(1<<1) == 1<<1)
	log.Println(5&(1<<2) == 1<<2)
	log.Println(5>>2 == 1)
	log.Println(5&(1<<2)>>2 == 1)
	log.Println(1 >> 1)
	log.Println(1 >> 0)
}

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

func openOne(r relay.Relay) {
	_, _ = r.OpenOne(1)
	_, _ = r.OpenOne(2)
}
func TestRelay_OpenOne(t *testing.T) {
	openOne(conn())
}

func closeOne(r relay.Relay) {
	_, _ = r.CloseOne(1)
	_, _ = r.CloseOne(2)
}

func TestRelay_CloseOne(t *testing.T) {
	closeOne(conn())
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

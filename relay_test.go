package relay_test

import (
	"github.com/zhangrxiang/relay-xk"
	"testing"
)

func conn() relay.Relay {
	r := relay.Relay{
		Config: &relay.Config{
			Port:          3,
			Baud:          9600,
			ReadTimeout:   10,
			CircuitNumber: 8,
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

func TestRelay_OpenOne(t *testing.T) {
	r := conn()
	_, _ = r.OpenOne(1)
	_, _ = r.OpenOne(2)
}

func TestRelay_CloseOne(t *testing.T) {
	r := conn()
	_, _ = r.CloseOne(1)
	_, _ = r.CloseOne(2)
}

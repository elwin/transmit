package scion

import (
	"fmt"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/squic"
)

func Listen(address string) (Listener, error) {

	addr, _ := snet.AddrFromString(address)

	err := initNetwork(*addr)
	if err != nil {
		return nil, err
	}

	listener, err := squic.ListenSCION(nil, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to listen: %s", err)
	}

	return &ScionListener{
		listener,
		*addr,
	}, nil
}

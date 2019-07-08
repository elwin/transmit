package scion

import (
	"fmt"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/squic"
)

func Listen(address string) (Listener, error) {

	addr, _ := snet.AddrFromString(address)

	// Initialize Network if uninitialized
	if snet.DefNetwork == nil {

		sciond := sciond.GetDefaultSCIONDPath(&addr.IA)
		dispatcher := ""

		err := snet.Init(addr.IA, sciond, dispatcher)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize SCION: %s", err)
		}
	}

	// Initialize SQUIC
	err := squic.Init("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to initilaze SQUIC: %s", err)
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

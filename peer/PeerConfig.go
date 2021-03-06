package peer

import (
	"copernicus/protocol"
	"copernicus/network"
)

type PeerConfig struct {
	NewBlock          protocol.HashFunc
	HostToAddressFunc protocol.HostToNetAddrFunc
	BestAddress       network.PeerAddressFunc

	Proxy            string
	UserAgent        string
	UserAgentVersion string

	// BIP 14
	UserAgentComments []string

	ServicesFlag    protocol.ServiceFlag
	ProtocolVersion uint32
	DisableRelayTx  bool
	Listener        MessageListener
	ChainParams *protocol.BitcoinParams
}

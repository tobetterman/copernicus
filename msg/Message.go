package msg

import (
	"copernicus/protocol"
	"io"
)

const (
	// MessageHeaderSize is the number of bytes in a bitcoin msg header.
	// Bitcoin network (magic) 4 bytes + command 12 bytes + payload length 4 bytes +
	// checksum 4 bytes.
	MESSAGE_HEADER_SIZE = 24

)

type Message struct {
	Net      protocol.BitcoinNet
	Command  string
	Length   uint32
	Checksum [4]byte
}

func (msg*Message) BitcoinParse(reader io.Reader, size uint32) error {
	return nil
}
func (msg *Message) BitcoinSerialize(writer io.Writer, size uint32) error {
	return nil
}
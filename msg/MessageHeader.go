package msg

import (
	"copernicus/protocol"
	"io"
	"bytes"
	"copernicus/utils"
	"fmt"
	"github.com/pkg/errors"
	"copernicus/crypto"
	"unicode/utf8"
)

const (
	MESSAGE_HEADER_SIZE = 24
)

type MessageHeader struct {
	Net      protocol.BitcoinNet // 4 bytes
	Command  string              // 12 bytes
	Length   uint32              // 4 bytes
	Checksum [4]byte             // 4 bytes
}

func ReadMessageHeader(reader io.Reader) (int, *MessageHeader, error) {
	var headerBytes [MESSAGE_HEADER_SIZE]byte
	n, err := io.ReadFull(reader, headerBytes[:])
	if err != nil {
		return n, nil, err
	}
	header := bytes.NewReader(headerBytes[:])
	hdr := MessageHeader{}
	var command [COMMAND_SIZE]byte
	utils.ReadElements(header, &hdr.Net, &command, &hdr.Length, &hdr.Checksum)
	hdr.Command = string(bytes.TrimRight(command[:], string(0)))
	return n, &hdr, nil

}
func DiscardInput(reader io.Reader, n uint32) {
	maxSize := uint32(10 * 1024)
	numReads := n / maxSize
	bytesRemaining := n % maxSize
	if n > 0 {
		buf := make([]byte, maxSize)
		for i := uint32(0); i < numReads; i++ {
			io.ReadFull(reader, buf)
		}
	}
	if bytesRemaining > 0 {
		buf := make([]byte, bytesRemaining)
		io.ReadFull(reader, buf)
	}

}
func WriteMessage(w io.Writer, message Message, pver uint32, net protocol.BitcoinNet) (int, error) {
	totalBytes := 0
	var command [COMMAND_SIZE]byte
	cmd := message.Command()
	if len(cmd) > COMMAND_SIZE {
		str := fmt.Sprintf("command %s is too long max %v", cmd, COMMAND_SIZE)
		return totalBytes, errors.New(str)

	}
	copy(command[:], []byte(cmd))
	var buf bytes.Buffer
	err := message.BitcoinParse(&buf, pver)
	if err != nil {
		return totalBytes, err
	}
	payload := buf.Bytes()
	payloadLength := len(payload)
	if payloadLength > protocol.MAX_MESSAGE_PAYLOAD {
		errStr := fmt.Sprintf("message payload is too large -encoed %d bytes ,but maximum message payload is %d bytes",
			payloadLength, protocol.MAX_MESSAGE_PAYLOAD)
		return totalBytes, errors.New(errStr)
	}
	maxPayloadLength := message.MaxPayloadLength(pver)
	if uint32(payloadLength) > maxPayloadLength {
		errStr := fmt.Sprintf("message payload is too large - encode %d bytes ,but maximum message payload size ,type:%s ,%d",
			payloadLength, cmd, maxPayloadLength)
		return totalBytes, errors.New(errStr)
	}
	messageHeader := MessageHeader{Net: net, Command: cmd, Length: uint32(payloadLength)}
	copy(messageHeader.Checksum[:], crypto.DoubleSha256Bytes(payload)[0:4])
	headerBuf := bytes.NewBuffer(make([]byte, 0, MESSAGE_HEADER_SIZE))
	utils.WriteElements(headerBuf, messageHeader.Net, command, messageHeader.Length, messageHeader.Checksum)
	n, err := w.Write(headerBuf.Bytes())
	totalBytes += n
	if err != nil {
		return totalBytes, err
	}
	n, err = w.Write(payload)
	totalBytes += n
	if err != nil {
		return totalBytes, err
	}
	return totalBytes, nil
}

func ReadMessage(reader io.Reader, pver uint32, bitcoinNet protocol.BitcoinNet) (int, Message, []byte, error) {
	totalBytes := 0
	n, messageHeader, err := ReadMessageHeader(reader)
	totalBytes += n
	if err != nil {
		return totalBytes, nil, nil, err
	}
	if messageHeader.Length > protocol.MAX_MESSAGE_PAYLOAD {
		str := fmt.Sprintf("message payload is too large -header indicates %d bytes, but max message payload is %d bytes ",
			messageHeader.Length, protocol.MAX_MESSAGE_PAYLOAD)
		return totalBytes, nil, nil, errors.New(str)

	}
	if messageHeader.Net != bitcoinNet {
		DiscardInput(reader, messageHeader.Length)
		str := fmt.Sprintf("message from other metwork %v", messageHeader.Net)
		return totalBytes, nil, nil, errors.New(str)
	}
	command := messageHeader.Command
	if !utf8.ValidString(command) {
		DiscardInput(reader, messageHeader.Length)
		str := fmt.Sprintf("invalid command %v", []byte(command))
		return totalBytes, nil, nil, errors.New(str)
	}
	message, err := makeEmptyMessage(command)
	if err != nil {
		DiscardInput(reader, messageHeader.Length)
		return totalBytes, nil, nil, err
	}
	maxPayloadLength := message.MaxPayloadLength(pver)
	if messageHeader.Length > maxPayloadLength {
		DiscardInput(reader, messageHeader.Length)
		str := fmt.Sprintf("payload exceeds max length -header indicates %v bytes,"+
			"but max payload size for messages of type %v is %v",
			messageHeader.Length, command, maxPayloadLength)
		return totalBytes, nil, nil, errors.New(str)
	}
	payload := make([]byte, messageHeader.Length)
	n, err = io.ReadFull(reader, payload)
	totalBytes += n
	if err != nil {
		return totalBytes, nil, nil, err
	}
	checksum := crypto.DoubleSha256Bytes(payload)[0:4]
	if !bytes.Equal(checksum[:], messageHeader.Checksum[:]) {
		str := fmt.Sprintf("payload checksum failed header indicates %v ,but actual checksum is %v",
			messageHeader.Checksum, checksum)
		return totalBytes, nil, nil, errors.New(str)
	}
	payloadBuf := bytes.NewBuffer(payload)
	err = message.BitcoinParse(payloadBuf, pver)
	if err != nil {
		return totalBytes, nil, nil, err
	}
	return totalBytes, message, payload, nil

}

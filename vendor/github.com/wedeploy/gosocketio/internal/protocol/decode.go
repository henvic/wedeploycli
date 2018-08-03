package protocol

import (
	"fmt"
	"strconv"
	"strings"
)

// Decode message.
func Decode(source []byte) (msg *Message, err error) {
	data := string(source)

	msg = &Message{
		Source: data,
	}

	if len(data) == 0 {
		return nil, ErrorWrongMessageType
	}

	var majorType = data[0:1]

	if majorType == OpenMessage {
		msg.Type = MessageTypeOpen

		if len(data) > 1 {
			msg.Data = []byte(data[1:])
		}

		return msg, nil
	}

	msg.Type, err = getMessageType(data)

	if err != nil {
		return nil, err
	}

	if msg.Type == MessageTypeEmpty {
		msg.Method = OnConnection
	}

	if len(data) > 2 {
		data = data[2:]
		msg.Namespace, data = extractNamespace(data)
	}

	switch msg.Type {
	case MessageTypeClose,
		MessageTypePing,
		MessageTypePong,
		MessageTypeEmpty,
		MessageTypeError:
		return msg, nil
	}

	if majorType != RegularMessage {
		return nil, fmt.Errorf("can't decode message type %v", majorType)
	}

	if msg.Type == MessageTypeAckResponse {
		ack, rest, err := getAckFromPacket(data)

		if err != nil {
			return nil, err
		}

		msg.AckID = ack
		msg.Data = []byte(rest)
		return msg, nil
	}

	msg.Type = MessageTypeEmit
	msg.Method, msg.Data, err = decodePacket(data)

	return msg, err
}

func extractNamespace(data string) (namespace string, rest string) {
	var pos int

	if len(data) == 0 {
		return "", ""
	}

	for i, c := range data {
		if c == ',' {
			pos = i
			break
		}

		if c == '"' {
			return "", data
		}
	}

	namespace = data[0:pos]

	if len(data) > pos+1 {
		rest = data[pos+1:]
	}

	return namespace, rest
}

func getMessageType(data string) (string, error) {
	switch data[0:1] {
	case OpenMessage:
		return MessageTypeOpen, nil
	case CloseMessage:
		return MessageTypeClose, nil
	case PingMessage:
		return MessageTypePing, nil
	case PongMessage:
		return MessageTypePong, nil
	case RegularMessage:
		return getRegularMessageType(data)
	}

	return "", ErrorWrongMessageType
}

func getRegularMessageType(data string) (string, error) {
	if len(data) == 1 {
		return "", ErrorWrongMessageType
	}

	switch data[0:2] {
	case NamespaceClose:
		return MessageTypeClose, nil
	case EmptyMessage:
		return MessageTypeEmpty, nil
	case CommonMessage:
		return MessageTypeAckRequest, nil
	case AckMessage:
		return MessageTypeAckResponse, nil
	case ErrorMessage:
		return MessageTypeError, nil
	}

	return "", ErrorWrongMessageType
}

func getAckFromPacket(text string) (ackID int, restText string, err error) {
	if len(text) < 2 {
		return 0, "", ErrorWrongPacket
	}

	pos := strings.IndexByte(text, '[')

	if pos == -1 {
		return 0, "", ErrorWrongPacket
	}

	ack, err := strconv.Atoi(text[0:pos])
	if err != nil {
		return 0, "", err
	}

	return ack, text[pos:], nil
}

func decodePacket(input string) (method string, packet []byte, err error) {
	var start, end, rest, countQuote int

	for i, c := range input {
		if c == '"' {
			switch countQuote {
			case 0:
				start = i + 1
			case 1:
				end = i
				rest = i + 1
			default:
				return "", []byte(""), ErrorWrongPacket
			}

			countQuote++
		}

		if c == ',' {
			if countQuote < 2 {
				continue
			}

			rest = i + 1
			break
		}
	}

	if (end < start) || (rest >= len(input)) {
		return "", []byte(""), ErrorWrongPacket
	}

	b := append([]byte("["), []byte(input[rest:])...)
	return input[start:end], b, nil
}

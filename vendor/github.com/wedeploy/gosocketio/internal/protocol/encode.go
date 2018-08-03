package protocol

import (
	"encoding/json"
	"fmt"
)

// Encode message.
func Encode(msg *Message, args ...interface{}) (packet string, err error) {
	// preventing json/encoding "index out of range" panic
	defer func() {
		if r := recover(); r != nil && err == nil {
			err = r.(error)
		}
	}()

	result, err := typeToText(msg.Type)
	if err != nil {
		return "", err
	}

	if msg.Type == MessageTypeEmpty || msg.Type == MessageTypePing ||
		msg.Type == MessageTypePong {
		return result, nil
	}

	if msg.Type == MessageTypeAckRequest || msg.Type == MessageTypeAckResponse {
		result += fmt.Sprintf("%v", msg.AckID)
	}

	if msg.Type == MessageTypeOpen || msg.Type == MessageTypeClose {
		return fmt.Sprintf("%s%s", result, msg.Data), nil
	}

	if msg.Type == MessageTypeAckResponse {
		return fmt.Sprintf("%s%s", result, msg.Data), nil
	}

	if msg.Type == MessageTypeNamespace {
		return fmt.Sprintf("%s%s", result, msg.Method), nil
	}

	if args == nil {
		return fmt.Sprintf(`%s%s,["%s"]`, result, msg.Namespace, msg.Method), nil
	}

	args = append([]interface{}{msg.Method}, args...)
	json, err := json.Marshal(&args)

	if err != nil {
		return "", err
	}

	var format = `%s%s,%s`

	if msg.Namespace == "" {
		format = `%s%s%s`
	}

	packet = fmt.Sprintf(format, result, msg.Namespace, json)
	return packet, err
}

func typeToText(msgType string) (string, error) {
	switch msgType {
	case MessageTypeOpen:
		return OpenMessage, nil
	case MessageTypeClose:
		return CloseMessage, nil
	case MessageTypePing:
		return PingMessage, nil
	case MessageTypePong:
		return PongMessage, nil
	case MessageTypeEmpty, MessageTypeNamespace:
		return EmptyMessage, nil
	case MessageTypeEmit, MessageTypeAckRequest:
		return CommonMessage, nil
	case MessageTypeAckResponse:
		return AckMessage, nil
	}

	return "", ErrorWrongMessageType
}

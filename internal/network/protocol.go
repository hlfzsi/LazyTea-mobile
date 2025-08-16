package network
import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)
const (
	ProtocolVersion = "1.0"
	Separator       = "\x1e"  
)
type MessageHeader struct {
	MsgID         string  `json:"msg_id"`
	MsgType       string  `json:"msg_type"`  
	CorrelationID *string `json:"correlation_id,omitempty"`
	Timestamp     float64 `json:"timestamp"`
}
type ProtocolMessage struct {
	Version string        `json:"version"`
	Header  MessageHeader `json:"header"`
	Payload interface{}   `json:"payload"`
}
type RequestPayload struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}
type ResponsePayload struct {
	Code  int         `json:"code"`
	Time  int64       `json:"time"`
	Data  interface{} `json:"data"`
	Error *string     `json:"error,omitempty"`
}
type HeartbeatPayload struct {
	Status string `json:"status"`
}
func EncodeMessage(header MessageHeader, payload interface{}) (string, error) {
	message := ProtocolMessage{
		Version: ProtocolVersion,
		Header:  header,
		Payload: payload,
	}
	data, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}
	return string(data) + Separator, nil
}
func DecodeMessage(rawData string) (*MessageHeader, interface{}, error) {
	msgData := strings.TrimSuffix(rawData, Separator)
	var message ProtocolMessage
	if err := json.Unmarshal([]byte(msgData), &message); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return &message.Header, message.Payload, nil
}
func NewRequestHeader(msgID string, correlationID *string) MessageHeader {
	return MessageHeader{
		MsgID:         msgID,
		MsgType:       "request",
		CorrelationID: correlationID,
		Timestamp:     float64(time.Now().UnixNano()) / 1e9,
	}
}
func NewResponseHeader(msgID string, correlationID string) MessageHeader {
	return MessageHeader{
		MsgID:         msgID,
		MsgType:       "response",
		CorrelationID: &correlationID,
		Timestamp:     float64(time.Now().UnixNano()) / 1e9,
	}
}

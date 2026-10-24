// Package tunnel 实现多路复用 WebSocket 隧道协议。
// 帧格式：1 字节类型 + 4 字节 streamId + 4 字节头长度 + JSON 头 + 二进制载荷，整数均为大端。
package tunnel

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

// 帧类型与客户端 Dart 实现保持相同数值，不能随意改动。
const (
	TypeHello        byte = 0
	TypeWelcome      byte = 1
	TypePing         byte = 2
	TypePong         byte = 3
	TypeConfigUpdate byte = 4
	TypeHTTPReq      byte = 5
	TypeHTTPResp     byte = 6
	TypeBodyChunk    byte = 7
	TypeStreamError  byte = 8
	TypeCancel       byte = 9
	TypeWSOpen       byte = 10
	TypeWSOpenAck    byte = 11
	TypeWSMsg        byte = 12
	TypeWSClose      byte = 13
)

const (
	maxHeaderSize  = 256 * 1024
	maxPayloadSize = 64 * 1024
	frameOverhead  = 9
	ChunkSize      = 32 * 1024
)

// Frame 一条隧道消息。Header 必须是 JSON 对象；无头时编码为 {}。
type Frame struct {
	Type     byte
	StreamID uint32
	Header   json.RawMessage
	Payload  []byte
}

// EncodeFrame 将帧编码为 WebSocket 二进制消息。
func EncodeFrame(f Frame) ([]byte, error) {
	header := f.Header
	if len(header) == 0 {
		header = json.RawMessage(`{}`)
	}
	if len(header) > maxHeaderSize {
		return nil, fmt.Errorf("帧头过大: %d", len(header))
	}
	if len(f.Payload) > maxPayloadSize {
		return nil, fmt.Errorf("帧载荷过大: %d", len(f.Payload))
	}
	buf := make([]byte, frameOverhead+len(header)+len(f.Payload))
	buf[0] = f.Type
	binary.BigEndian.PutUint32(buf[1:5], f.StreamID)
	binary.BigEndian.PutUint32(buf[5:9], uint32(len(header)))
	copy(buf[9:], header)
	copy(buf[9+len(header):], f.Payload)
	return buf, nil
}

// DecodeFrame 解析二进制消息。长度字段必须与剩余字节一致。
func DecodeFrame(data []byte) (Frame, error) {
	if len(data) < frameOverhead {
		return Frame{}, fmt.Errorf("帧过短")
	}
	headerLen := int(binary.BigEndian.Uint32(data[5:9]))
	if headerLen < 0 || headerLen > maxHeaderSize {
		return Frame{}, fmt.Errorf("帧头长度非法: %d", headerLen)
	}
	if len(data) < frameOverhead+headerLen {
		return Frame{}, fmt.Errorf("帧头截断")
	}
	payload := data[frameOverhead+headerLen:]
	if len(payload) > maxPayloadSize {
		return Frame{}, fmt.Errorf("帧载荷过大: %d", len(payload))
	}
	header := json.RawMessage(append([]byte(nil), data[9:9+headerLen]...))
	out := Frame{
		Type:     data[0],
		StreamID: binary.BigEndian.Uint32(data[1:5]),
		Header:   header,
		Payload:  append([]byte(nil), payload...),
	}
	return out, nil
}

// MustHeader 将结构编码进帧头，失败时返回空对象。调用方应优先用 HeaderFrom。
func HeaderFrom(v any) (json.RawMessage, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("编码帧头: %w", err)
	}
	return json.RawMessage(b), nil
}

// UnmarshalHeader 解析帧头。
func UnmarshalHeader(f Frame, v any) error {
	if len(f.Header) == 0 {
		return fmt.Errorf("帧头为空")
	}
	if err := json.Unmarshal(f.Header, v); err != nil {
		return fmt.Errorf("解析帧头: %w", err)
	}
	return nil
}

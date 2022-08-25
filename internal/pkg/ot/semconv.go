package ot

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	//RPCNameKey Name of message transmitted or received.
	RPCNameKey = attribute.Key("name")

	//RPCMessageDeliveryKey Type of message transmitted or received.
	RPCMessageDeliveryKey = attribute.Key("message.type")

	//RPCMessageIDKey Identifier of message transmitted or received.
	RPCMessageIDKey = attribute.Key("message.id")

	//RPCMessageCompressedSizeKey The compressed size of the message transmitted or received in bytes.
	RPCMessageCompressedSizeKey = attribute.Key("message.compressed_size")

	//RPCMessageSizeKey The uncompressed size of the message transmitted or received in bytes.
	RPCMessageSizeKey = attribute.Key("message.size")

	//semconv.NetPeerIPKey
)

const (
	//NetPeerUnixSocketKey ...
	NetPeerUnixSocketKey = attribute.Key("net.peer.unix_socket")
)

// Semantic conventions for common RPC attributes.
var (
	//RPCSystemGRPC semantic convention for gRPC as the remoting system.
	RPCSystemGRPC = semconv.RPCSystemKey.String("grpc")

	//RPCNameMessage semantic convention for a message named message.
	RPCNameMessage = RPCNameKey.String("message")

	//RPCMessageSent semantic conventions for RPC message types.
	RPCMessageSent = RPCMessageDeliveryKey.String("SENT")

	//RPCMessageReceived semantic conventions for RPC message types.
	RPCMessageReceived = RPCMessageDeliveryKey.String("RECEIVED")
)

var (
	_ = RPCSystemGRPC
	_ = RPCNameMessage
	_ = RPCMessageSent
	_ = RPCMessageReceived
)

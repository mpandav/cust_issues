package tibems

/*
#include <tibems/tibems.h>
*/
import "C"

// SSLEncodingType specifies the format a TLS certificate or key is given in
type SSLEncodingType int32

const (
	SslEncodingAuto     = SSLEncodingType(0x0000)
	SslEncodingPem      = SSLEncodingType(0x0001)
	SslEncodingDer      = SSLEncodingType(0x0002)
	SslEncodingBer      = SSLEncodingType(0x0004)
	SslEncodingPkcs7    = SSLEncodingType(0x0010)
	SslEncodingPkcs8    = SSLEncodingType(0x0020)
	SslEncodingPkcs12   = SSLEncodingType(0x0040)
	SslEncodingEntrust  = SSLEncodingType(0x0100)
	SslEncodingKeystore = SSLEncodingType(0x0200)
)

const DefaultBroker = "tcp://localhost:7222"

type FactoryLoadBalanceMetric int32

const (
	FactoryLoadBalanceMetricNone        = FactoryLoadBalanceMetric(0)
	FactoryLoadBalanceMetricConnections = FactoryLoadBalanceMetric(1)
	FactoryLoadBalanceMetricByteRate    = FactoryLoadBalanceMetric(2)
)

const (
	tibems_OK = C.tibems_status(0)

	tibems_ILLEGAL_STATE       = C.tibems_status(1)
	tibems_INVALID_CLIENT_ID   = C.tibems_status(2)
	tibems_INVALID_DESTINATION = C.tibems_status(3)
	tibems_INVALID_SELECTOR    = C.tibems_status(4)

	tibems_EXCEPTION          = C.tibems_status(5)
	tibems_SECURITY_EXCEPTION = C.tibems_status(6)

	tibems_MSG_EOF = C.tibems_status(7)

	tibems_MSG_NOT_READABLE  = C.tibems_status(9)
	tibems_MSG_NOT_WRITEABLE = C.tibems_status(10)

	tibems_SERVER_NOT_CONNECTED = C.tibems_status(11)
	tibems_VERSION_MISMATCH     = C.tibems_status(12)
	tibems_SUBJECT_COLLISION    = C.tibems_status(13)

	tibems_INVALID_PROTOCOL = C.tibems_status(15)
	tibems_INVALID_HOSTNAME = C.tibems_status(17)
	tibems_INVALID_PORT     = C.tibems_status(18)
	tibems_NO_MEMORY        = C.tibems_status(19)
	tibems_INVALID_ARG      = C.tibems_status(20)

	tibems_SERVER_LIMIT = C.tibems_status(21)

	tibems_MSG_DUPLICATE = C.tibems_status(22)

	tibems_SERVER_DISCONNECTED = C.tibems_status(23)
	tibems_SERVER_RECONNECTING = C.tibems_status(24)

	tibems_NOT_PERMITTED = C.tibems_status(27)

	tibems_SERVER_RECONNECTED = C.tibems_status(28)

	tibems_INVALID_NAME      = C.tibems_status(30)
	tibems_INVALID_TYPE      = C.tibems_status(31)
	tibems_INVALID_SIZE      = C.tibems_status(32)
	tibems_INVALID_COUNT     = C.tibems_status(33)
	tibems_NOT_FOUND         = C.tibems_status(35)
	tibems_ID_IN_USE         = C.tibems_status(36)
	tibems_ID_CONFLICT       = C.tibems_status(37)
	tibems_CONVERSION_FAILED = C.tibems_status(38)

	tibems_INVALID_MSG      = C.tibems_status(42)
	tibems_INVALID_FIELD    = C.tibems_status(43)
	tibems_INVALID_INSTANCE = C.tibems_status(44)
	tibems_CORRUPT_MSG      = C.tibems_status(45)

	tibems_PRODUCER_FAILED = C.tibems_status(47)

	tibems_TIMEOUT                    = C.tibems_status(50)
	tibems_INTR                       = C.tibems_status(51)
	tibems_DESTINATION_LIMIT_EXCEEDED = C.tibems_status(52)
	tibems_MEM_LIMIT_EXCEEDED         = C.tibems_status(53)
	tibems_USER_INTR                  = C.tibems_status(54)

	tibems_INVALID_QUEUE_GROUP   = C.tibems_status(63)
	tibems_INVALID_TIME_INTERVAL = C.tibems_status(64)
	tibems_INVALID_IO_SOURCE     = C.tibems_status(65)
	tibems_INVALID_IO_CONDITION  = C.tibems_status(66)
	tibems_SOCKET_LIMIT          = C.tibems_status(67)

	tibems_OS_ERROR = C.tibems_status(68)

	tibems_WOULD_BLOCK = C.tibems_status(69)

	tibems_INSUFFICIENT_BUFFER = C.tibems_status(70)

	tibems_EOF            = C.tibems_status(71)
	tibems_INVALID_FILE   = C.tibems_status(72)
	tibems_FILE_NOT_FOUND = C.tibems_status(73)
	tibems_IO_FAILED      = C.tibems_status(74)

	tibems_WOULD_BLOCK_WANT_READ  = C.tibems_status(75)
	tibems_WOULD_BLOCK_WANT_WRITE = C.tibems_status(76)

	tibems_NOT_FILE_OWNER = C.tibems_status(80)

	tibems_ALREADY_EXISTS = C.tibems_status(91)

	tibems_INVALID_CONNECTION = C.tibems_status(100)
	tibems_INVALID_SESSION    = C.tibems_status(101)
	tibems_INVALID_CONSUMER   = C.tibems_status(102)
	tibems_INVALID_PRODUCER   = C.tibems_status(103)
	tibems_INVALID_USER       = C.tibems_status(104)
	tibems_INVALID_GROUP      = C.tibems_status(105)

	tibems_TRANSACTION_FAILED   = C.tibems_status(106)
	tibems_TRANSACTION_ROLLBACK = C.tibems_status(107)
	tibems_TRANSACTION_RETRY    = C.tibems_status(108)

	tibems_INVALID_XARESOURCE = C.tibems_status(109)

	tibems_FT_SERVER_LACKS_TRANSACTION = C.tibems_status(110)

	tibems_LDAP_ERROR         = C.tibems_status(120)
	tibems_INVALID_PROXY_USER = C.tibems_status(121)

	/* TLS related errors */
	tibems_INVALID_CERT         = C.tibems_status(150)
	tibems_INVALID_CERT_NOT_YET = C.tibems_status(151)
	tibems_INVALID_CERT_EXPIRED = C.tibems_status(152)
	tibems_INVALID_CERT_DATA    = C.tibems_status(153)
	tibems_ALGORITHM_ERROR      = C.tibems_status(154)
	tibems_SSL_ERROR            = C.tibems_status(155)
	tibems_INVALID_PRIVATE_KEY  = C.tibems_status(156)
	tibems_INVALID_ENCODING     = C.tibems_status(157)
	tibems_NOT_ENOUGH_RANDOM    = C.tibems_status(158)
	tibems_INVALID_CRL_DATA     = C.tibems_status(159)
	tibems_CRL_OFF              = C.tibems_status(160)
	tibems_EMPTY_CRL            = C.tibems_status(161)

	tibems_NOT_INITIALIZED    = C.tibems_status(200)
	tibems_INIT_FAILURE       = C.tibems_status(201)
	tibems_ARG_CONFLICT       = C.tibems_status(202)
	tibems_SERVICE_NOT_FOUND  = C.tibems_status(210)
	tibems_INVALID_CALLBACK   = C.tibems_status(211)
	tibems_INVALID_QUEUE      = C.tibems_status(212)
	tibems_INVALID_EVENT      = C.tibems_status(213)
	tibems_INVALID_SUBJECT    = C.tibems_status(214)
	tibems_INVALID_DISPATCHER = C.tibems_status(215)

	tibems_NO_MEMORY_FOR_OBJECT = C.tibems_status(237)

	tibems_UFO_CONNECTION_FAILURE = C.tibems_status(240)

	tibems_NOT_IMPLEMENTED = C.tibems_status(255)
)

// AcknowledgeMode specifies how and when message acknowledgments are sent by consumers
type AcknowledgeMode int32

const (
	AckModeSessionTransacted = AcknowledgeMode(0)
	AckModeAutoAcknowledge   = AcknowledgeMode(1)
	AckModeClientAcknowledge = AcknowledgeMode(2)
	AckModeDupsOkAcknowledge = AcknowledgeMode(3)

	AckModeNoAcknowledge                   = AcknowledgeMode(22) /* Extensions */
	AckModeExplicitClientAcknowledge       = AcknowledgeMode(23)
	AckModeExplicitClientDupsOkAcknowledge = AcknowledgeMode(24)
)

// DestinationType specifies a JMS destination type (topic or queue).
type DestinationType int32

const (
	DestTypeUnknown   = DestinationType(0)
	DestTypeQueue     = DestinationType(1)
	DestTypeTopic     = DestinationType(2)
	DestTypeUndefined = DestinationType(256)
)

// DeliveryMode specifies the persistence guarantees of a message
type DeliveryMode int32

const (
	DeliveryModeNonPersistent = DeliveryMode(1)
	DeliveryModePersistent    = DeliveryMode(2)
	DeliveryModeReliable      = DeliveryMode(22) /* Extension */
)

// NPSendCheckMode specifies the non-persistent send check mode for sends
type NPSendCheckMode int32

const (
	NpSendCheckDefault  = NPSendCheckMode(0)
	NpSendCheckAlways   = NPSendCheckMode(1)
	NpSendCheckNever    = NPSendCheckMode(2)
	NpSendCheckTempDest = NPSendCheckMode(3)
	NpSendCheckAuth     = NPSendCheckMode(4)
	NpSendCheckTempAuth = NPSendCheckMode(5)
)

// FieldType specifies the type (bool, short, int array, etc.) of a message field.
type FieldType int32

const (
	FieldTypeNull   = FieldType(0)
	FieldTypeBool   = FieldType(1)
	FieldTypeByte   = FieldType(2)
	FieldTypeWchar  = FieldType(3) /* double byte */
	FieldTypeShort  = FieldType(4)
	FieldTypeInt    = FieldType(5)
	FieldTypeLong   = FieldType(6)
	FieldTypeFloat  = FieldType(7)
	FieldTypeDouble = FieldType(8)
	/* Explicit size items */
	FieldTypeUTF8  = FieldType(9) /* UTF8-encoded String */
	FieldTypeBytes = FieldType(10)
	/* Extended MapMsg types */
	FieldTypeMapMsg      = FieldType(11)
	FieldTypeStreamMsg   = FieldType(12)
	FieldTypeShortArray  = FieldType(20)
	FieldTypeIntArray    = FieldType(21)
	FieldTypeLongArray   = FieldType(22)
	FieldTypeFloatArray  = FieldType(23)
	FieldTypeDoubleArray = FieldType(24)
)

// MsgType specifies the JMS type (StreamMessage, BytesMessage, MapMessage, etc.) of a message body.
type MsgType int32

const (
	MsgTypeUnknown       = MsgType(0)
	MsgTypeMessage       = MsgType(1)
	MsgTypeBytesMessage  = MsgType(2)
	MsgTypeMapMessage    = MsgType(3)
	MsgTypeObjectMessage = MsgType(4)
	MsgTypeStreamMessage = MsgType(5)
	MsgTypeTextMessage   = MsgType(6)
	MsgTypeUndefined     = MsgType(256)
)

type TMFlag uint32

const (
	TMNOFLAGS   = TMFlag(0x00000000)
	TMREGISTER  = TMFlag(0x00000001)
	TMNOMIGRATE = TMFlag(0x00000002)
	TMUSEASYNC  = TMFlag(0x00000004)

	TMASYNC      = TMFlag(0x80000000)
	TMONEPHASE   = TMFlag(0x40000000)
	TMFAIL       = TMFlag(0x20000000)
	TMNOWAIT     = TMFlag(0x10000000)
	TMRESUME     = TMFlag(0x08000000)
	TMSUCCESS    = TMFlag(0x04000000)
	TMSUSPEND    = TMFlag(0x02000000)
	TMSTARTRSCAN = TMFlag(0x01000000)
	TMENDRSCAN   = TMFlag(0x00800000)
	TMMULTIPLE   = TMFlag(0x00400000)
	TMJOIN       = TMFlag(0x00200000)
	TMMIGRATE    = TMFlag(0x00100000)
)

type AXReturn int32

const (
	TM_JOIN    = AXReturn(2)
	TM_RESUME  = AXReturn(1)
	TM_OK      = AXReturn(0)
	TMER_TMERR = AXReturn(-1)
	TMER_INVAL = AXReturn(-2)
	TMER_PROTO = AXReturn(-3)
)

type XAError int32

const (
	XA_OK          = XAError(0)
	XA_HEURCOM     = XAError(7)
	XA_HEURAZ      = XAError(8)
	XA_HEURMIX     = XAError(5)
	XA_HEURRB      = XAError(6)
	XA_NOMIGRATE   = XAError(9)
	XA_RBBASE      = XAError(100)
	XA_RBCOMMFAIL  = XAError(101)
	XA_RBDEADLOCK  = XAError(102)
	XA_RBEND       = XAError(107)
	XA_RBINTEGRITY = XAError(103)
	XA_RBOTHER     = XAError(104)
	XA_RBPROTO     = XAError(105)
	XA_RBROLLBACK  = XAError(100)
	XA_RBTIMEOUT   = XAError(106)
	XA_RBTRANSIENT = XAError(107)
	XA_RDONLY      = XAError(3)
	XA_RETRY       = XAError(4)
	XAER_ASYNC     = XAError(-2)
	XAER_DUPID     = XAError(-8)
	XAER_INVAL     = XAError(-5)
	XAER_NOTA      = XAError(-4)
	XAER_OUTSIDE   = XAError(-9)
	XAER_PROTO     = XAError(-6)
	XAER_RMERR     = XAError(-3)
	XAER_RMFAIL    = XAError(-7)
)

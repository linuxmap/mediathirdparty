package sdp

import "time"

// MediaType is the media type for an SDP session description.
const MediaType = "application/sdp"

// Attribute values for indication of a media stream direction.
// See RFC 4566 Section 6.
const (
	ModeSendRecv = "sendrecv"
	ModeRecvOnly = "recvonly"
	ModeSendOnly = "sendonly"
	ModeInactive = "inactive"
)

// Connection mode values for an SDP "setup" attribute.
// See RFC 4145 Section 4.
const (
	SetupActive  = "active"
	SetupPassive = "passive"
	SetupActPass = "actpass"
	SetupHold    = "holdconn"
)

// Description represents an SDP session description. RFC 4566 Section 5.
type Description struct {
	Version     int            // Protocol Version ("v=")
	Origin      *Origin        // Origin ("o=")
	Session     string         // Session Name ("s=")
	Information string         // Session Information ("i=")
	URI         string         // URI ("u=")
	Email       []string       // Email Address ("e=")
	Phone       []string       // Phone Number ("p=")
	Connection  *Connection    // Connection Data ("c=")
	Bandwidth   map[string]int // Bandwidth ("b=")
	Timing      *Timing        // Timing ("t=")
	TimeZones   []*TimeZone    // TimeZone ("t=")
	Key         *Key           // Encryption Keys ("k=")
	Attributes  []*Attribute   // Attribute ("a=")
	Media       []*Media       // Media Descriptions ("m=")
	Mode        string
	SSRC        string // SSRC ("y=")
}

// String returns the encoded session description according the SDP specification.
func (desc *Description) String() string {
	enc := NewEncoder()
	enc.Encode(desc)
	return enc.String()
}

// Parse parses text into a Description structure.
func ParseSdp(text string) (*Description, error) {
	dec := NewDecoderString(text)
	return dec.Decode()
}

// Origin represents an originator of the session. RFC 4566 Section 5.2.
type Origin struct {
	Username       string
	SessionId      int64
	SessionVersion int64
	Network        string
	Type           string
	Address        string
}

// Media contains media description. RFC 4566 Section 5.14.
type Media struct {
	Type        string
	Port        int
	PortNum     int
	Proto       string
	Formats     map[int]*Format
	Information string         // Media Information ("i=")
	Connection  *Connection    // Connection Data ("c=")
	Bandwidth   map[string]int // Bandwidth ("b=")
	Key         *Key           // Encryption Keys ("k=")
	Attributes  []*Attribute   // Attribute ("a=")
	Mode        string
	SSRC        string   // SSRC ("y=")
	Control     *Control // RTCP description
}

// Format is a media format description represented by "rtpmap", "fmtp" SDP attributes.
type Format struct {
	Payload  int
	Codec    string
	Clock    int
	Channels int
	Feedback []string
	Params   []string
}

// Control contains description of an RTCP endpoint.
type Control struct {
	Muxed   bool
	Network string
	Type    string
	Address string
	Port    int
}

// Key contains a key exchange information.
// It's use is not recommended, supported for compatibility with older implementations.
type Key struct {
	Type, Value string
}

// Attribute represents an a session or media attribute. RFC 4566 Section 5.14.
type Attribute struct {
	Name, Value string
}

func (a *Attribute) String() string {
	if a.Value == "" {
		return a.Name
	}
	return a.Name + ":" + a.Value
}

// Connection contains connection data. RFC 4566 Section 5.7.
type Connection struct {
	Network string
	Type    string
	Address string
}

// Timing specifies start and stop times for a session.
type Timing struct {
	Start  time.Time
	Stop   time.Time
	Repeat *Repeat // Repeat Times ("r=")
}

// Repeat specifies repeat times for a session.
type Repeat struct {
	Interval time.Duration
	Duration time.Duration
	Offsets  []time.Duration
}

// TimeZone represents a time zones change information for a repeated session.
type TimeZone struct {
	Time   time.Time
	Offset time.Duration
}

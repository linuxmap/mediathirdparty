package sdp

import (
	"sort"
	"strconv"
	"time"
)

// An Encoder writes an SDP description to a buffer.
type Encoder struct {
	buf  []byte
	pos  int
	cont bool
}

// NewEncoder returns a new encoder.
func NewEncoder() *Encoder {
	return &Encoder{}
}

func (enc *Encoder) next(n int) (b []byte) {
	p := enc.pos + n
	if len(enc.buf) < p {
		enc.grow(n)
	}
	b, enc.pos = enc.buf[enc.pos:p], p
	return
}

func (enc *Encoder) grow(n int) {
	p := enc.pos + n
	b := make([]byte, (1+((p-1)>>10))<<10)
	if enc.pos > 0 {
		copy(b, enc.buf[:enc.pos])
	}
	enc.buf = b
}

func (enc *Encoder) line(typ byte) {
	if enc.cont {
		b := enc.next(4)
		b[0] = '\r'
		b[1] = '\n'
		b[2] = typ
		b[3] = '='
	} else {
		b := enc.next(2)
		b[0] = typ
		b[1] = '='
		enc.cont = true
	}
}

func (enc *Encoder) char(ch byte) {
	b := enc.next(1)
	b[0] = ch
}

func (enc *Encoder) int(v int64) {
	b := enc.next(20)
	enc.pos += len(strconv.AppendInt(b[:0], v, 10)) - len(b)
}

func (enc *Encoder) string(v string) {
	copy(enc.next(len(v)), v)
}

func (enc *Encoder) fields(v ...string) {
	n := len(v) - 1
	for _, it := range v {
		n += len(it)
	}
	if n < 0 {
		return
	}
	b := enc.next(n)
	i := 0
	for _, it := range v {
		if i > 0 {
			b[i] = ' '
			i++
		}
		copy(b[i:], it)
		i += len(it)
	}
}

// Bytes returns a slice of buffer holding the encoded SDP description.
func (enc *Encoder) Bytes() []byte {
	if enc.cont {
		b := enc.next(2)
		b[0] = '\r'
		b[1] = '\n'
		enc.cont = false
	}
	return enc.buf[:enc.pos]
}

// String returns the encoded SDP description as string.
func (enc *Encoder) String() string {
	return string(enc.Bytes())
}

// Encode writes the SDP description into the buffer.
func (enc *Encoder) Encode(desc *Description) {
	enc.pos = 0
	enc.cont = false

	enc.line('v')
	enc.int(int64(desc.Version))

	if desc.Origin != nil {
		enc.encodeOrigin(desc.Origin)
	}
	enc.line('s')
	if desc.Session == "" {
		enc.char('-')
	} else {
		enc.string(desc.Session)
	}
	if desc.Information != "" {
		enc.line('i')
		enc.string(desc.Information)
	}
	if desc.URI != "" {
		enc.line('u')
		enc.string(desc.URI)
	}
	enc.encodeList('e', desc.Email)
	enc.encodeList('p', desc.Phone)
	if c := desc.Connection; c != nil {
		enc.line('c')
		enc.encodeConn(c.Network, c.Type, c.Address)
	}
	for typ, v := range desc.Bandwidth {
		enc.encodeBandwidth(typ, v)
	}
	enc.encodeTiming(desc.Timing)
	enc.encodeTimezones(desc.TimeZones)

	if k := desc.Key; k != nil {
		enc.encodeKey(k.Type, k.Value)
	}
	if desc.Mode != "" {
		enc.encodeAttr(desc.Mode, "")
	}
	for _, it := range desc.Attributes {
		enc.encodeAttr(it.Name, it.Value)
	}
	for _, it := range desc.Media {
		enc.encodeMediaDesc(it)
	}
	if desc.SSRC != "" {
		enc.line('y')
		enc.string(desc.SSRC)
	}
}

func (enc *Encoder) encodeMediaDesc(m *Media) {
	fmts := make([]int, 0, len(m.Formats))
	for p := range m.Formats {
		fmts = append(fmts, p)
	}
	sort.Ints(fmts)

	enc.line('m')
	enc.string(m.Type)
	enc.char(' ')
	enc.int(int64(m.Port))
	if m.PortNum != 0 {
		enc.char('/')
		enc.int(int64(m.PortNum))
	}
	enc.char(' ')
	enc.string(m.Proto)
	for _, p := range fmts {
		enc.char(' ')
		enc.int(int64(p))
	}
	if c := m.Control; c != nil {
		if c.Muxed {
			enc.line('a')
			enc.string("rtcp-mux")
		} else {
			enc.line('a')
			enc.string("rtcp:")
			enc.int(int64(c.Port))
			enc.encodeConn(c.Network, c.Type, c.Address)
		}
	}
	if m.Information != "" {
		enc.line('i')
		enc.string(m.Information)
	}
	if m.SSRC != "" {
		enc.line('y')
		enc.string(m.SSRC)
	}
	if c := m.Connection; c != nil {
		enc.line('c')
		enc.encodeConn(c.Network, c.Type, c.Address)
	}
	for typ, v := range m.Bandwidth {
		enc.encodeBandwidth(typ, v)
	}
	if k := m.Key; k != nil {
		enc.encodeKey(k.Type, k.Value)
	}
	for _, p := range fmts {
		enc.encodeMediaMap(m.Formats[p])
	}
	if m.Mode != "" {
		enc.encodeAttr(m.Mode, "")
	}
	for _, it := range m.Attributes {
		enc.encodeAttr(it.Name, it.Value)
	}
}

func (enc *Encoder) encodeMediaMap(f *Format) {
	if f == nil {
		return
	}
	enc.line('a')
	enc.string("rtpmap:")
	enc.int(int64(f.Payload))
	enc.char(' ')
	enc.string(f.Codec)
	enc.char('/')
	enc.int(int64(f.Clock))
	if f.Channels != 0 {
		enc.char('/')
		enc.int(int64(f.Channels))
	}
	for _, it := range f.Params {
		enc.line('a')
		enc.string("fmtp:")
		enc.int(int64(f.Payload))
		enc.char(' ')
		enc.string(it)
	}
}

func (enc *Encoder) encodeTiming(t *Timing) {
	enc.line('t')
	if t == nil {
		enc.string("0 0")
	} else {
		enc.encodeTime(t.Start)
		enc.char(' ')
		enc.encodeTime(t.Stop)
		if t.Repeat != nil {
			enc.encodeRepeat(t.Repeat)
		}
	}
}

func (enc *Encoder) encodeRepeat(r *Repeat) {
	enc.line('r')
	enc.encodeDuration(r.Interval)
	enc.char(' ')
	enc.encodeDuration(r.Duration)
	for _, it := range r.Offsets {
		enc.char(' ')
		enc.encodeDuration(it)
	}
}

func (enc *Encoder) encodeTimezones(z []*TimeZone) {
	if len(z) > 0 {
		enc.line('z')
		for i, it := range z {
			if i > 0 {
				enc.char(' ')
			}
			enc.encodeTime(it.Time)
			enc.char(' ')
			enc.encodeDuration(it.Offset)
		}
	}
}

func (enc *Encoder) encodeAttr(k, v string) {
	enc.line('a')
	enc.string(k)
	if v != "" {
		enc.char(':')
		enc.string(v)
	}
}

func (enc *Encoder) encodeKey(k, v string) {
	enc.line('k')
	enc.string(k)
	if v != "" {
		enc.char(':')
		enc.string(v)
	}
}

func (enc *Encoder) encodeList(typ byte, v []string) {
	for _, it := range v {
		enc.line(typ)
		enc.string(it)
	}
}

func (enc *Encoder) encodeTime(t time.Time) {
	if t.IsZero() {
		enc.char('0')
	} else {
		d := int64(t.Sub(ntpEpoch).Seconds())
		enc.int(d)
	}
}

func (enc *Encoder) encodeDuration(d time.Duration) {
	sec := int64(d.Seconds())
	if sec == 0 {
		enc.char('0')
	} else if sec%86400 == 0 {
		enc.int(sec / 86400)
		enc.char('d')
	} else if sec%3600 == 0 {
		enc.int(sec / 3600)
		enc.char('h')
	} else if sec%60 == 0 {
		enc.int(sec / 60)
		enc.char('m')
	} else {
		enc.int(sec)
	}
}

func (enc *Encoder) encodeOrigin(orig *Origin) {
	enc.line('o')
	if orig.Username == "" {
		enc.char('-')
	} else {
		enc.string(orig.Username)
	}
	enc.char(' ')
	enc.int(orig.SessionId)
	enc.char(' ')
	enc.int(orig.SessionVersion)
	enc.char(' ')
	enc.encodeConn(orig.Network, orig.Type, orig.Address)
}

func (enc *Encoder) encodeConn(network, typ, addr string) {
	if network == "" {
		network = "IN"
	}
	if typ == "" {
		typ = "IP4"
	}
	if addr == "" {
		addr = "127.0.0.1"
	}
	enc.fields(network, typ, addr)
}

func (enc *Encoder) encodeBandwidth(typ string, v int) {
	enc.line('b')
	enc.string(typ)
	enc.char(':')
	enc.int(int64(v))
}

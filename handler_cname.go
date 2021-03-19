package fastdns

import (
	"net"
)

func CNAMERecord(rw ResponseWriter, req *Request, cnames []string, ips []net.IP, ttl uint32) {
	b := AcquireByteBuffer()
	defer ReleaseByteBuffer(b)

	b.B = AppendRequestToResponse(b.B[:0], req, NOERROR, 1, uint16(len(cnames)+len(ips)), 0, 0)

	offset := 12
	for _, cname := range cnames {
		answer := [...]byte{
			// NAME
			0xc0, byte(offset),
			// TYPE
			0x00, byte(QTypeCNAME),
			// CLASS
			byte(req.Question.QClass >> 8), byte(req.Question.QClass),
			// TTL
			byte(ttl >> 24), byte(ttl >> 16), byte(ttl >> 8), byte(ttl),
			// RDLENGTH
			0x00, byte(1 + len(cname) + 1),
		}
		b.B = append(b.B, answer[:]...)
		// set offset
		offset += len(b.B)
		// RDATA
		b.B = encodeDomain(b.B, cname)
	}

	for _, ip := range ips {
		if ip4 := ip.To4(); ip4 != nil {
			_ = ip4[3]
			answer := [...]byte{
				// NAME
				0xc0, byte(offset),
				// TYPE
				0x00, byte(QTypeA),
				// CLASS
				byte(req.Question.QClass >> 8), byte(req.Question.QClass),
				// TTL
				byte(ttl >> 24), byte(ttl >> 16), byte(ttl >> 8), byte(ttl),
				// RDLENGTH
				0x00, 0x04,
				// RDATA
				ip4[0], ip4[1], ip4[2], ip4[3],
			}
			b.B = append(b.B, answer[:]...)
		} else {
			_ = ip[15]
			answer := [...]byte{
				// NAME
				0xc0, byte(offset),
				// TYPE
				0x00, byte(QTypeAAAA),
				// CLASS
				byte(req.Question.QClass >> 8), byte(req.Question.QClass),
				// TTL
				byte(ttl >> 24), byte(ttl >> 16), byte(ttl >> 8), byte(ttl),
				// RDLENGTH
				0x00, 0x10,
				// RDATA
				ip[0], ip[1], ip[2], ip[3],
				ip[4], ip[5], ip[6], ip[7],
				ip[8], ip[9], ip[10], ip[11],
				ip[12], ip[13], ip[14], ip[15],
			}
			b.B = append(b.B, answer[:]...)
		}
	}

	rw.Write(b.B)
}
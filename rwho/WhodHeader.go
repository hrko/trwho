package rwho

import (
	"time"
)

// whod represents rwho protocol packet.
type WhodHeader struct {
	Version     byte     // protocol version #
	Type        byte     // packet type
	Padding     [2]byte  // padding
	Sendtime    int32    // time stamp by sender
	Recvtime    int32    // time stamp applied by receiver
	Hostname    [32]byte // host's name
	LoadAverage [3]int32 // load average as in uptime
	Boottime    int32    // time system booted
}

func (h *WhodHeader) GetSendTime() time.Time {
	return time.Unix(int64(h.Sendtime), 0)
}
func (h *WhodHeader) GetRecvTime() time.Time {
	return time.Unix(int64(h.Recvtime), 0)
}
func (h *WhodHeader) GetBootTime() time.Time {
	return time.Unix(int64(h.Boottime), 0)
}
func (h *WhodHeader) GetUptime() time.Duration {
	return time.Since(h.GetBootTime())
}
func (h *WhodHeader) GetHostname() string {
	return bytesToString(h.Hostname[:])
}
func (h *WhodHeader) GetLoadAverage1min() float64 {
	return float64(h.LoadAverage[0]) / 100.0
}
func (h *WhodHeader) GetLoadAverage5min() float64 {
	return float64(h.LoadAverage[1]) / 100.0
}
func (h *WhodHeader) GetLoadAverage15min() float64 {
	return float64(h.LoadAverage[2]) / 100.0
}
func (h *WhodHeader) IsDown() bool {
	return time.Since(h.GetSendTime()) > time.Minute*10
}

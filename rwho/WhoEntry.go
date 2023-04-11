package rwho

import (
	"time"
)

// WhoEntry represents tty information and idle time.
type WhoEntry struct {
	Tty       [8]byte // tty name
	User      [8]byte // user id
	LoginTime int32   // time on
	IdleTime  int32   // tty idle time
}

func (e *WhoEntry) GetUser() string {
	return bytesToString(e.User[:])
}
func (e *WhoEntry) GetTty() string {
	return bytesToString(e.Tty[:])
}
func (e *WhoEntry) GetLoginTime() time.Time {
	return time.Unix(int64(e.LoginTime), 0)
}
func (e *WhoEntry) GetIdleTime() time.Duration {
	return time.Duration(int64(e.IdleTime) * 1000000000)
}

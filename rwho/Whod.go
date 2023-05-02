package rwho

type Whod struct {
	Header     *WhodHeader
	WhoEntries []*WhoEntry
}

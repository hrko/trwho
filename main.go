package main

import (
	"fmt"
	"log"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/hrko/it-rmap-go/rwho"

	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type host struct {
	Whod *rwho.Whod
}

type tableData struct {
	tview.TableContentReadOnly
	hosts []*host
}

func (d *tableData) GetCell(row, column int) *tview.TableCell {
	head := []string{"Hostname", "DNS Record", "Uptime", "Load", "Users(idle time)"}
	if row > len(d.hosts) {
		return nil
	}
	if column > len(head)-1 {
		return nil
	}

	c := tview.NewTableCell("")

	// set text
	txt := ""
	if row == 0 { // head
		txt = head[column]
	} else { // data
		h := d.hosts[row-1]
		whodHeader := h.Whod.Header
		switch column {
		case 0:
			txt = whodHeader.GetHostname()
		case 1:
			if ss, err := net.LookupHost(whodHeader.GetHostname()); err == nil {
				txt = ss[0]
			}
		case 2:
			if whodHeader.IsDown() {
				txt = "[red]down"
			} else {
				txt = fmtDuration(whodHeader.GetUptime())
			}
		case 3:
			txt = fmt.Sprintf("%1.2f", whodHeader.GetLoadAverage1min())
		case 4:
			for _, e := range h.Whod.WhoEntries {
				txt += e.GetUser()
				txt += "@" + e.GetTty()
				txt += "(" + fmtDuration(e.GetIdleTime()) + ")" + " "
			}
			txt = strings.TrimSpace(txt)
		}
	}
	c.SetText(surround(txt, " "))

	// set properties
	if row == 0 {
		c.SetSelectable(false)
		c.SetTextColor(tcell.ColorBlack)
		c.SetBackgroundColor(tcell.ColorWhite)
		c.SetAttributes(tcell.AttrBold)
	} else {
		if d.hosts[row-1].Whod.Header.IsDown() {
			c.SetAttributes(tcell.AttrDim)
		}
	}
	switch column {
	case 0:
		c.SetAlign(tview.AlignLeft)
	case 1:
		c.SetAlign(tview.AlignLeft)
	case 2:
		c.SetAlign(tview.AlignRight)
	case 3:
		c.SetAlign(tview.AlignRight)
	case 4:
		c.SetAlign(tview.AlignLeft)
		c.SetExpansion(1)
	}

	return c
}

func (d *tableData) GetRowCount() int {
	return len(d.hosts) + 1
}

func (d *tableData) GetColumnCount() int {
	return 5
}

func getHosts() []*host {
	whodFiles, err := filepath.Glob("/var/spool/rwho/whod.*")
	if err != nil {
		log.Fatalln(err)
	}
	hostCnt := len(whodFiles)
	hosts := make([]*host, hostCnt)
	for i, f := range whodFiles {
		h := new(host)
		hosts[i] = h
		h.Whod.Header, h.Whod.WhoEntries, err = rwho.ReadWhod(f)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return hosts
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	m := (d % time.Hour) / time.Minute
	return fmt.Sprintf("%d:%02d", h, m)
}

func surround(s string, c string) string {
	return c + s + c
}

func main() {
	// application config
	app := tview.NewApplication()

	// table config
	table := tview.NewTable()
	table.SetSelectable(true, false)
	table.SetBackgroundColor(tcell.ColorReset)
	var selectedStyle tcell.Style
	selectedStyle = selectedStyle.Background(tcell.ColorBlue)
	selectedStyle = selectedStyle.Foreground(tcell.ColorBlack)
	selectedStyle = selectedStyle.Attributes(tcell.AttrInvalid)
	table.SetSelectedStyle(selectedStyle)
	data := &tableData{hosts: getHosts()}
	table.SetContent(data)
	table.SetFixed(1, 0)
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			app.Stop()
		}
		return event
	})

	// watch changes in /var/spool/rwho
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}
	defer watcher.Close()

	if err := watcher.Add("/var/spool/rwho"); err != nil {
		log.Fatalln(err)
	}

	// update when a new packet arrives
	go func() {
		for {
			select {
			case <-watcher.Events:
				data.hosts = getHosts()
				app.QueueUpdateDraw(func() {})
			case err := <-watcher.Errors:
				log.Fatalln(err)
			}
		}
	}()

	// update once a minute
	go func() {
		for {
			time.Sleep(time.Minute)
			app.QueueUpdateDraw(func() {})
		}
	}()

	frame := tview.NewFrame(table).
		SetBorders(0, 0, 0, 0, 0, 0).
		AddText("         ^F,PGDN:page down  ^B,PGUP:page up  g,HOME:top  G,END:bottom", false, tview.AlignLeft, tcell.ColorWhite).
		AddText(" q:quit  j,↓:down           k,↑:up           h,←:left    l,→:right", false, tview.AlignLeft, tcell.ColorWhite)

	if err := app.SetRoot(frame, true).SetFocus(table).Run(); err != nil {
		log.Fatalln(err)
	}
}

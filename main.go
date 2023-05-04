package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/hrko/trwho/rwho"

	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/exp/slices"
)

const appName = "trwho"

type HostProperty int

const (
	Hostname HostProperty = iota
	Note
	ResolvedAddr
	Uptime
	Load
	Users
	HostPropertyCount
)

func (i HostProperty) String() string {
	switch i {
	case Hostname:
		return "Hostname"
	case Note:
		return "Note"
	case ResolvedAddr:
		return "Resolved Addr"
	case Uptime:
		return "Uptime"
	case Load:
		return "Load"
	case Users:
		return "Users"
	default:
		return ""
	}
}

type Host struct {
	Hostname  string
	Whod      *rwho.Whod
	Config    *ConfigHostEntry
	IpAddress string
}

func NewHost(hostname string) *Host {
	h := new(Host)
	h.Hostname = hostname
	if ss, err := net.LookupHost(hostname); err == nil {
		h.IpAddress = ss[0]
	} else {
		h.IpAddress = ""
	}
	return h
}

func (h *Host) IsDown() bool {
	if h.Whod == nil {
		return true
	}
	return h.Whod.Header.IsDown()
}

func (h *Host) Value(i HostProperty) string {
	switch i {
	case Hostname:
		return h.Hostname
	case Note:
		if h.Config != nil {
			return h.Config.Note
		} else {
			return ""
		}
	case ResolvedAddr:
		return h.IpAddress
	case Uptime:
		if h.IsDown() {
			return "[red]down"
		} else {
			return fmtDuration(h.Whod.Header.GetUptime())
		}
	case Load:
		if h.IsDown() {
			return ""
		} else {
			return fmt.Sprintf("%1.2f", h.Whod.Header.GetLoadAverage1min())
		}
	case Users:
		if h.IsDown() {
			return ""
		} else {
			text := ""
			for _, e := range h.Whod.WhoEntries {
				text += e.GetUser()
				text += "@" + e.GetTty()
				text += "(" + fmtDuration(e.GetIdleTime()) + ")" + " "
			}
			text = strings.TrimSpace(text)
			return text
		}
	default:
		return ""
	}
}

func InitHostList(config *Config) ([]*Host, error) {
	hosts := make([]*Host, 0)

	if config != nil {
		for _, configHostEntry := range config.Hosts {
			h := NewHost(configHostEntry.Hostname)
			h.Config = configHostEntry
			hosts = append(hosts, h)
		}
	}

	whods, err := rwho.ScanHosts()
	if err != nil {
		return nil, err
	}
	for _, whod := range whods {
		idx := slices.IndexFunc(hosts, func(h *Host) bool {
			return h.Hostname == whod.Header.GetHostname()
		})
		if idx != -1 {
			hosts[idx].Whod = whod
		} else {
			h := NewHost(whod.Header.GetHostname())
			h.Whod = whod
			hosts = append(hosts, h)
		}
	}

	return hosts, nil
}

type TableData struct {
	tview.TableContentReadOnly
	Hosts []*Host
}

func (d *TableData) GetCell(row, column int) *tview.TableCell {
	c := tview.NewTableCell("")

	if row >= d.GetRowCount() {
		return c
	}
	if column >= d.GetColumnCount() {
		return c
	}

	// set text
	cellText := ""
	if row == 0 { // head
		cellText = HostProperty(column).String()
	} else { // data
		host := d.Hosts[row-1]
		cellText = host.Value(HostProperty(column))
	}
	c.SetText(surround(cellText, " "))

	// set properties
	if row == 0 {
		c.SetSelectable(false)
		c.SetTextColor(tcell.ColorBlack)
		c.SetBackgroundColor(tcell.ColorWhite)
		c.SetAttributes(tcell.AttrBold)
	} else {
		if d.Hosts[row-1].Whod != nil {
			if d.Hosts[row-1].Whod.Header.IsDown() {
				c.SetAttributes(tcell.AttrDim)
			}
		} else {
			c.SetAttributes(tcell.AttrDim)
		}
	}
	switch HostProperty(column) {
	case Hostname:
		c.SetAlign(tview.AlignLeft)
	case Note:
		c.SetAlign(tview.AlignLeft)
	case ResolvedAddr:
		c.SetAlign(tview.AlignLeft)
	case Uptime:
		c.SetAlign(tview.AlignRight)
	case Load:
		c.SetAlign(tview.AlignRight)
	case Users:
		c.SetAlign(tview.AlignLeft)
		c.SetExpansion(1)
	}

	return c
}

func (d *TableData) GetRowCount() int {
	return len(d.Hosts) + 1
}

func (d *TableData) GetColumnCount() int {
	return int(HostPropertyCount)
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
	// load config
	configWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln("Cannot create watcher for config file.")
	}
	defer configWatcher.Close()
	var config *Config
	configPath, err := SearchConfigFile()
	if err == nil {
		config, err = ReadConfig(configPath)
		if err != nil {
			log.Fatalln("Cannot read config file.")
		}
		err = configWatcher.Add(configPath)
		if err != nil {
			log.Fatalln("Cannot watch config file.")
		}
	}

	// initialize host list
	hosts, err := InitHostList(config)
	if err != nil {
		log.Fatalln("Failed to initialize host list at startup.")
	}

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
	data := &TableData{Hosts: hosts}
	table.SetContent(data)
	table.SetFixed(1, 1)
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			app.Stop()
		}
		return event
	})

	// watch changes in /var/spool/rwho
	rwhoWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}
	defer rwhoWatcher.Close()
	if err := rwhoWatcher.Add("/var/spool/rwho"); err != nil {
		log.Fatalln(err)
	}

	// update when a new packet arrives
	go func() {
		for {
			select {
			case event, ok := <-rwhoWatcher.Events:
				if !ok {
					return
				}
				whodPath := event.Name
				whod, err := rwho.ReadWhodFile(whodPath)
				if err != nil {
					log.Println(err)
					log.Printf("Cannot read whod file \"%s\".\n", whodPath)
				}
				idx := slices.IndexFunc(data.Hosts, func(h *Host) bool {
					return h.Hostname == whod.Header.GetHostname()
				})
				if idx != -1 {
					data.Hosts[idx].Whod = whod
				} else {
					h := NewHost(whod.Header.GetHostname())
					h.Whod = whod
					data.Hosts = append(data.Hosts, h)
				}
				app.QueueUpdateDraw(func() {})
			case err, ok := <-rwhoWatcher.Errors:
				if !ok {
					return
				}
				log.Fatalln(err)
			}
		}
	}()

	// update on config change
	go func() {
		for {
			select {
			case event, ok := <-configWatcher.Events:
				if !ok {
					return
				}
				config, err = ReadConfig(event.Name)
				if err != nil {
					log.Println("Cannot reload config.")
					continue
				}
				hosts, err := InitHostList(config)
				if err != nil {
					log.Println("Cannot re-init host list.")
					continue
				}
				data.Hosts = hosts
				app.QueueUpdateDraw(func() {})
			case err, ok := <-configWatcher.Errors:
				if !ok {
					return
				}
				log.Fatalln(err)
			}
		}
	}()

	// update once a second
	go func() {
		for {
			time.Sleep(time.Second)
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

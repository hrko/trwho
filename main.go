package main

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/hrko/it-rmap-go/rwho"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
	whodFiles, err := filepath.Glob("/var/spool/rwho/whod.*")
	if err != nil {
		panic(err)
	}
	hostCnt := len(whodFiles)
	headers := make([]*rwho.WhodHeader, hostCnt)
	entries := make([][]*rwho.WhoEntry, hostCnt)
	for i, p := range whodFiles {
		headers[i], entries[i], err = rwho.ReadWhod(p)
		if err != nil {
			panic(err)
		}
	}

	app := tview.NewApplication()

	table := tview.NewTable()
	table.SetSelectable(true, false)
	table.SetBackgroundColor(tcell.ColorReset)
	var selectedStyle tcell.Style
	selectedStyle = selectedStyle.Background(tcell.ColorBlue)
	selectedStyle = selectedStyle.Foreground(tcell.ColorBlack)
	selectedStyle = selectedStyle.Attributes(tcell.AttrInvalid)
	table.SetSelectedStyle(selectedStyle)
	// table.SetSeparator('|')

	type column struct {
		head  string
		align int
		cells []*tview.TableCell
	}

	cols := make([]*column, 0)
	cols = append(cols, &column{
		head:  "Hostname",
		align: tview.AlignLeft,
		cells: make([]*tview.TableCell, 0),
	})
	cols = append(cols, &column{
		head:  "IP Address",
		align: tview.AlignLeft,
		cells: make([]*tview.TableCell, 0),
	})
	cols = append(cols, &column{
		head:  "Uptime",
		align: tview.AlignRight,
		cells: make([]*tview.TableCell, 0),
	})
	cols = append(cols, &column{
		head:  "Load",
		align: tview.AlignRight,
		cells: make([]*tview.TableCell, 0),
	})
	cols = append(cols, &column{
		head:  "Users(idle time)",
		align: tview.AlignLeft,
		cells: make([]*tview.TableCell, 0),
	})
	for _, c := range cols {
		cell := tview.NewTableCell(surround(c.head, " "))
		cell.SetAttributes(tcell.AttrBold)
		cell.SetSelectable(false)
		cell.SetBackgroundColor(tcell.ColorWhite)
		cell.SetTextColor(tcell.ColorBlack)
		c.cells = append(c.cells, cell)
	}
	for i := 0; i < hostCnt; i++ {
		h := headers[i]

		ipAddrStr := ""
		if ss, err := net.LookupHost(h.GetHostname()); err == nil {
			ipAddrStr = ss[0]
		}
		loadavStr := fmt.Sprintf("%1.2f", h.GetLoadAverage1min())
		usersStr := ""
		for _, e := range entries[i] {
			usersStr += e.GetUser()
			usersStr += "@" + e.GetTty()
			usersStr += "(" + fmtDuration(e.GetIdleTime()) + ")" + " "
		}
		usersStr = strings.TrimSpace(usersStr)

		cellHostname := tview.NewTableCell(surround(h.GetHostname(), " "))
		cellIpAddr := tview.NewTableCell(surround(ipAddrStr, " "))
		cellUptime := tview.NewTableCell(surround(fmtDuration(h.GetUptime()), " "))
		cellLoadav := tview.NewTableCell(surround(loadavStr, " "))
		cellUsers := tview.NewTableCell(surround(usersStr, " "))

		cols[0].cells = append(cols[0].cells, cellHostname)
		cols[1].cells = append(cols[1].cells, cellIpAddr)
		cols[2].cells = append(cols[2].cells, cellUptime)
		cols[3].cells = append(cols[3].cells, cellLoadav)
		cols[4].cells = append(cols[4].cells, cellUsers)

		if h.IsDown() {
			cellHostname.SetAttributes(tcell.AttrDim)
			cellIpAddr.SetAttributes(tcell.AttrDim)
			cellUptime.SetAttributes(tcell.AttrDim)
			cellLoadav.SetAttributes(tcell.AttrDim)
			cellUsers.SetAttributes(tcell.AttrDim)
		}
	}
	for colIdx, col := range cols {
		for rowIdx, cell := range col.cells {
			cell.SetAlign(col.align)
			table.SetCell(rowIdx, colIdx, cell)
		}
	}

	if err := app.SetRoot(table, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}
}

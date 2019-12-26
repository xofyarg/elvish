// Command widget allows manually testing a single widget.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/cli/el/codearea"
	"github.com/elves/elvish/pkg/cli/el/combobox"
	"github.com/elves/elvish/pkg/cli/el/listbox"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/ui"
)

var (
	maxHeight  = flag.Int("max-height", 10, "maximum height")
	horizontal = flag.Bool("horizontal", false, "use horizontal listbox layout")
)

func makeWidget() el.Widget {
	items := listbox.TestItems{Prefix: "list item "}
	w := combobox.NewComboBox(combobox.ComboBoxSpec{
		CodeArea: codearea.CodeAreaSpec{
			Prompt: func() ui.Text {
				return ui.T(" NUMBER ", ui.Bold, ui.BgMagenta).ConcatText(ui.T(" "))
			},
		},
		ListBox: listbox.ListBoxSpec{
			State:       listbox.ListBoxState{Items: &items},
			Placeholder: ui.T("(no items)"),
			Horizontal:  *horizontal,
		},
		OnFilter: func(w combobox.ComboBox, filter string) {
			if n, err := strconv.Atoi(filter); err == nil {
				items.NItems = n
			}
		},
	})
	return w
}

func main() {
	flag.Parse()
	widget := makeWidget()

	tty := cli.NewTTY(os.Stdin, os.Stderr)
	restore, err := tty.Setup()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer restore()
	events := tty.StartInput()
	defer tty.StopInput()
	for {
		h, w := tty.Size()
		if h > *maxHeight {
			h = *maxHeight
		}
		tty.UpdateBuffer(nil, widget.Render(w, h), false)
		event := <-events
		handled := widget.Handle(event)
		if !handled && event == term.K('D', ui.Ctrl) {
			tty.UpdateBuffer(nil, term.NewBufferBuilder(w).Buffer(), true)
			break
		}
	}
}

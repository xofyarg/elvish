package cliedit

import (
	"fmt"
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/apptest"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/store/storedefs"
)

var Styles = apptest.Styles

type fixture struct {
	Editor  *Editor
	TTYCtrl cli.TTYCtrl
	Evaler  *eval.Evaler
	Store   storedefs.Store
	Home    string

	width   int
	codeCh  <-chan string
	errCh   <-chan error
	cleanup func()
}

func rc(codes ...string) func(*fixture) {
	return func(f *fixture) { evals(f.Evaler, codes...) }
}

func assign(name string, val interface{}) func(*fixture) {
	return func(f *fixture) {
		f.Evaler.Global["temp"] = vars.NewReadOnly(val)
		evals(f.Evaler, name+` = $temp`)
	}
}

func storeOp(storeFn func(storedefs.Store)) func(*fixture) {
	return func(f *fixture) {
		storeFn(f.Store)
		// TODO(xiaq): Don't depend on this Elvish API.
		evals(f.Evaler, "edit:history:fast-forward")
	}
}

func setup(fns ...func(*fixture)) *fixture {
	st, cleanupStore := store.MustGetTempStore()
	home, cleanupFs := eval.InTempHome()
	tty, ttyCtrl := cli.NewFakeTTY()
	ev := eval.NewEvaler()
	ed := NewEditor(tty, ev, st)
	ev.InstallModule("edit", ed.Ns())
	evals(ev,
		`use edit`,
		// This will simplify most tests against the terminal.
		"edit:rprompt = { }")
	f := &fixture{Editor: ed, TTYCtrl: ttyCtrl, Evaler: ev, Store: st, Home: home}
	for _, fn := range fns {
		fn(f)
	}
	_, f.width = tty.Size()
	f.codeCh, f.errCh = apptest.StartReadCode(f.Editor.ReadCode)
	f.cleanup = func() {
		f.Editor.app.CommitEOF()
		f.Wait()
		cleanupFs()
		cleanupStore()
	}
	return f
}

func (f *fixture) Wait() (string, error) {
	return <-f.codeCh, <-f.errCh
}

func (f *fixture) Cleanup() {
	f.cleanup()
}

func (f *fixture) MakeBuffer(args ...interface{}) *term.Buffer {
	return term.NewBufferBuilder(f.width).MarkLines(args...).Buffer()
}

func (f *fixture) TestTTY(t *testing.T, args ...interface{}) {
	t.Helper()
	f.TTYCtrl.TestBuffer(t, f.MakeBuffer(args...))
}

func (f *fixture) TestTTYNotes(t *testing.T, args ...interface{}) {
	t.Helper()
	f.TTYCtrl.TestNotesBuffer(t, f.MakeBuffer(args...))
}

func feedInput(ttyCtrl cli.TTYCtrl, s string) {
	for _, r := range s {
		ttyCtrl.Inject(term.K(r))
	}
}

func evals(ev *eval.Evaler, codes ...string) {
	for _, code := range codes {
		err := ev.EvalSourceInTTY(eval.NewInteractiveSource(code))
		if err != nil {
			panic(fmt.Errorf("eval %q: %s", code, err))
		}
	}
}

func getGlobal(ev *eval.Evaler, name string) interface{} {
	return ev.Global[name].Get()
}

func testGlobals(t *testing.T, ev *eval.Evaler, wantVals map[string]interface{}) {
	t.Helper()
	for name, wantVal := range wantVals {
		testGlobal(t, ev, name, wantVal)
	}
}

func testGlobal(t *testing.T, ev *eval.Evaler, name string, wantVal interface{}) {
	t.Helper()
	if val := getGlobal(ev, name); !vals.Equal(val, wantVal) {
		t.Errorf("$%s = %s, want %s",
			name, vals.Repr(val, vals.NoPretty), vals.Repr(wantVal, vals.NoPretty))
	}
}

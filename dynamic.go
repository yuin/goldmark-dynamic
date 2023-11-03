// Package dynamic is an extension for goldmark that makes goldmark 'dynamic' .
package dynamic

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

// Extension is a dynamic extension file for goldmark-dynamic.
type Extension struct {
	File    string
	Options any
}

// Option is an option for the goldmark-dynamic extension.
type Option func(*dynamic)

// WithFS is an option that sets [fs.StatFS].
// This defaults to os.StatFS(".") .
func WithFS(f fs.StatFS) Option {
	return func(e *dynamic) {
		e.fs = f
	}
}

// WithExtensions is an option that sets files for scripts.
func WithExtensions(f func() []Extension) Option {
	return func(e *dynamic) {
		e.extensions = f
	}
}

// WithOnError is an option that sets function for script errors.
// By default, goldmark-dynamic panics when script errors occur.
func WithOnError(f func(error)) Option {
	return func(e *dynamic) {
		e.onError = f
	}
}

type options interface {
	OnError() func(error)
}

type dynamic struct {
	fs         fs.StatFS
	extensions func() []Extension
	onError    func(error)
}

// New creates a new goldmark-dynamic extension.
func New(opts ...Option) goldmark.Extender {
	e := &dynamic{
		fs: os.DirFS(".").(fs.StatFS),
		extensions: func() []Extension {
			return []Extension{}
		},
		onError: func(err error) {
			panic(err)
		},
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *dynamic) OnError() func(error) {
	return e.onError
}

func exportGoldmark(l *lua.LState, opts options) {
	l.PreloadModule("goldmark", func(l *lua.LState) int {
		mod := l.NewTable()
		for _, def := range []struct {
			name  string
			value any
		}{
			{
				name:  "withExtensions",
				value: goldmark.WithExtensions,
			},
			{
				name:  "withParser",
				value: goldmark.WithParser,
			},
			{
				name:  "withParserOptions",
				value: goldmark.WithParserOptions,
			},
			{
				name:  "withRenderer",
				value: goldmark.WithRenderer,
			},
			{
				name:  "withRendererOptions",
				value: goldmark.WithRendererOptions,
			},
		} {
			mod.RawSetString(def.name, luar.New(l, def.value))
		}
		l.Push(mod)
		return 1
	})
}

func (e *dynamic) Extend(m goldmark.Markdown) {
	l := lua.NewState()
	exportGoBytes(l, e)
	exportGoldmark(l, e)
	exportGoldmarkUtil(l, e)
	exportGoldmarkText(l, e)
	exportGoldmarkAST(l, e)
	exportGoldmarkParser(l, e)
	exportGoldmarkRenderer(l, e)
	exportGoldmarkRendererHTML(l, e)

	findFile := func(l *lua.LState, name, pname string) (string, string) {
		name = strings.Replace(name, ".", string(os.PathSeparator), -1)
		lv := l.GetField(l.GetField(l.Get(lua.EnvironIndex), "package"), pname)
		path, ok := lv.(lua.LString)
		if !ok {
			l.RaiseError("package.%s must be a string", pname)
		}
		messages := []string{}
		for _, pattern := range strings.Split(string(path), ";") {
			luapath := strings.Replace(pattern, "?", name, -1)
			_, err := e.fs.Stat(luapath)
			if err == nil {
				return luapath, ""
			}
			messages = append(messages, err.Error())
		}
		return "", strings.Join(messages, "\n\t")
	}

	fsLoader := func(l *lua.LState) int {
		name := l.CheckString(1)
		path, msg := findFile(l, name, "path")
		if len(path) == 0 {
			l.Push(lua.LString(msg))
			return 1
		}
		fn, err1 := loadLuaFileFS(l, e.fs, path)
		if err1 != nil {
			l.RaiseError(err1.Error())
		}
		l.Push(fn)
		return 1
	}

	// TODO: support io.* functions?

	loaders, _ := l.GetField(l.Get(lua.RegistryIndex), "_LOADERS").(*lua.LTable)
	loaders.Append(l.NewFunction(fsLoader))

	for _, extension := range e.extensions() {
		fn, err := loadLuaFileFS(l, e.fs, extension.File)
		if err != nil {
			e.onError(err)
		}
		if err := l.CallByParam(lua.P{
			Fn:      fn,
			NRet:    1,
			Protect: true,
		}); err != nil {
			e.onError(err)
		}
		ret := l.Get(-1)
		l.Pop(1)
		if _, err := mustLValue(ret, lua.LTFunction); err != nil {
			e.onError(fmt.Errorf("goldmark-dynamic lua extension returns an invalid value: %w", err))
			return
		}

		if err := l.CallByParam(lua.P{
			Fn:      ret.(*lua.LFunction),
			NRet:    1,
			Protect: true,
		}, luar.New(l, m), luar.New(l, extension.Options)); err != nil {
			e.onError(err)
		}
	}
}

func loadLuaFileFS(l *lua.LState, f fs.FS, path string) (*lua.LFunction, error) {
	file, err := f.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	c, err := reader.ReadByte()
	if err != nil && err != io.EOF {
		return nil, err
	}
	if c == byte('#') {
		_, _, err = reader.ReadLine()
		if err != nil {
			return nil, err
		}
	}

	if err != io.EOF {
		err = reader.UnreadByte()
		if err != nil {
			return nil, err
		}
	}

	return l.Load(reader, path)
}

type propTable struct {
	l       *lua.LState
	Name    string
	Table   *lua.LTable
	onError func(error)
}

func newPropTable(l *lua.LState, name string, table *lua.LTable, onError func(error)) *propTable {
	return &propTable{
		l:       l,
		Name:    name,
		Table:   table,
		onError: onError,
	}
}

func (t *propTable) Get(key string, types ...lua.LValueType) lua.LValue {
	lv, err := mustLValue(t.l.GetField(t.Table, key), types...)
	if err != nil {
		t.onError(fmt.Errorf("%s.%s: %w", t.Name, key, err))
	}
	return lv
}

func (t *propTable) Bool(key string) bool {
	lv := t.l.GetField(t.Table, key)
	return !lua.LVIsFalse(lv)
}

func (t *propTable) Bytes(key string) []byte {
	lv := t.Get(key, lua.LTString)
	var bs []byte
	if lv != lua.LNil {
		bs = []byte(string(lv.(lua.LString)))
	}
	return bs
}

func (t *propTable) Int(key string) int {
	lv := t.Get(key, lua.LTNumber)
	if lv == lua.LNil {
		return 0
	}
	return int(lv.(lua.LNumber))
}

func mustLValue(v lua.LValue, types ...lua.LValueType) (lua.LValue, error) {
	for _, typ := range types {
		if v.Type() == typ {
			return v, nil
		}
	}
	var buf []string
	for _, typ := range types {
		buf = append(buf, typ.String())
	}
	return lua.LNil, fmt.Errorf("must be a %s", strings.Join(buf, " or "))
}

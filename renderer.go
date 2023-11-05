package dynamic

import (
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func exportGoldmarkRenderer(l *lua.LState, opts options) {
	l.PreloadModule("goldmark.renderer", func(l *lua.LState) int {
		mod := l.NewTable()
		for _, def := range []struct {
			name  string
			value any
		}{
			{
				name:  "withNodeRenderers",
				value: renderer.WithNodeRenderers,
			},
		} {
			mod.RawSetString(def.name, luar.New(l, def.value))
		}

		l.Push(mod)
		return 1
	})
}

func exportGoldmarkRendererHTML(l *lua.LState, opts options) {
	l.PreloadModule("goldmark.renderer.html", func(l *lua.LState) int {
		mod := l.NewTable()

		for _, def := range []struct {
			name  string
			value any
		}{
			{
				name:  "withXHTML",
				value: html.WithXHTML,
			},
			{
				name:  "withHardWraps",
				value: html.WithHardWraps,
			},
			{
				name:  "withUnsafe",
				value: html.WithUnsafe,
			},
			{
				name:  "withWriter",
				value: html.WithWriter,
			},
			{
				name:  "withEastAsianLineBreaks",
				value: html.WithEastAsianLineBreaks,
			},
			{
				name:  "withEscapedSpace",
				value: html.WithEscapedSpace,
			},
			{
				name:  "isDangerousURL",
				value: html.IsDangerousURL,
			},
			{
				name:  "renderAttributes",
				value: html.RenderAttributes,
			},
			{
				name:  "globalAttributeFilter",
				value: html.GlobalAttributeFilter,
			},
		} {
			mod.RawSetString(def.name, luar.New(l, def.value))
		}
		mod.RawSetString("newRenderer", l.NewFunction(func(l *lua.LState) int {
			value := newDynamicHTMLRenderer(l, l.CheckTable(1), opts.OnError())
			ud := luar.New(l, value)
			l.Push(ud)
			return 1
		}))

		l.Push(mod)
		return 1
	})
}

var _ renderer.NodeRenderer = (*dynamicHTMLRenderer)(nil)

type dynamicHTMLRenderer struct {
	html.Config

	l       *lua.LState
	props   *lua.LTable
	onError func(error)

	registerFuncs lua.LValue
}

func newDynamicHTMLRenderer(l *lua.LState, props *lua.LTable, onError func(error)) *dynamicHTMLRenderer {
	pt := newPropTable(l, "Renderer", props, onError)
	return &dynamicHTMLRenderer{
		l:       l,
		props:   props,
		onError: onError,

		registerFuncs: pt.Get("registerFuncs", lua.LTFunction),
	}
}

func (r *dynamicHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	if r.registerFuncs == lua.LNil {
		return
	}
	if err := r.l.CallByParam(lua.P{
		Fn:      r.registerFuncs.(*lua.LFunction),
		NRet:    0,
		Protect: true,
	}, luar.New(r.l, r), luar.New(r.l, reg)); err != nil {
		r.onError(err)
	}
}

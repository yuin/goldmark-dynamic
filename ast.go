package dynamic

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func exportGoldmarkAST(l *lua.LState, opts options) {
	l.PreloadModule("goldmark.ast", func(l *lua.LState) int {
		mod := l.NewTable()
		for _, def := range []struct {
			name  string
			value any
		}{
			{
				name:  "newNodeKind",
				value: ast.NewNodeKind,
			},
			{
				name:  "walkContinue",
				value: ast.WalkContinue,
			},
			{
				name:  "walkSkipChildren",
				value: ast.WalkSkipChildren,
			},
			{
				name:  "walkStop",
				value: ast.WalkStop,
			},
			{
				name:  "walk",
				value: ast.Walk,
			},
			{
				name:  "isParagraph",
				value: ast.IsParagraph,
			},
			{
				name:  "mergeOrAppendTextSegment",
				value: ast.MergeOrAppendTextSegment,
			},
			{
				name:  "mergeOrReplaceTextSegment",
				value: ast.MergeOrReplaceTextSegment,
			},
			{
				name:  "kindAutoLink",
				value: ast.KindAutoLink,
			},
			{
				name:  "kindBlockquote",
				value: ast.KindBlockquote,
			},
			{
				name:  "kindCodeBlock",
				value: ast.KindCodeBlock,
			},
			{
				name:  "kindCodeSpan",
				value: ast.KindCodeSpan,
			},
			{
				name:  "kindDocument",
				value: ast.KindDocument,
			},
			{
				name:  "kindEmphasis",
				value: ast.KindEmphasis,
			},
			{
				name:  "kindFencedCodeBlock",
				value: ast.KindFencedCodeBlock,
			},
			{
				name:  "kindHTMLBlock",
				value: ast.KindHTMLBlock,
			},
			{
				name:  "kindHeading",
				value: ast.KindHeading,
			},
			{
				name:  "kindImage",
				value: ast.KindImage,
			},
			{
				name:  "kindLink",
				value: ast.KindLink,
			},
			{
				name:  "kindList",
				value: ast.KindList,
			},
			{
				name:  "kindListItem",
				value: ast.KindListItem,
			},
			{
				name:  "kindParagraph",
				value: ast.KindParagraph,
			},
			{
				name:  "kindRawHTML",
				value: ast.KindRawHTML,
			},
			{
				name:  "kindString",
				value: ast.KindString,
			},
			{
				name:  "kindText",
				value: ast.KindText,
			},
			{
				name:  "kindTextBlock",
				value: ast.KindTextBlock,
			},
			{
				name:  "kindThematicBreak",
				value: ast.KindThematicBreak,
			},
		} {
			mod.RawSetString(def.name, luar.New(l, def.value))
		}

		mod.RawSetString("newInlineNode", l.NewFunction(func(l *lua.LState) int {
			value := newDynamicInlineNode(l, l.CheckTable(1), opts.OnError())
			ud := luar.New(l, value)
			l.Push(ud)
			return 1
		}))

		mod.RawSetString("newBlockNode", l.NewFunction(func(l *lua.LState) int {
			value := newDynamicBlockNode(l, l.CheckTable(1), opts.OnError())
			ud := luar.New(l, value)
			l.Push(ud)
			return 1
		}))

		l.Push(mod)
		return 1
	})
}

var _ ast.Node = (*dynamicInlineNode)(nil)

type dynamicInlineNode struct {
	ast.BaseInline
	l       *lua.LState
	props   *lua.LTable
	onError func(error)

	kind  ast.NodeKind
	isRaw lua.LValue
	p     lua.LValue
}

func newDynamicInlineNode(l *lua.LState, props *lua.LTable, onError func(error)) *dynamicInlineNode {
	pt := newPropTable(l, "InlineNode", props, onError)

	return &dynamicInlineNode{
		l:       l,
		props:   props,
		onError: onError,

		kind:  ast.NodeKind(pt.Int("kind")),
		isRaw: pt.Get("isRaw", lua.LTFunction, lua.LTNil),
		p:     pt.Get("props", lua.LTTable, lua.LTNil),
	}
}

func (n *dynamicInlineNode) Dump(source []byte, level int) {
	p := map[string]string{}
	if n.p != lua.LNil {
		n.p.(*lua.LTable).ForEach(func(key, value lua.LValue) {
			p[key.String()] = value.String()
		})
	}
	ast.DumpHelper(n, source, level, p, nil)
}

func (n *dynamicInlineNode) Kind() ast.NodeKind {
	return n.kind
}

func (n *dynamicInlineNode) IsRaw() bool {
	if n.isRaw == lua.LNil {
		return false
	}
	if err := n.l.CallByParam(lua.P{
		Fn:      n.isRaw.(*lua.LFunction),
		NRet:    1,
		Protect: true,
	}); err != nil {
		n.onError(err)
	}
	ret := n.l.Get(-1)
	n.l.Pop(1)

	if _, err := mustLValue(ret, lua.LTBool); err != nil {
		n.onError(fmt.Errorf("InlineNode:isRaw returns an invalid value: %w", err))
		return false
	}

	return bool(ret.(lua.LBool))
}

func (n *dynamicInlineNode) Prop(name string) any {
	if n.p == lua.LNil {
		return nil
	}
	return n.l.GetField(n.p, name)
}

var _ ast.Node = (*dynamicBlockNode)(nil)

type dynamicBlockNode struct {
	ast.BaseBlock
	l       *lua.LState
	props   *lua.LTable
	onError func(error)

	kind  ast.NodeKind
	isRaw lua.LValue
	p     lua.LValue
}

func newDynamicBlockNode(l *lua.LState, props *lua.LTable, onError func(error)) *dynamicBlockNode {
	pt := newPropTable(l, "BlockNode", props, onError)

	return &dynamicBlockNode{
		l:       l,
		props:   props,
		onError: onError,

		kind:  ast.NodeKind(pt.Int("kind")),
		isRaw: pt.Get("isRaw", lua.LTFunction, lua.LTNil),
		p:     pt.Get("props", lua.LTTable, lua.LTNil),
	}
}

func (n *dynamicBlockNode) Dump(source []byte, level int) {
	p := map[string]string{}
	if n.p != lua.LNil {
		n.p.(*lua.LTable).ForEach(func(key, value lua.LValue) {
			p[key.String()] = value.String()
		})
	}
	ast.DumpHelper(n, source, level, p, nil)
}

func (n *dynamicBlockNode) Kind() ast.NodeKind {
	return n.kind
}

func (n *dynamicBlockNode) IsRaw() bool {
	if n.isRaw == lua.LNil {
		return false
	}
	if err := n.l.CallByParam(lua.P{
		Fn:      n.isRaw.(*lua.LFunction),
		NRet:    1,
		Protect: true,
	}); err != nil {
		n.onError(err)
	}
	ret := n.l.Get(-1)
	n.l.Pop(1)
	if _, err := mustLValue(ret, lua.LTBool); err != nil {
		n.onError(fmt.Errorf("InlineNode:isRaw returns an invalid value: %w", err))
		return false
	}

	return bool(ret.(lua.LBool))
}

func (n *dynamicBlockNode) Prop(name string) any {
	if n.p == lua.LNil {
		return nil
	}
	return n.l.GetField(n.p, name)
}

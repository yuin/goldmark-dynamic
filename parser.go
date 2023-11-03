package dynamic

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func exportGoldmarkParser(l *lua.LState, opts options) {
	l.PreloadModule("goldmark.parser", func(l *lua.LState) int {
		mod := l.NewTable()
		for _, def := range []struct {
			name  string
			value any
		}{
			{
				name:  "none",
				value: parser.None,
			},
			{
				name:  "close",
				value: parser.Close,
			},
			{
				name:  "continue",
				value: parser.Continue,
			},
			{
				name:  "hasChildren",
				value: parser.HasChildren,
			},
			{
				name:  "noChildren",
				value: parser.NoChildren,
			},
			{
				name:  "requireParagraph",
				value: parser.RequireParagraph,
			},
			{
				name:  "newContextKey",
				value: parser.NewContextKey,
			},
			{
				name:  "scanDelimiter",
				value: parser.ScanDelimiter,
			},
			{
				name:  "withInlineParsers",
				value: parser.WithInlineParsers,
			},
			{
				name:  "withBlockParsers",
				value: parser.WithBlockParsers,
			},
			{
				name:  "withASTTransformers",
				value: parser.WithASTTransformers,
			},
			{
				name:  "withParagraphTransformers",
				value: parser.WithParagraphTransformers,
			},
		} {
			mod.RawSetString(def.name, luar.New(l, def.value))
		}

		mod.RawSetString("newInlineParser", l.NewFunction(func(l *lua.LState) int {
			parser := newDynamicInlineParser(l, l.CheckTable(1), opts.OnError())
			ud := luar.New(l, parser)
			l.Push(ud)
			return 1
		}))

		mod.RawSetString("newBlockParser", l.NewFunction(func(l *lua.LState) int {
			parser := newDynamicBlockParser(l, l.CheckTable(1), opts.OnError())
			ud := luar.New(l, parser)
			l.Push(ud)
			return 1
		}))

		mod.RawSetString("newASTTransformer", l.NewFunction(func(l *lua.LState) int {
			parser := newDynamicASTTransformer(l, l.CheckTable(1), opts.OnError())
			ud := luar.New(l, parser)
			l.Push(ud)
			return 1
		}))

		mod.RawSetString("newParagraphTransformer", l.NewFunction(func(l *lua.LState) int {
			parser := newDynamicParagraphTransformer(l, l.CheckTable(1), opts.OnError())
			ud := luar.New(l, parser)
			l.Push(ud)
			return 1
		}))
		mod.RawSetString("newDelimiterProcessor", l.NewFunction(func(l *lua.LState) int {
			parser := newDynamicDelimiterProcessor(l, l.CheckTable(1), opts.OnError())
			ud := luar.New(l, parser)
			l.Push(ud)
			return 1
		}))

		l.Push(mod)
		return 1
	})
}

var _ parser.InlineParser = (*dynamicInlineParser)(nil)

type dynamicInlineParser struct {
	l       *lua.LState
	props   *lua.LTable
	onError func(error)

	trigger    []byte
	parse      lua.LValue
	closeBlock lua.LValue
}

func newDynamicInlineParser(l *lua.LState, props *lua.LTable, onError func(error)) *dynamicInlineParser {
	pt := newPropTable(l, "InlineParser", props, onError)

	return &dynamicInlineParser{
		l:       l,
		props:   props,
		onError: onError,

		trigger:    pt.Bytes("triggers"),
		parse:      pt.Get("parse", lua.LTFunction),
		closeBlock: pt.Get("closeBlock", lua.LTFunction, lua.LTNil),
	}
}

func (s *dynamicInlineParser) Trigger() []byte {
	return s.trigger
}

func (s *dynamicInlineParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	if s.parse == lua.LNil {
		return nil
	}
	l := s.l

	if err := l.CallByParam(lua.P{
		Fn:      s.parse.(*lua.LFunction),
		NRet:    1,
		Protect: true,
	}, luar.New(l, s), luar.New(l, parent), luar.New(l, block), luar.New(l, pc)); err != nil {
		s.onError(err)
	}
	ret := l.Get(-1)
	l.Pop(1)
	if ret == lua.LNil {
		return nil
	}

	if _, err := mustLValue(ret, lua.LTUserData); err != nil {
		s.onError(fmt.Errorf("InlineParser.parse returns an invalid value: %w", err))
		return nil
	}
	node, ok := ret.(*lua.LUserData).Value.(ast.Node)
	if !ok {
		s.onError(fmt.Errorf("InlineParser.open must return an ast.Node"))
	}

	return node
}

func (s *dynamicInlineParser) CloseBlock(parent ast.Node, block text.Reader, pc parser.Context) {
	if s.closeBlock == lua.LNil {
		return
	}
	l := s.l

	if err := l.CallByParam(lua.P{
		Fn:      s.closeBlock.(*lua.LFunction),
		NRet:    0,
		Protect: true,
	}, luar.New(l, s), luar.New(l, parent), luar.New(l, block), luar.New(l, pc)); err != nil {
		s.onError(err)
	}
}

var _ parser.BlockParser = (*dynamicBlockParser)(nil)

type dynamicBlockParser struct {
	l       *lua.LState
	props   *lua.LTable
	onError func(error)

	trigger               []byte
	fopen                 lua.LValue
	fcontinue             lua.LValue
	fclose                lua.LValue
	canInterruptParagraph bool
	canAcceptIndentedLine bool
}

func newDynamicBlockParser(l *lua.LState, props *lua.LTable, onError func(error)) *dynamicBlockParser {
	pt := newPropTable(l, "BlockParser", props, onError)
	trigger := pt.Bytes("triggers")
	if len(trigger) == 0 {
		onError(fmt.Errorf("Can not define BlockParser without triggers"))
	}

	return &dynamicBlockParser{
		l:       l,
		props:   props,
		onError: onError,

		trigger:               trigger,
		fopen:                 pt.Get("open", lua.LTFunction),
		fcontinue:             pt.Get("continue", lua.LTFunction),
		fclose:                pt.Get("close", lua.LTFunction),
		canInterruptParagraph: pt.Bool("canInterruptParagraph"),
		canAcceptIndentedLine: pt.Bool("canAcceptIndentedLine"),
	}
}

func (s *dynamicBlockParser) Trigger() []byte {
	return s.trigger
}

func (s *dynamicBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	if s.fopen == lua.LNil {
		return nil, parser.Close
	}
	l := s.l

	if err := l.CallByParam(lua.P{
		Fn:      s.fopen.(*lua.LFunction),
		NRet:    2,
		Protect: true,
	}, luar.New(l, s), luar.New(l, parent), luar.New(l, reader), luar.New(l, pc)); err != nil {
		s.onError(err)
	}
	ret2 := l.Get(-1)
	l.Pop(1)
	ret1 := l.Get(-1)
	l.Pop(1)
	if ret1 == lua.LNil {
		return nil, parser.Close
	}

	if _, err := mustLValue(ret1, lua.LTUserData); err != nil {
		s.onError(fmt.Errorf("BlockParser.open returns an invalid value: %w", err))
		return nil, parser.Close
	}
	if _, err := mustLValue(ret2, lua.LTNumber); err != nil {
		s.onError(fmt.Errorf("BlockParser.open returns an invalid value: %w", err))
		return nil, parser.Close
	}
	ud, ok := ret1.(*lua.LUserData).Value.(ast.Node)
	if !ok {
		s.onError(fmt.Errorf("BlockParser.open must return an ast.Node"))
	}

	return ud, parser.State(int(ret2.(lua.LNumber)))

}

func (s *dynamicBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	if s.fcontinue == lua.LNil {
		return parser.Close
	}
	l := s.l

	if err := l.CallByParam(lua.P{
		Fn:      s.fcontinue.(*lua.LFunction),
		NRet:    1,
		Protect: true,
	}, luar.New(l, s), luar.New(l, node), luar.New(l, reader), luar.New(l, pc)); err != nil {
		s.onError(err)
	}
	ret1 := l.Get(-1)
	l.Pop(1)
	if _, err := mustLValue(ret1, lua.LTNumber); err != nil {
		s.onError(fmt.Errorf("BlockParser.continue returns an invalid value: %w", err))
		return parser.Close
	}
	return parser.State(int(ret1.(lua.LNumber)))
}

func (s *dynamicBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	if s.fclose == lua.LNil {
		return
	}

	l := s.l

	if err := l.CallByParam(lua.P{
		Fn:      s.fclose.(*lua.LFunction),
		NRet:    0,
		Protect: true,
	}, luar.New(l, s), luar.New(l, node), luar.New(l, reader), luar.New(l, pc)); err != nil {
		s.onError(err)
	}
}

func (s *dynamicBlockParser) CanInterruptParagraph() bool {
	return s.canInterruptParagraph
}

func (s *dynamicBlockParser) CanAcceptIndentedLine() bool {
	return s.canAcceptIndentedLine
}

var _ parser.ASTTransformer = (*dynamicASTTransformer)(nil)

type dynamicASTTransformer struct {
	l       *lua.LState
	props   *lua.LTable
	onError func(error)

	transform lua.LValue
}

func newDynamicASTTransformer(l *lua.LState, props *lua.LTable, onError func(error)) *dynamicASTTransformer {
	pt := newPropTable(l, "ASTTransformer", props, onError)

	return &dynamicASTTransformer{
		l:       l,
		props:   props,
		onError: onError,

		transform: pt.Get("transform", lua.LTFunction),
	}
}

func (s *dynamicASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	if s.transform == lua.LNil {
		return
	}
	l := s.l

	if err := l.CallByParam(lua.P{
		Fn:      s.transform.(*lua.LFunction),
		NRet:    0,
		Protect: true,
	}, luar.New(l, s), luar.New(l, node), luar.New(l, reader), luar.New(l, pc)); err != nil {
		s.onError(err)
	}
}

var _ parser.ParagraphTransformer = (*dynamicParagraphTransformer)(nil)

type dynamicParagraphTransformer struct {
	l       *lua.LState
	props   *lua.LTable
	onError func(error)

	transform lua.LValue
}

func newDynamicParagraphTransformer(l *lua.LState, props *lua.LTable,
	onError func(error)) *dynamicParagraphTransformer {
	pt := newPropTable(l, "ParagraphTransformer", props, onError)

	return &dynamicParagraphTransformer{
		l:       l,
		props:   props,
		onError: onError,

		transform: pt.Get("transform", lua.LTFunction),
	}
}

func (s *dynamicParagraphTransformer) Transform(node *ast.Paragraph, reader text.Reader, pc parser.Context) {
	if s.transform == lua.LNil {
		return
	}
	l := s.l

	if err := l.CallByParam(lua.P{
		Fn:      s.transform.(*lua.LFunction),
		NRet:    0,
		Protect: true,
	}, luar.New(l, s), luar.New(l, node), luar.New(l, reader), luar.New(l, pc)); err != nil {
		s.onError(err)
	}
}

var _ parser.DelimiterProcessor = (*dynamicDelimiterProcessor)(nil)

type dynamicDelimiterProcessor struct {
	l       *lua.LState
	props   *lua.LTable
	onError func(error)

	isDelimiter   lua.LValue
	canOpenCloser lua.LValue
	onMatch       lua.LValue
}

func newDynamicDelimiterProcessor(l *lua.LState, props *lua.LTable, onError func(error)) *dynamicDelimiterProcessor {
	pt := newPropTable(l, "ParagraphTransformer", props, onError)

	return &dynamicDelimiterProcessor{
		l:       l,
		props:   props,
		onError: onError,

		isDelimiter:   pt.Get("isDelimiter", lua.LTFunction),
		canOpenCloser: pt.Get("canOpenCloser", lua.LTFunction),
		onMatch:       pt.Get("onMatch", lua.LTFunction),
	}
}

func (s *dynamicDelimiterProcessor) IsDelimiter(b byte) bool {
	l := s.l

	if err := l.CallByParam(lua.P{
		Fn:      s.isDelimiter.(*lua.LFunction),
		NRet:    1,
		Protect: true,
	}, luar.New(l, s), luar.New(l, b)); err != nil {
		s.onError(err)
	}
	ret := l.Get(-1)
	l.Pop(1)
	_, err := mustLValue(ret, lua.LTBool)
	if err != nil {
		s.onError(fmt.Errorf("DelimiterProcessor.isDelimiter returns an invalid value: %w", err))
	}
	return bool(ret.(lua.LBool))

}

func (s *dynamicDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	l := s.l

	if err := l.CallByParam(lua.P{
		Fn:      s.canOpenCloser.(*lua.LFunction),
		NRet:    1,
		Protect: true,
	}, luar.New(l, s), luar.New(l, opener), luar.New(l, closer)); err != nil {
		s.onError(err)
	}
	ret := l.Get(-1)
	l.Pop(1)
	_, err := mustLValue(ret, lua.LTBool)
	if err != nil {
		s.onError(fmt.Errorf("DelimiterProcessor.canOpenCloser returns an invalid value: %w", err))
	}
	return bool(ret.(lua.LBool))

}

func (s *dynamicDelimiterProcessor) OnMatch(consumes int) ast.Node {
	l := s.l

	if err := l.CallByParam(lua.P{
		Fn:      s.onMatch.(*lua.LFunction),
		NRet:    1,
		Protect: true,
	}, luar.New(l, consumes)); err != nil {
		s.onError(err)
	}
	ret := l.Get(-1)
	l.Pop(1)
	if ret == lua.LNil {
		return nil
	}

	if _, err := mustLValue(ret, lua.LTUserData); err != nil {
		s.onError(fmt.Errorf("DelimiterProcessor.onMatch returns an invalid value: %w", err))
		return nil
	}
	node, ok := ret.(*lua.LUserData).Value.(ast.Node)
	if !ok {
		s.onError(fmt.Errorf("DelimiterProcessor.onMatch must return an ast.Node"))
	}

	return node
}

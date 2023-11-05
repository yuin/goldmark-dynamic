package dynamic

import (
	"github.com/yuin/goldmark/text"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func exportGoldmarkText(l *lua.LState, opts options) {
	l.PreloadModule("goldmark.text", func(l *lua.LState) int {
		mod := l.NewTable()
		mod.RawSetString("FindClosureOptions", luar.NewType(l, text.FindClosureOptions{}))

		mt := l.NewTypeMetatable("Segment")
		mod.RawSetString("Segment", mt)
		l.SetField(mt, "new", l.NewFunction(func(l *lua.LState) int {
			n := l.GetTop()
			ud := l.NewUserData()
			l.SetMetatable(ud, l.GetTypeMetatable("Segment"))
			start := l.CheckNumber(1)
			stop := l.CheckNumber(2)
			if n == 2 {
				ud.Value = text.NewSegment(int(start), int(stop))
			}
			if n == 3 {
				padding := l.CheckNumber(3)
				ud.Value = text.NewSegmentPadding(int(start), int(stop), int(padding))
			}
			l.Push(ud)
			return 1
		}))
		l.SetField(mt, "__index", l.NewFunction(func(l *lua.LState) int {
			ud := l.CheckUserData(1)
			prop := l.CheckString(2)
			switch prop {
			case "start":
				l.Push(lua.LNumber(ud.Value.(text.Segment).Start))
			case "stop":
				l.Push(lua.LNumber(ud.Value.(text.Segment).Stop))
			case "padding":
				l.Push(lua.LNumber(ud.Value.(text.Segment).Padding))
			default:
				l.Push(lua.LNil)
			}
			return 1
		}))
		l.Push(mod)
		return 1
	})
	l.PreloadModule("goldmark.text.segment", func(l *lua.LState) int {
		mod := l.NewTable()
		for _, def := range []struct {
			name  string
			value any
		}{
			{
				name: "value",
				value: func(s text.Segment, buffer []byte) []byte {
					return s.Value(buffer)
				},
			},
			{
				name: "len",
				value: func(s text.Segment) int {
					return s.Len()
				},
			},
			{
				name: "between",
				value: func(s, s2 text.Segment) text.Segment {
					return s.Between(s2)
				},
			},
			{
				name: "isEmpty",
				value: func(s text.Segment) bool {
					return s.IsEmpty()
				},
			},
			{
				name: "trimRightSpace",
				value: func(s text.Segment, buffer []byte) text.Segment {
					return s.TrimRightSpace(buffer)
				},
			},
			{
				name: "trimLeftSpace",
				value: func(s text.Segment, buffer []byte) text.Segment {
					return s.TrimLeftSpace(buffer)
				},
			},
			{
				name: "trimLeftSpaceWidth",
				value: func(s text.Segment, width int, buffer []byte) text.Segment {
					return s.TrimLeftSpaceWidth(width, buffer)
				},
			},
			{
				name: "withStart",
				value: func(s text.Segment, v int) text.Segment {
					return s.WithStart(v)
				},
			},
			{
				name: "withStop",
				value: func(s text.Segment, v int) text.Segment {
					return s.WithStop(v)
				},
			},
			{
				name: "concatPadding",
				value: func(s text.Segment, buffer []byte) []byte {
					return s.ConcatPadding(buffer)
				},
			},
		} {
			mod.RawSetString(def.name, luar.New(l, def.value))
		}

		l.Push(mod)
		return 1
	})
}

package dynamic

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func exportGoBytes(l *lua.LState, opts options) {
	l.PreloadModule("go.bytes", func(l *lua.LState) int {
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
				name:  "clone",
				value: bytes.Clone,
			},
			{
				name:  "compare",
				value: bytes.Compare,
			},
			{
				name:  "contains",
				value: bytes.Contains,
			},
			{
				name:  "containsAny",
				value: bytes.ContainsAny,
			},
			{
				name:  "containsRune",
				value: bytes.ContainsRune,
			},
			{
				name:  "count",
				value: bytes.Count,
			},
			{
				name:  "cut",
				value: bytes.Cut,
			},
			{
				name:  "cutPrefix",
				value: bytes.CutPrefix,
			},
			{
				name:  "cutSuffix",
				value: bytes.CutSuffix,
			},
			{
				name:  "equal",
				value: bytes.Equal,
			},
			{
				name:  "equalFold",
				value: bytes.EqualFold,
			},
			{
				name:  "fields",
				value: bytes.Fields,
			},
			{
				name:  "fieldsFunc",
				value: bytes.FieldsFunc,
			},
			{
				name:  "hasPrefix",
				value: bytes.HasPrefix,
			},
			{
				name:  "hasSuffix",
				value: bytes.HasSuffix,
			},
			{
				name:  "index",
				value: bytes.Index,
			},
			{
				name:  "indexAny",
				value: bytes.IndexAny,
			},
			{
				name:  "indexByte",
				value: bytes.IndexByte,
			},
			{
				name:  "indexFunc",
				value: bytes.IndexFunc,
			},
			{
				name:  "indexRune",
				value: bytes.IndexRune,
			},
			{
				name:  "join",
				value: bytes.Join,
			},
			{
				name:  "lastIndex",
				value: bytes.LastIndex,
			},
			{
				name:  "lastIndexAny",
				value: bytes.LastIndexAny,
			},
			{
				name:  "lastIndexByte",
				value: bytes.LastIndexByte,
			},
			{
				name:  "lastIndexFunc",
				value: bytes.LastIndexFunc,
			},
			{
				name:  "map",
				value: bytes.Map,
			},
			{
				name:  "repeat",
				value: bytes.Repeat,
			},
			{
				name:  "replace",
				value: bytes.Replace,
			},
			{
				name:  "replaceAll",
				value: bytes.ReplaceAll,
			},
			{
				name:  "runes",
				value: bytes.Runes,
			},
			{
				name:  "split",
				value: bytes.Split,
			},
			{
				name:  "splitAfter",
				value: bytes.SplitAfter,
			},
			{
				name:  "splitAfterN",
				value: bytes.SplitAfterN,
			},
			{
				name:  "splitN",
				value: bytes.SplitN,
			},
			{
				name:  "toLower",
				value: bytes.ToLower,
			},
			{
				name:  "toLowerSpecial",
				value: bytes.ToLowerSpecial,
			},
			{
				name:  "toTitle",
				value: bytes.ToTitle,
			},
			{
				name:  "toTitleSpecial",
				value: bytes.ToTitleSpecial,
			},
			{
				name:  "toUpper",
				value: bytes.ToUpper,
			},
			{
				name:  "toUpperSpecial",
				value: bytes.ToUpperSpecial,
			},
			{
				name:  "toValidUTF8",
				value: bytes.ToValidUTF8,
			},
			{
				name:  "trim",
				value: bytes.Trim,
			},
			{
				name:  "trimFunc",
				value: bytes.TrimFunc,
			},
			{
				name:  "trimLeft",
				value: bytes.TrimLeft,
			},
			{
				name:  "trimLeftFunc",
				value: bytes.TrimLeftFunc,
			},
			{
				name:  "trimPrefix",
				value: bytes.TrimPrefix,
			},
			{
				name:  "trimRight",
				value: bytes.TrimRight,
			},
			{
				name:  "trimRightFunc",
				value: bytes.TrimRightFunc,
			},
			{
				name:  "trimSpace",
				value: bytes.TrimSpace,
			},
			{
				name:  "trimSuffix",
				value: bytes.TrimSuffix,
			},
			{
				name: "sub",
				value: func(b []byte, from, to int) []byte {
					return b[from:to]
				},
			},
			{
				name: "equalString",
				value: func(b []byte, s string) bool {
					return bytes.Equal(b, util.StringToReadOnlyBytes(s))

				},
			},
			{
				name: "string",
				value: func(b []byte) string {
					return string(b)
				},
			},
			{
				name: "fromString",
				value: func(s string) []byte {
					return []byte(s)
				},
			},
		} {
			mod.RawSetString(def.name, luar.New(l, def.value))
		}
		for _, def := range []struct {
			name  string
			value any
		}{
			{
				name:  "Buffer",
				value: bytes.Buffer{},
			},
			{
				name:  "Reader",
				value: bytes.Reader{},
			},
		} {
			mod.RawSetString(def.name, luar.NewType(l, def.value))
		}

		l.Push(mod)
		return 1
	})
}

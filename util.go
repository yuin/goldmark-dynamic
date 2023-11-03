package dynamic

import (
	"github.com/yuin/goldmark/util"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func exportGoldmarkUtil(l *lua.LState, opts options) {
	l.PreloadModule("goldmark.util", func(l *lua.LState) int {
		mod := l.NewTable()
		for _, def := range []struct {
			name  string
			value any
		}{
			{
				name:  "prioritized",
				value: util.Prioritized,
			},
			{
				name:  "bytesToReadOnlyString",
				value: util.BytesToReadOnlyString,
			},
			{
				name:  "doFullUnicodeCaseFolding",
				value: util.DoFullUnicodeCaseFolding,
			},
			{
				name:  "eastAsianWidth",
				value: util.EastAsianWidth,
			},
			{
				name:  "escapeHTML",
				value: util.EscapeHTML,
			},
			{
				name:  "escapeHTMLByte",
				value: util.EscapeHTMLByte,
			},
			{
				name:  "findEmailIndex",
				value: util.FindEmailIndex,
			},
			{
				name:  "findURLIndex",
				value: util.FindURLIndex,
			},
			{
				name:  "firstNonSpacePosition",
				value: util.FirstNonSpacePosition,
			},
			{
				name:  "indentPosition",
				value: util.IndentPosition,
			},
			{
				name:  "indentPositionPadding",
				value: util.IndentPositionPadding,
			},
			{
				name:  "indentWidth",
				value: util.IndentWidth,
			},
			{
				name:  "isAlphaNumeric",
				value: util.IsAlphaNumeric,
			},
			{
				name:  "isBlank",
				value: util.IsBlank,
			},
			{
				name:  "isEastAsianWideRune",
				value: util.IsEastAsianWideRune,
			},
			{
				name:  "isEscapedPunctuation",
				value: util.IsEscapedPunctuation,
			},
			{
				name:  "isHexDecimal",
				value: util.IsHexDecimal,
			},
			{
				name:  "isNumeric",
				value: util.IsNumeric,
			},
			{
				name:  "isPunct",
				value: util.IsPunct,
			},
			{
				name:  "isPunctRune",
				value: util.IsPunctRune,
			},
			{
				name:  "isSpace",
				value: util.IsSpace,
			},
			{
				name:  "isSpaceDiscardingUnicodeRune",
				value: util.IsSpaceDiscardingUnicodeRune,
			},
			{
				name:  "isSpaceRune",
				value: util.IsSpaceRune,
			},
			{
				name:  "readWhile",
				value: util.ReadWhile,
			},
			{
				name:  "replaceSpaces",
				value: util.ReplaceSpaces,
			},
			{
				name:  "resolveEntityNames",
				value: util.ResolveEntityNames,
			},
			{
				name:  "resolveNumericReferences",
				value: util.ResolveNumericReferences,
			},
			{
				name:  "stringToReadOnlyBytes",
				value: util.StringToReadOnlyBytes,
			},
			{
				name:  "tabWidth",
				value: util.TabWidth,
			},
			{
				name:  "toLinkReference",
				value: util.ToLinkReference,
			},
			{
				name:  "toRune",
				value: util.ToRune,
			},
			{
				name:  "toValidRune",
				value: util.ToValidRune,
			},
			{
				name:  "trimLeft",
				value: util.TrimLeft,
			},
			{
				name:  "trimLeftLength",
				value: util.TrimLeftLength,
			},
			{
				name:  "trimLeftSpace",
				value: util.TrimLeftSpace,
			},
			{
				name:  "trimLeftSpaceLength",
				value: util.TrimLeftSpaceLength,
			},
			{
				name:  "trimRight",
				value: util.TrimRight,
			},
			{
				name:  "trimRightLength",
				value: util.TrimRightLength,
			},
			{
				name:  "trimRightSpace",
				value: util.TrimRightSpace,
			},
			{
				name:  "trimRightSpaceLength",
				value: util.TrimRightSpaceLength,
			},
			{
				name:  "urlEscape",
				value: util.URLEscape,
			},
			{
				name:  "utf8Len",
				value: util.UTF8Len,
			},
			{
				name:  "unescapePunctuations",
				value: util.UnescapePunctuations,
			},
			{
				name:  "visualizeSpaces",
				value: util.VisualizeSpaces,
			},
			{
				name:  "emptyBytesFilter",
				value: util.NewBytesFilter(),
			},
		} {
			mod.RawSetString(def.name, luar.New(l, def.value))
		}
		l.Push(mod)
		return 1
	})

	l.PreloadModule("bit32", func(l *lua.LState) int {
		mod := l.NewTable()
		for _, def := range []struct {
			name  string
			value any
		}{
			{
				name: "bor",
				value: func(v1, v2 int) int {
					return v1 | v2
				},
			},
		} {
			mod.RawSetString(def.name, luar.New(l, def.value))
		}
		l.Push(mod)
		return 1
	})
}

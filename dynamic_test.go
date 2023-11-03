package dynamic_test

import (
	"testing"

	. "github.com/yuin/goldmark-dynamic"
	"github.com/yuin/goldmark/testutil"

	"github.com/yuin/goldmark"
)

func TestDynamic(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			New(
				WithExtensions(func() []Extension {
					return []Extension{
						{
							File: "_examples/mention.lua",
							Options: map[string]string{
								"class": "user-mention",
							},
						},
						{
							File: "_examples/admonition.lua",
							Options: map[string]string{
								"prefix": "admonition-",
							},
						},
					}
				}),
			),
		),
	)

	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          1,
			Description: "Inline and block parsers",
			Markdown: `
@yuin aaa

::: note
bbbb
*ccc*
:::
`,
			Expected: `
<p><span class="user-mention">@yuin</span> aaa</p>
<div class="admonition-note"><p>bbbb
<em>ccc</em></p>
</div>
`,
		},
		t,
	)
}

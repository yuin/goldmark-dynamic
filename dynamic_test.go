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
				WithExtensions([]Extension{
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
					{
						File: "_examples/open_in_new_window.lua",
						Options: map[string]string{
							"base": "http://self.example.com",
						},
					},
				}),
			),
		),
	)

	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          1,
			Description: "Inline parsers, block parsers and AST transformers",
			Markdown: `
@yuin aaa

::: note
bbbb
*ccc*
:::

[link1](/index.html)
[link2](http://self.example.com)
[external link](http://external.example.com)
`,
			Expected: `
<p><span class="user-mention">@yuin</span> aaa</p>
<div class="admonition-note"><p>bbbb
<em>ccc</em></p>
</div><p><a href="/index.html">link1</a>
<a href="http://self.example.com">link2</a>
<a href="http://external.example.com" target="_blank">external link</a></p>
`,
		},
		t,
	)
}

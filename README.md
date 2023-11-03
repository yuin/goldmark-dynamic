goldmark-dynamic
=========================

[![GoDev][godev-image]][godev-url]

[godev-image]: https://pkg.go.dev/badge/github.com/yuin/goldmark-dynamic
[godev-url]: https://pkg.go.dev/github.com/yuin/goldmark-dynamic

goldmark-dynamic is an extension for the [goldmark](http://github.com/yuin/goldmark) 
that allows loading extensions without re-compilation.

goldmark-dynamic can load extensions written in Lua.

Supported Go versions
--------------------
`>=1.20`

Installation
--------------------

```
go get github.com/yuin/goldmark-dynamic
```

Status
--------------------
This project is an experimental project.
API may changes, but almost API is stable enough for most users.

Limitations
--------------------

- **This extension is not goroutine safe.** Please note that a goldmark with this extension can not be used by multiple goroutines.
- **Lua extension is much slower than Go extension.**  Do not recklessly use goldmark-dynamic! You should use this extension only if you really do not want to recompile applications even get the sacrifice of the performance.
- TODO: This extension does not export all functionalities of the goldmark for now. Exporting goldmark functionalities is a monotonous work(I got bored of doing the work over and over), so contributions are welcome.

Architecture
--------------------
- This extension uses the [gopher-lua](https://github.com/yuin/gopher-lua) as a Lua interpreter.
  - gopher-lua is a fast(in comparison with other script languages written in pure Go), easy to integrate pure Go implementation of the Lua language. Importantly, author of the gopher-lua is the identical person of which is the goldmark author(and this extension) :-).
- Most objects are converted with [gopher-luar](https://github.com/layeh/gopher-luar).

Usage
--------------------
### Go API

```go
import (
    "bytes"
    "fmt"
    "os"

    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark-dynamic"
)

func main() {
    markdown := goldmark.New(
        goldmark.WithExtensions(
            dynamic.New(
                  dynamic.WithExtensions(func() []dynamic.Extension {
                    return []dynamic.Extension{
                        {
                            File: "mention.lua",
                            Options: map[string]string{
                                "class": "user-mention",
                            },
                        },
                    }
                  }),
                  dynamic.WithOnError(func(err error) {
                     fmt.Fprintln(os.Stderr, err.Error())
                  }),
            ),
        ),
    )
    // codes use markdown
}
```

See `dynamic_test.go` for detailed usage.

Since Lua is a dynamic language, unexpected errors may orccur at a runtime. 
You can set a function that will be called if such errors occur. Default
`OnError` just panics if errors occur.

### Lua API
This extension preloads below modules:

| package name | |
| ------------ | ------------------- |
| `go.bytes`   | exports Go's bytes package functionalities |
| `goldmark.ast`   | exports goldmark/ast functionalities |
| `goldmark.parser`   | exports goldmark/parser package functionalities |
| `goldmark.renderer.html`   | exports goldmark/renderer/htm functionalities |
| `goldmark.renderer`   | exports goldmark/renderer functionalities |
| `goldmark.text.segment`   | exports goldmark/text.Segment functionalities |
| `goldmark.text`   | exports goldmark/text package functionalities |
| `goldmark.uti`   | exports goldmark/util package functionalities |

See `_examples` directory for detailed usage.

Most of Lua API is exported by [gopher-luar](https://github.com/layeh/gopher-luar). So you can access simple properties by `obj.propName` and access methods by `obj:propName`.

Dynamic extensions are almost same as extensions written in Go. Basic structure is like the following:

```lua
-- loads required packages
local bit32 = require 'bit32'
local bytes = require 'go.bytes'
local gparser = require 'goldmark.parser'
local gutil = require 'goldmark.util'
local gtext = require 'goldmark.text'
local gsegment = require 'goldmark.text.segment'
local gast = require 'goldmark.ast'
local grenderer = require 'goldmark.renderer'
local hrenderer = require 'goldmark.renderer.html'

-- assign properties to local variables for the performance(this is a common technique in Lua scripts)
local format = string.format
local bstring, bsub, btrimspace, bequlstring, bfromstring = bytes.string, bytes.sub, bytes.trimSpace, bytes.equalString, bytes.fromString
local isspace, prioritized = gutil.isSpace, gutil.prioritized
local seglen = gsegment.len
local bitor = bit32.bor
local walkcontinue = gast.walkContinue
local pclose, pnochildren, phaschildren, pcontinue = gparser.close, gparser.noChildren, gparser.hasChildren, gparser.continue

-- define my AST node kind
local kindadmonition = gast.newNodeKind("admonition")

-- a function applies this extension to a goldmark.
-- This function takes a goldmark.Markdown and arbitary options.
return function(m, opts) 
  -- creates a new BlockParser(, Inline Parser, AST Transformers...)
  local admonitionBlockParser = gparser.newBlockParser({
    triggers = ":",
    open = function(self, parent, reader, pc)
      local line, segment = reader:peekLine()
      if not bequlstring(bsub(line, 0, 3), ":::") then
        return nil, pnochildren
      end
      local class = btrimspace(bsub(line, 3, #line))
      if not class or #class == 0 then
        class = bfromstring("admonition")
      end
      local node = gast.newBlockNode({
        kind = kindadmonition,
        props = {
          class = class
        }
      })
      reader:advance(seglen(segment) - 1)
      return node, phaschildren
    end,
    continue = function(self, node, reader ,pc)
      local line, segment = reader:peekLine()
      if bequlstring(bsub(line, 0, 3), ":::") then
        reader:advance(seglen(segment) - 1)
        return pclose
      end
      return bitor(phaschildren, pcontinue)
    end,
    close = function(self, node, reader, pc)
      -- nothing to do
    end
  })
  
  -- creates a new HTML renderer
  local prefix = opts.prefix or ""
  local admonitionHTMLRenderer = hrenderer.newRenderer({
    registerFuncs = function(self, reg)
      reg:register(kindadmonition,  function(w, source, n, entering) 
        if entering then
          w:writeString(format("<div class=\"%s%s\">", prefix, bstring(n:prop("class"))))
        else
          w:writeString("</div>")
        end
        return walkcontinue, nil
      end)
    end
  })


  -- adds our functionalities to a goldmark.Markdown
  m:parser():addOptions(
    gparser.withBlockParsers(
      prioritized(admonitionBlockParser, 999)
    )
  )
  m:renderer():addOptions(
    grenderer.withNodeRenderers(
      prioritized(admonitionHTMLRenderer, 999)
    ),
    hrenderer.withXHTML()
  )
end
```

Note that goldmark heavily uses `[]byte`. `go.bytes` package simply exports Go functions by gopher-luar, so these functions use 0-started index unlike Lua functions(Lua has an 1-started index).

### For dynamic extension authors
It is recommended that dynamic extensions have a name prefixed with `goldmark-dynamic-` allow users to distinguish a language in which an extension written. For instance, `goldmark-dynamic-admonition'(an extension written in Lua) and `goldmark-admonition`(an extension written in Go).

### List of dynamic extensions
Please let me known your dynamic extensions by a pull requst that updates the list.

- [_examples](https://github.com/yuin/goldmark-dynamic/_examples) : dynamic extension examples

TODO
--------------------
Contributions are welcome.

- Export rest of goldmark functionalities
- Write tests
- Add convinience gopher-lua libraries like [gluare](https://github.com/yuin/gluare), [glua-lfs](https://github.com/layeh/gopher-lfs) etc.

License
--------------------
MIT

Author
--------------------
Yusuke Inuzuka

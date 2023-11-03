local bit32 = require 'bit32'
local bytes = require 'go.bytes'
local gparser = require 'goldmark.parser'
local gutil = require 'goldmark.util'
local gtext = require 'goldmark.text'
local gsegment = require 'goldmark.text.segment'
local gast = require 'goldmark.ast'
local grenderer = require 'goldmark.renderer'
local hrenderer = require 'goldmark.renderer.html'

local format = string.format
local bstring, bsub, btrimspace, bequlstring, bfromstring = bytes.string, bytes.sub, bytes.trimSpace, bytes.equalString, bytes.fromString
local isspace, prioritized = gutil.isSpace, gutil.prioritized
local seglen = gsegment.len
local bitor = bit32.bor
local walkcontinue = gast.walkContinue
local pclose, pnochildren, phaschildren, pcontinue = gparser.close, gparser.noChildren, gparser.hasChildren, gparser.continue

local kindadmonition = gast.newNodeKind("admonition")

return function(m, opts) 
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


local bytes = require 'go.bytes'
local gparser = require 'goldmark.parser'
local gutil = require 'goldmark.util'
local gtext = require 'goldmark.text'
local gsegment = require 'goldmark.text.segment'
local gast = require 'goldmark.ast'
local grenderer = require 'goldmark.renderer'
local hrenderer = require 'goldmark.renderer.html'

local format = string.format
local bstring, bsub = bytes.string, bytes.sub
local isspace, prioritized = gutil.isSpace, gutil.prioritized
local walkcontinue = gast.walkContinue

local kindMention = gast.newNodeKind("mention")

return function(m, opts) 
  local mentionInlineParser = gparser.newInlineParser({
    triggers = "@",
    parse = function(self, parent, block, pc)
      local line, segment = block:peekLine()
      local length = 0
      for i = 1, #line do
        if isspace(line[i]) or i == #line then
          length = i - 1
          break
        end
      end
      block:advance(length)
      if length ~= 0 then
        return gast.newInlineNode({
          kind = kindMention,
          props = {
            name = bstring(bsub(line, 1, length))
          }
        })
      end
      return nil
    end
  })
  
  class = opts.class or "mention"
  local mentionHTMLRenderer = hrenderer.newRenderer({
    registerFuncs = function(self, reg)
      reg:register(kindMention,  function(w, source, n, entering) 
        if entering then
          w:writeString(format("<span class=\"%s\">@%s</span>", class, n:prop("name")))
        end
        return walkcontinue, nil
      end)
    end
  })

  m:parser():addOptions(
    gparser.withInlineParsers(
      prioritized(mentionInlineParser, 999)
    )
  )
  m:renderer():addOptions(
    grenderer.withNodeRenderers(
      prioritized(mentionHTMLRenderer, 999)
    )
  )
end


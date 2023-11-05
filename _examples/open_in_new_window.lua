local bytes = require 'go.bytes'
local gparser = require 'goldmark.parser'
local gutil = require 'goldmark.util'
local gast = require 'goldmark.ast'

local bhasprefix, b = bytes.hasPrefix, bytes.fromString
local  prioritized = gutil.prioritized
local walk, walkcontinue, kindautolink, kindlink = gast.walk, gast.walkContinue, gast.kindAutoLink, gast.kindLink

local bdot = b(".")
local bslash = b("/")
local btarget = b("target")
local bblank = b("_blank")

return function(m, opts) 
  local baseurl = b(opts.base or "")
  local linkTransformer = gparser.newASTTransformer({
    transform = function(self, node, reader, pc)
      walk(node, function(n, entering)
        if not entering then
          return walkcontinue, nil
        end
        local kind = n:kind()
        if kind == kindautolink or kind == kindlink then
          local dest = n.destination
          if not (bhasprefix(dest, bdot) or bhasprefix(dest, bslash) or bhasprefix(dest, baseurl)) then
            n:setAttribute(btarget, bblank)
          end
        end
        return walkcontinue, nil
      end)
    end
  })

  m:parser():addOptions(
    gparser.withASTTransformers(
      prioritized(linkTransformer, 999)
    )
  )
end


return {
  -- Noice has no default route for :! output, so it gets silently discarded.
  -- Show it in a split instead: `q` closes it. Upstream fix: folke/noice.nvim#1098
  {
    "folke/noice.nvim",
    opts = function(_, opts)
      opts.routes = opts.routes or {}
      table.insert(opts.routes, 1, {
        filter = { event = "msg_show", kind = { "shell_cmd", "shell_out", "shell_err", "shell_ret" } },
        view = "messages",
      })
    end,
  },
}

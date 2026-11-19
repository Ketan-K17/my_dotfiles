-- Autocmds are automatically loaded on the VeryLazy event
-- Default autocmds that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/autocmds.lua
--
-- Add any additional autocmds here
-- with `vim.api.nvim_create_autocmd`
--
-- Or remove existing autocmds by their group name (which is prefixed with `lazyvim_` for the defaults)
-- e.g. vim.api.nvim_del_augroup_by_name("lazyvim_wrap_spell")

-- Open mini.files when nvim is started on a directory (e.g. `nvim .`).
-- This file loads on VeryLazy, i.e. after VimEnter, so check argv directly.
local arg = vim.fn.argv(0)
if arg ~= "" and vim.fn.isdirectory(arg) == 1 then
  vim.schedule(function()
    require("mini.files").open(arg, false)
  end)
end

-- No completion popups while writing prose (blink.cmp honors vim.b.completion).
vim.api.nvim_create_autocmd("FileType", {
  group = vim.api.nvim_create_augroup("prose_no_completion", { clear = true }),
  pattern = { "text", "markdown" },
  callback = function(ev)
    vim.b[ev.buf].completion = false
  end,
})

-- Same for scratch buffers with no filetype (plain `nvim`); FileType never fires for those.
vim.api.nvim_create_autocmd("InsertEnter", {
  group = "prose_no_completion",
  callback = function(ev)
    if vim.bo[ev.buf].filetype == "" then
      vim.b[ev.buf].completion = false
    end
  end,
})

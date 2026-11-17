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

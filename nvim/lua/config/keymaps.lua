-- Keymaps are automatically loaded on the VeryLazy event
-- Default keymaps that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/keymaps.lua
-- Add any additional keymaps here

-- AeroSpace owns Alt+h/j/k/l, so drop LazyVim's Alt+j/k "move line" maps
for _, mode in ipairs({ "n", "i", "v" }) do
  pcall(vim.keymap.del, mode, "<A-j>")
  pcall(vim.keymap.del, mode, "<A-k>")
end

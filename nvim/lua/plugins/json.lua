return {
  -- Graph view for JSON: collapse, jump, and edit nodes. `:Videre` or <leader>jv
  {
    "Owen-Dechow/videre.nvim",
    cmd = "Videre",
    keys = {
      { "<leader>jv", "<cmd>Videre<cr>", ft = { "json", "jsonc" }, desc = "JSON graph view (Videre)" },
    },
    opts = {
      box_style = "sharp",
    },
  },
}

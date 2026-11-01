vim.api.nvim_create_user_command("Journal", function() require("sn-journal").open() end, {
  desc = "Open today's diary entry from Standard Notes",
})

vim.api.nvim_create_user_command("Journal", function() require("sn-journal").open() end, {
  desc = "Open today's diary entry from Standard Notes",
})
vim.api.nvim_create_user_command("JournalProtect", function() require("sn-journal").protect() end, {
  desc = "Protect the open diary entry; opening it then asks for the Standard Notes password",
})

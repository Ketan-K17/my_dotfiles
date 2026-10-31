-- JSON helpers: format/minify the buffer with jq, run jq queries.
local M = {}

-- Only replaces the buffer if jq succeeds, so invalid JSON never gets overwritten by an error.
local function jq_buffer(args)
  local buf = vim.api.nvim_get_current_buf()
  local input = table.concat(vim.api.nvim_buf_get_lines(buf, 0, -1, false), "\n")
  local res = vim.system(vim.list_extend({ "jq" }, args), { stdin = input, text = true }):wait()
  if res.code ~= 0 then
    vim.notify("jq: " .. vim.trim(res.stderr or ""), vim.log.levels.ERROR)
    return
  end
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, vim.split(vim.trim(res.stdout), "\n"))
end

function M.setup(buf)
  local function map(lhs, fn, desc)
    vim.keymap.set("n", lhs, fn, { buffer = buf, desc = desc })
  end
  map("<leader>jf", function() jq_buffer({ "." }) end, "Format JSON (jq)")
  map("<leader>jm", function() jq_buffer({ "-c", "." }) end, "Minify JSON (jq)")
  map("<leader>jq", function()
    vim.ui.input({ prompt = "jq filter: " }, function(filter)
      if not filter or filter == "" then
        return
      end
      local out = vim.fn.system({ "jq", filter }, vim.api.nvim_buf_get_lines(buf, 0, -1, false))
      vim.cmd("new")
      vim.bo.buftype, vim.bo.bufhidden, vim.bo.filetype = "nofile", "wipe", "json"
      vim.api.nvim_buf_set_lines(0, 0, -1, false, vim.split(vim.trim(out), "\n"))
    end)
  end, "jq query -> scratch buffer")
  pcall(function()
    require("which-key").add({ { "<leader>j", group = "json", buffer = buf } })
  end)
end

return M

-- :Journal opens today's diary entry from Standard Notes in a buffer.
-- Every :w pushes it in the background, one push at a time. Closing the
-- buffer (or quitting nvim) deletes the local file once Standard Notes has
-- the latest text, and clears registers holding text yanked from it.

local M = {}

local root = vim.fn.fnamemodify(debug.getinfo(1, "S").source:sub(2), ":h:h:h")
local helper = root .. "/bin/sn-journal"
local state = (vim.env.XDG_STATE_HOME or vim.fn.expand("~/.local/state")) .. "/sn-journal"
local file = state .. "/entry.txt"
local meta = state .. "/entry.meta"

local current = nil -- the open entry's session, if any

local function notify(msg, level)
  vim.schedule(function() vim.notify("journal: " .. msg, level or vim.log.levels.WARN) end)
end

local function tag()
  return vim.g.sn_journal_tag or vim.env.SN_JOURNAL_TAG or "musings / nsfw"
end

-- Registers yanks/deletes land in; "+" and "*" because LazyVim yanks to the clipboard.
local registers = vim.split('"0123456789-abcdefghijklmnopqrstuvwxyz+*', "")

local function clear_registers(yanked)
  for _, r in ipairs(registers) do
    local ok, lines = pcall(vim.fn.getreg, r, 1, true)
    if ok and #lines > 0 and yanked[table.concat(lines, "\n")] then pcall(vim.fn.setreg, r, "") end
  end
end

-- Wires push-on-save and cleanup to the entry buffer.
local function attach(buf, e)
  local cmd = { helper, "push", "--tag", e.tag, "--title", e.title, "--file", file, "--created", e.created }
  local running, pending, closed = nil, false, false
  -- Standard Notes may lack what is on disk; a leftover starts out unsynced.
  local needs_push = e.leftover
  local yanked = {}

  local function finished(res)
    running = nil
    if res.code ~= 0 then
      needs_push = true
      notify("push failed, entry kept locally: " .. vim.trim(res.stderr or ""))
    end
  end

  local function push()
    if running then
      pending = true
      return
    end
    pending, needs_push = false, false
    running = vim.system(cmd, { text = true }, function(res)
      if closed then return end
      finished(res)
      if pending then vim.schedule(push) end
    end)
  end

  local function close()
    if closed then return end
    closed = true
    current = nil
    if running then finished(running:wait()) end
    if needs_push or pending then
      pending, needs_push = false, false
      finished(vim.system(cmd, { text = true }):wait())
    end
    clear_registers(yanked)
    if needs_push then
      notify(("%q did not reach Standard Notes; :Journal reopens it"):format(e.title), vim.log.levels.ERROR)
    else
      os.remove(file)
      os.remove(meta)
    end
  end

  local group = vim.api.nvim_create_augroup("sn_journal", { clear = true })
  vim.api.nvim_create_autocmd("BufWritePost", {
    group = group,
    buffer = buf,
    callback = function()
      needs_push = true
      push()
    end,
  })
  vim.api.nvim_create_autocmd("TextYankPost", {
    group = group,
    buffer = buf,
    callback = function() yanked[table.concat(vim.v.event.regcontents, "\n")] = true end,
  })
  vim.api.nvim_create_autocmd({ "BufDelete", "BufWipeout" }, { group = group, buffer = buf, callback = close })
  vim.api.nvim_create_autocmd("VimLeavePre", { group = group, callback = close })

  if e.leftover then
    notify(("opened %q, which never reached Standard Notes. It is pushed on save or close."):format(e.title))
  end
end

local function edit(e)
  -- noswapfile must apply before the buffer loads.
  vim.cmd("noswapfile edit " .. vim.fn.fnameescape(file))
  local buf = vim.api.nvim_get_current_buf()
  vim.bo[buf].swapfile = false
  vim.bo[buf].undofile = false
  current = { buf = buf }
  attach(buf, e)
end

local function read_meta()
  local f = io.open(meta, "r")
  if not f then return nil end
  local title, created, t = f:read("l"), f:read("l"), f:read("l")
  f:close()
  if title and created and t then return { title = title, created = created, tag = t, leftover = true } end
end

function M.open()
  if current and vim.api.nvim_buf_is_valid(current.buf) then
    vim.api.nvim_set_current_buf(current.buf)
    return
  end
  vim.fn.mkdir(state, "p", tonumber("700", 8))

  if vim.uv.fs_stat(file) then
    local e = read_meta()
    if not e then return notify(file .. " exists but its metadata is missing; move it away", vim.log.levels.ERROR) end
    return edit(e)
  end

  local res = vim.system({ helper, "title" }, { text = true }):wait()
  if res.code ~= 0 then return notify("title: " .. vim.trim(res.stderr or ""), vim.log.levels.ERROR) end
  local e = {
    title = vim.trim(res.stdout),
    created = os.date("!%Y-%m-%dT%H:%M:%SZ"),
    tag = tag(),
    leftover = false,
  }

  vim.notify("journal: syncing…")
  vim.system({ helper, "pull", "--tag", e.tag, "--title", e.title, "--out", file }, { text = true }, function(r)
    vim.schedule(function()
      if r.code ~= 0 then
        os.remove(file)
        local why = r.code == 3 and ("more than one %q in %q; merge them in Standard Notes first"):format(e.title, e.tag)
          or vim.trim(r.stderr or "")
        return notify(why, vim.log.levels.ERROR)
      end
      local f = assert(io.open(meta, "w"))
      f:write(e.title, "\n", e.created, "\n", e.tag, "\n")
      f:close()
      edit(e)
    end)
  end)
end

return M

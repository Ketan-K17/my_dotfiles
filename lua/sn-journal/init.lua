-- :Journal opens today's diary entry from Standard Notes in a buffer.
-- Every :w pushes it in the background, one push at a time. Closing the
-- buffer (or quitting nvim) deletes the local file once Standard Notes has
-- the latest text, and clears registers holding text yanked from it.
-- An entry protected in Standard Notes (or with :JournalProtect) opens only
-- after the account password checks out, the same one the app asks for.

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

local function write_meta(e)
  local f = assert(io.open(meta, "w"))
  f:write(e.title, "\n", e.created, "\n", e.tag, "\n", e.protected and "protected" or "open", "\n")
  f:close()
end

-- Prompts for the account password until it checks out, then calls
-- on_ok(password). Gives up after three tries or an empty answer.
local function ask_password(on_ok)
  for _ = 1, 3 do
    local pw = vim.fn.inputsecret("Standard Notes password: ")
    vim.cmd.redraw()
    if pw == "" then break end
    local r = vim.system({ helper, "unlock" }, { text = true, stdin = pw }):wait()
    if r.code == 0 then return on_ok(pw) end
    if r.code ~= 5 then return notify("unlock: " .. vim.trim(r.stderr or ""), vim.log.levels.ERROR) end
    vim.api.nvim_echo({ { "journal: incorrect password", "ErrorMsg" } }, true, {})
  end
  notify("entry is protected; not opened", vim.log.levels.ERROR)
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
-- Returns the function :JournalProtect calls for this entry.
local function attach(buf, e)
  local function cmd()
    local c = { helper, "push", "--tag", e.tag, "--title", e.title, "--file", file, "--created", e.created }
    if e.protected then table.insert(c, "--protect") end
    return c
  end
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
    running = vim.system(cmd(), { text = true }, function(res)
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
      finished(vim.system(cmd(), { text = true }):wait())
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

  return function()
    if e.protected then return notify(("%q is already protected"):format(e.title), vim.log.levels.INFO) end
    e.protected = true
    write_meta(e)
    needs_push = true
    push()
    notify(("protected %q; opening it now asks for your Standard Notes password"):format(e.title), vim.log.levels.INFO)
  end
end

local function edit(e)
  -- noswapfile must apply before the buffer loads.
  vim.cmd("noswapfile edit " .. vim.fn.fnameescape(file))
  local buf = vim.api.nvim_get_current_buf()
  vim.bo[buf].swapfile = false
  vim.bo[buf].undofile = false
  -- Plain writing: no spell squiggles (LazyVim enables spell for text) and no completion popups.
  vim.wo.spell = false
  vim.b[buf].completion = false
  current = { buf = buf }
  current.protect = attach(buf, e)
end

local function read_meta()
  local f = io.open(meta, "r")
  if not f then return nil end
  local title, created, t, prot = f:read("l"), f:read("l"), f:read("l"), f:read("l")
  f:close()
  -- Metadata from before protection support says nothing either way; ask.
  if title and created and t then
    return { title = title, created = created, tag = t, leftover = true, protected = prot == "protected", locked = prot ~= "open" }
  end
end

-- Pulls the entry into file and opens it. A protected entry comes back
-- with exit code 4 and is retried with the password.
local function pull(e, password)
  local cmd = { helper, "pull", "--tag", e.tag, "--title", e.title, "--out", file }
  if password then table.insert(cmd, "--password-stdin") end
  vim.notify("journal: syncing…")
  vim.system(cmd, { text = true, stdin = password }, function(r)
    vim.schedule(function()
      if r.code == 4 then
        return ask_password(function(pw) pull(e, pw) end)
      end
      if r.code ~= 0 then
        os.remove(file)
        local why = r.code == 3 and ("more than one %q in %q; merge them in Standard Notes first"):format(e.title, e.tag)
          or vim.trim(r.stderr or "")
        return notify(why, vim.log.levels.ERROR)
      end
      e.protected = password ~= nil
      write_meta(e)
      edit(e)
    end)
  end)
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
    if e.locked then return ask_password(function() edit(e) end) end
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
  pull(e)
end

function M.protect()
  if not (current and vim.api.nvim_buf_is_valid(current.buf)) then
    return notify("open today's entry with :Journal first", vim.log.levels.ERROR)
  end
  current.protect()
end

return M

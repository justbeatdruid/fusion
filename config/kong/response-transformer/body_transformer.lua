local cjson = require("cjson.safe").new()


local insert = table.insert
local find = string.find
local type = type
local sub = string.sub
local gsub = string.gsub
local match = string.match
local lower = string.lower
local kong = kong


cjson.decode_array_with_array_mt(true)


local noop = function() end


local _M = {}


local function toboolean(value)
  if value == "true" then
    return true
  else
    return false
  end
end


local function cast_value(value, value_type)
  if value_type == "number" then
    return tonumber(value)
  elseif value_type == "boolean" then
    return toboolean(value)
  else
    return value
  end
end


local function read_json_body(body)
  if body then
    return cjson.decode(body)
  end
end


local function append_value(current_value, value)
  local current_value_type = type(current_value)

  if current_value_type  == "string" then
    return {current_value, value }
  end

  if current_value_type  == "table" then
    insert(current_value, value)
    return current_value
  end

  return { value }
end

local function iter(config_array)
  if type(config_array) ~= "table" then
    return noop
  end

  return function(config_array, i)
    i = i + 1

    local current_pair = config_array[i]
    if current_pair == nil then -- n + 1
      return nil
    end

    local current_name, current_value = match(current_pair, "^([^:]+):*(.-)$")
    if current_value == "" then
      current_value = nil
    end

    return i, current_name, current_value
  end, config_array, 0
end


function _M.is_json_body(content_type)
  return content_type and find(lower(content_type), "application/json", nil, true)
end


function RemoveKey(t, name)
    kong.log("body transformer begin remove key", t)
    for key,value in pairs(t) do
        if key == name then
		   kong.log("now can remove key", key)
           t[name] = nil
        end
        if type(value) == "table" then
             kong.log("body transformer this value is a table", key)
             if key == name then
			    kong.log("body transformer now can remove key and value is a table", key)
                t[key] = nil
             else
                RemoveKey(value, name)
             end
        end
    end
end

function ReplaceKey(t, name, v)
    kong.log("body transformer begin replace key", t)
    for key,value in pairs(t) do
        if key == name then
		   kong.log("body transformer begin replace key", key, v)
           t[name] = v
        end
        if type(value) == "table" then
			 kong.log("body transformer replace value is a table", key, value)
             if key == name then
			    kong.log("body transformer begin replace key and value is a table", key)
                t[key] = v
				kong.log("body transformer begin replace key and value is a table", key, t[key])
             else
                ReplaceKey(value, name, v)
             end
        end
    end
end


function _M.transform_json_body(conf, buffered_data)
  kong.log("body transformer begin transform_json_body", buffered_data)
  local json_body = read_json_body(buffered_data)
  if json_body == nil then
    return
  end

  -- remove key:value to body
  kong.log("body transformer remove json ", conf.remove.json)
  for _, name in iter(conf.remove.json) do
    RemoveKey(json_body, name)
  end
  --json_body.data.kongService = nil
   --json_body.data.disStatus = nil
  -- replace key:value to body
  for i, name, value in iter(conf.replace.json) do
    local v = cjson.encode(value)
    if v and sub(v, 1, 1) == [["]] and sub(v, -1, -1) == [["]] then
      v = gsub(sub(v, 2, -2), [[\"]], [["]]) -- To prevent having double encoded quotes
    end

    v = v and gsub(v, [[\/]], [[/]]) -- To prevent having double encoded slashes

    if conf.replace.json_types then
      local v_type = conf.replace.json_types[i]
      v = cast_value(v, v_type)
    end

    --if json_body[name] and v ~= nil then
    --  json_body[name] = v
    --end
    ReplaceKey(json_body, name, v)
  end

  -- add new key:value to body
  for i, name, value in iter(conf.add.json) do
    local v = cjson.encode(value)
    if v and sub(v, 1, 1) == [["]] and sub(v, -1, -1) == [["]] then
      v = gsub(sub(v, 2, -2), [[\"]], [["]]) -- To prevent having double encoded quotes
    end

    v = v and gsub(v, [[\/]], [[/]]) -- To prevent having double encoded slashes

    if conf.add.json_types then
      local v_type = conf.add.json_types[i]
      v = cast_value(v, v_type)
    end
    --添加当前只支持添加新key值 
    if not json_body[name] and v ~= nil then
      json_body[name] = v
    end
  end

  -- append new key:value or value to existing key
  for i, name, value in iter(conf.append.json) do
    local v = cjson.encode(value)
    if v and sub(v, 1, 1) == [["]] and sub(v, -1, -1) == [["]] then
      v = gsub(sub(v, 2, -2), [[\"]], [["]]) -- To prevent having double encoded quotes
    end

    v = v and gsub(v, [[\/]], [[/]]) -- To prevent having double encoded slashes

    if conf.append.json_types then
      local v_type = conf.append.json_types[i]
      v = cast_value(v, v_type)
    end

    if v ~= nil then
      json_body[name] = append_value(json_body[name],v)
    end
  end

  return cjson.encode(json_body)
end


return _M


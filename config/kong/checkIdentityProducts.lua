local http = require "socket.http"
local ltn12 = require "ltn12"
local cjson = require "cjson.safe"
local response_body = {}
local token = kong.request.get_header("token")   
kong.log("Get request token: ", token)
local isManager = kong.request.get_header("isManager")   
kong.log("Get isManager: ", isManager)
local path = ngx.var.upstream_uri
kong.log("Get request path: ", path)
local orgMethod = kong.request.get_method()
kong.log("Get request method: ", orgMethod)
local tenantId = kong.request.get_header("managerGroupUuid")
kong.log("Get request group uuid is:", tenantId)
local query = string.format("?path=%s&method=%s&isManager=%s", path, orgMethod, isManager)
local getUrl = "http://fusion-auth:8092/fusion-auth/sys/support/checkIdentity".. query
kong.log("Check identity url is: ", getUrl)
if token == nil then
    kong.log("Check identity token is nil and set tenantId and userId is null")
    kong.service.request.add_header("userId",  "")
    kong.service.request.add_header("tenantId", "")
    return
end

local _, _, response_headers = http.request{
    url = getUrl,
    method = "GET",
	headers = {
      ["token"] = token
    },
    sink = ltn12.sink.table(response_body),
  }
  
if type(response_body) ~= "table" then
    kong.response.exit(500, { message = "Check identity response_body is not table", code = 500 })
end 
local resp = table.concat(response_body)
local decoded, err = cjson.decode(resp)
if err then
    kong.log("Failed to decode check identity response body: ", err)
    kong.response.exit(500, { message = err, code = 500 })
end

local code = decoded.code
kong.log("Check identity return code is:", code)
local msg = decoded.msg
kong.log("Check identity return msg is:", msg)
if code~= 0 then
    kong.log("Check identity return code is: ", code)
    kong.response.exit(500, { message = decoded.msg, code = code })
end

local userId = decoded.userId
kong.log("Check identity return userId is:", userId)
kong.log("Check identity return tenantId is:", tenantId)
kong.service.request.add_header("userId",  userId)
if tenantId == nil then
    kong.log("Check identity tenantId is nil and no need set tenantId")
else
    kong.service.request.add_header("tenantId", tenantId)
end
kong.log("=======Check identity end=======") 
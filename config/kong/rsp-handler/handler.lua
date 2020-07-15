local body_transformer = require "kong.plugins.rsp-handler.body_transformer"
local header_transformer = require "kong.plugins.rsp-handler.header_transformer"
local http = require "socket.http"
local ltn12 = require "ltn12"
local cjson = require "cjson.safe"


local is_body_transform_set = header_transformer.is_body_transform_set
local is_json_body = header_transformer.is_json_body
local concat = table.concat
local kong = kong
local ngx = ngx




local ResponseTransformerHandler = {}


function ResponseTransformerHandler:header_filter(conf)
  header_transformer.transform_headers(conf, kong.response.get_headers())
end


function ResponseTransformerHandler:body_filter(conf)
    local ctx = ngx.ctx
    local chunk, eof = ngx.arg[1], ngx.arg[2]
    ctx.rt_body_chunks = ctx.rt_body_chunks or {}
    ctx.rt_body_chunk_number = ctx.rt_body_chunk_number or 1
    local response_body = {}
    if eof then
        local chunks = concat(ctx.rt_body_chunks)
        kong.log("begin rsp handler chunks ", chunks)
        local request_body = chunks
        kong.log("begin rsp handler url ", conf.function_url)
		--TODO后续可以优化先处理body再下发请求
        --local body = body_transformer.transform_json_body(conf, chunks)
        local res, code, response_headers = http.request{
        url = conf.function_url,
        method = "POST",
        headers = {
        ["Content-Type"] = "application/json";
        ["Content-Length"] = #request_body;
        },
        source = ltn12.source.string(request_body),  
        sink = ltn12.sink.table(response_body),
        }
        if type(response_body) ~= "table" then
            kong.log.err("end rsp handler rsp body is nil", res, code)
            ngx.arg[1] = nil
            return          
        end
        local resp = table.concat(response_body)
        kong.log("end rsp handler return rsp", resp) 
        ngx.arg[1] = resp or chunks 
        kong.log("end rsp handler ngx.arg[1]", ngx.arg[1])
    else
      ctx.rt_body_chunks[ctx.rt_body_chunk_number] = chunk
      ctx.rt_body_chunk_number = ctx.rt_body_chunk_number + 1
      ngx.arg[1] = nil
    end
end


ResponseTransformerHandler.PRIORITY = 799
ResponseTransformerHandler.VERSION = "2.0.0"


return ResponseTransformerHandler


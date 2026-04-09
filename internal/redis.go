package internal

var luaScript = `
local key = KEYS[1]
local window_start = tonumber(ARGV[2])
local limit = tonumber(ARGV[1])
local now = tonumber(ARGV[3])
local window = tonumber(ARGV[4])
local reqId = ARGV[5]

redis.call("ZREMRANGEBYSCORE", key, 0, window_start)

local count = redis.call("ZCARD", key)

if count < limit then
	redis.call("ZADD", key, now, reqId)
	redis.call("EXPIRE", key, window)
	return 1
else
	return 0
end
`

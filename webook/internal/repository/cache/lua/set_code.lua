-- 发送到的 key，也就是 code:业务：手机好吗
local key = KEYS[1]
-- 使用次数，也就是验证次数
local cntKey = key..":cnt"
-- 你准备的存储的验证吗
local val = ARGV[1]
local ttl = tonumber(redis.call("ttl", key))

-- 发送到的 key，也就是 code:业务:手机号码
local key = KEYS[1]
-- 可以使用次数，也就是剩余的验证次数
local cntKey = key..":cnt"
-- 你准备存储的验证吗
local val = ARGV[1]

-- ttl:
-- -2 key does not exist.
-- -1 key exists but has no associated expire.
local ttl = tonumber(redis.call("ttl", key))
if ttl == -1 then
    -- key 存在，但没有过期时间
    return -2
elseif ttl == -2 or ttl < 540 then
    -- 可以发验证码
    redis.call("set", key, val)
    -- 600 秒
    redis.call("expire", key, 600)
    redis.call("set", cntKey, 3)
    redis.call("expire", cntKey, 600)
    return 0
else
    -- 发送太频繁
    return -1
end
local key = KEYS[1]
-- 使用次数，也就是验证次数
local cntKey = key..":cnt"
-- 用户输入的验证码
local expectedCode = ARGV[1]

local cnt = tonumber(redis.call("get", cntKey))
local code = redis.call("get", key)

-- 验证次数已经耗尽了
if cnt == nil or cnt <= 0 then
    return -1
end
-- 验证码相等
-- 不能删除验证码，因为如果你删除了就有可能有人跟你过不去
-- 立刻再次发送验证码
if code == expectedCode then
    redis.call("set", cntKey, 0)
    return 0
else
    redis.call("decr", cntKey)
    -- 不相等，用户输错了
    return -2
end
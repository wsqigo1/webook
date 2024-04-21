//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(localhost:13306)/webook",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
}

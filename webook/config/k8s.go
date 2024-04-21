//go:build k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(webook-record-mysql:13306)/webook",
	},
	Redis: RedisConfig{
		Addr: "webook-record-mysql:6379",
	},
}

module customs

go 1.23

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.2.0
	gorm.io/driver/mysql v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/text v0.20.0 // indirect
)

replace (
	github.com/redis/go-redis/v9 => github.com/go-redis/redis/v8 v8.11.5
	golang.org/x/sys => golang.org/x/sys v0.25.0
	golang.org/x/time => golang.org/x/time v0.7.0
)

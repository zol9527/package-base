module package-base

go 1.24.5

require (
	github.com/duke-git/lancet/v2 v2.3.7
	github.com/go-playground/validator/v10 v10.27.0
	github.com/gofiber/fiber/v2 v2.52.8
	github.com/golang-jwt/jwt/v5 v5.2.3
	github.com/google/uuid v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.5.0
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.40.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	// 删除 github.com/fasthttp/session/v2 间接依赖（由 Logto/session 移除引起）
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

require (
	github.com/aliyun/aliyun-oss-go-sdk v3.0.2+incompatible
	github.com/go-redis/redis/v8 v8.11.5
	golang.org/x/text v0.27.0 // indirect
	gorm.io/driver/mysql v1.6.0
	gorm.io/gorm v1.30.0
)

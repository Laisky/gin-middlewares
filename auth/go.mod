module github.com/Laisky/gin-middlewares/auth

go 1.16

require (
	github.com/Laisky/gin-middlewares/library v0.0.0
	github.com/Laisky/go-utils v1.15.0
	github.com/Laisky/zap v1.12.3-0.20210804015521-853b5a8ec429
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible
	github.com/gin-gonic/gin v1.7.4
	github.com/pkg/errors v0.9.1
)

replace github.com/Laisky/gin-middlewares/library v0.0.0 => ../library

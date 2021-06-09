package webhandlers

import (
    "gorm.io/gorm"
    "github.com/appleboy/gin-jwt/v2"
    "github.com/gin-gonic/gin"
    
    "iot_meter/models"

    "os"
    "log"
    "time"
    "crypto/tls"
    "net/http"
)

const JwtIdKey = "iot_account_id"

type MainServer struct {
    DB     *gorm.DB
    Router *gin.Engine

    _authMiddleware *jwt.GinJWTMiddleware
}

func (server *MainServer) setupJwtMiddleware() (error) {

    var err error
	server._authMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "Fake IOT",
		Key:         []byte(os.Getenv("JWT_SECRET")),
		Timeout:     time.Hour * 2,
		MaxRefresh:  time.Hour * 6,
		IdentityKey: JwtIdKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*models.Account); ok {
				return jwt.MapClaims{
					JwtIdKey: v.ID,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &models.Account{
				ID: claims[JwtIdKey].(string),
			}
		},

		Authenticator: LoginHandler,

		Authorizator: func(data interface{}, c *gin.Context) bool {
			// TBI, defaults true
			return true
		},

		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},

        TokenLookup: "header: Authorization, query: token, cookie: jwt",
        SendCookie:       true,
        SecureCookie:     true,
        CookieHTTPOnly:   true,  // JS can't modify
	    CookieDomain:     "localhost:8080",
	})

	if err != nil {
		log.Println("JWT Error:" + err.Error())
		return err
	}

	// When you use jwt.New(), the function is already automatically called for checking,
	// which means you don't need to call it again.
	err = server._authMiddleware.MiddlewareInit()

	if err != nil {
		log.Println("authMiddleware.MiddlewareInit() Error:" + err.Error())
		return err
	}

    return nil
}

func (server *MainServer) setupDbMiddleware()  {
    server.Router.Use(func (c *gin.Context) {
        c.Set("DB", server.DB)
        c.Next()
    })
}

func (server *MainServer) setupMiddlewares() {
    server.setupDbMiddleware()
    server.setupJwtMiddleware()
}

func (server *MainServer) setupRoutes() {
    
    // Login endpoint to obtain JWT
    server.Router.POST("/login", server._authMiddleware.LoginHandler)

    // 404
    server.Router.NoRoute(server._authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

    // IoT API Group
    v1 := server.Router.Group("api/v1")
    v1.Use(server._authMiddleware.MiddlewareFunc())
    v1.POST("/metrics", Metrics)
    
}

func (server *MainServer) listenTLS() (error) {

	// Load SSL cert and Key from IOT client
	cer, err := tls.LoadX509KeyPair(
        os.Getenv("SSL_CERT_PATH"),
        os.Getenv("SSL_KEY_PATH"),
    )

	if err != nil {
		return err
	}

	// Harden HTTPS configuration
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS13,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		},
		Certificates: []tls.Certificate{cer},
    }
    
	srv := &http.Server{
		Addr:      ":" + os.Getenv("SERVER_PORT"),
		Handler:   server.Router,
		TLSConfig: cfg,
	}

	log.Println("Started TLS server")
	return srv.ListenAndServeTLS("", "")
}

func (server *MainServer) Initialize(db *gorm.DB) (error) {
    server.Router = gin.Default()
    server.DB = db
    
    server.setupMiddlewares()
    server.setupRoutes()

    return nil
}

func (server *MainServer) Run() (error) {
    err := server.listenTLS()
    if err != nil {
        return err
    }

    return nil
}
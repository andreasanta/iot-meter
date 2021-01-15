package main

import (
	"iot_meter/utils"
	"iot_meter/db"
	myhttp "iot_meter/http"
	"iot_meter/models"

	"time"
	"runtime"
	"os"
	"crypto/tls"
    "log"
    "net/http"

	"github.com/gin-gonic/gin"
	"github.com/appleboy/gin-jwt/v2"
)

const JWT_ID_KEY = "iot_account_id"

type login struct {
	Email string `form:"email" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func setupJwtMiddleware() (*jwt.GinJWTMiddleware) {
	
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "Fake IOT",
		Key:         []byte(os.Getenv("JWT_SECRET")),
		Timeout:     time.Hour * 2,
		MaxRefresh:  time.Hour * 6,
		IdentityKey: JWT_ID_KEY,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*models.Account); ok {
				return jwt.MapClaims{
					JWT_ID_KEY: v.ID,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &models.Account{
				ID: claims[JWT_ID_KEY].(string),
			}
		},

		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals login
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			/*
			email := loginVals.Email
			password := loginVals.Password

			
			 TBI fetch user from database
			
			if (userID == "admin" && password == "admin") || (userID == "test" && password == "test") {
				return &User{
					UserName:  userID,
					LastName:  "Bo-Yi",
					FirstName: "Wu",
				}, nil
			}*/

			return nil, jwt.ErrFailedAuthentication
		},
		
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
		
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	// When you use jwt.New(), the function is already automatically called for checking,
	// which means you don't need to call it again.
	errInit := authMiddleware.MiddlewareInit()

	if errInit != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
	}

	return authMiddleware

}

func setupServer() {

	r := gin.Default()

	// Maximize parallellism
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Setup JWT authentication
	jwtMiddleware := setupJwtMiddleware()

	v1 := r.Group("api/v1")
	v1.Use(jwtMiddleware.MiddlewareFunc())
	{
		v1.GET("/metrics", myhttp.Metrics)
	}

	// Load SSL cert and Key from IOT client
	cer, err := tls.LoadX509KeyPair(os.Getenv("SSL_CERT_PATH"), os.Getenv("SSL_KEY_PATH"))
    if err != nil {
        log.Fatal(err)
	}

	// Harden HTTPS configuration
    cfg := &tls.Config{
        MinVersion:               tls.VersionTLS12,
        CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
        PreferServerCipherSuites: true,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
            tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
		Certificates: []tls.Certificate{cer},
    }
    srv := &http.Server{
        Addr:         ":" + os.Getenv("SERVER_PORT"),
        Handler:      r,
        TLSConfig:    cfg,
        TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	
	log.Println("Started TLS server")
	log.Fatal(srv.ListenAndServeTLS("", ""))	
}

func main() {

	// Load env config variables
	utils.LoadConfig()

	// Load accounts database
	db.LoadDatabase()

	// Run http server
	setupServer()

}

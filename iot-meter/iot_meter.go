package main

import (
	"iot_meter/db"
	"iot_meter/models"
	"iot_meter/utils"
	"iot_meter/webhandlers"

	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

const JwtIdKey = "iot_account_id"

type login struct {
	Email    string `form:"email" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func setupJwtMiddleware() (*jwt.GinJWTMiddleware, error) {

	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
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
		log.Println("JWT Error:" + err.Error())
		return nil, err
	}

	// When you use jwt.New(), the function is already automatically called for checking,
	// which means you don't need to call it again.
	errInit := authMiddleware.MiddlewareInit()

	if errInit != nil {
		log.Println("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
		return nil, errInit
	}

	return authMiddleware, nil

}

func setupServer() {

	r := gin.Default()

	// Setup JWT authentication
	jwtMiddleware, err := setupJwtMiddleware()
	if err != nil {
		log.Fatal(err)
	}

	v1 := r.Group("api/v1")
	v1.Use(jwtMiddleware.MiddlewareFunc())
	{
		v1.GET("/metrics", webhandlers.Metrics)
	}

	// Load SSL cert and Key from IOT client
	cer, err := tls.LoadX509KeyPair(os.Getenv("SSL_CERT_PATH"), os.Getenv("SSL_KEY_PATH"))
	if err != nil {
		log.Fatal(err)
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
		Handler:   r,
		TLSConfig: cfg,
	}

	log.Println("Started TLS server")
	log.Fatal(srv.ListenAndServeTLS("", ""))
}

func main() {

	// Load env config variables
	err := utils.LoadConfig()
	if err != nil {
		log.Fatal("Unable to load environment configuration")
	}

	// Load accounts database
	_, err = db.LoadDatabase()
	if err != nil {
		log.Fatal("Unable to open or migrate database")
	}

	// Run http server
	setupServer()

}

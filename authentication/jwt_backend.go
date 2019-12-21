package authentication

import (
	"crypto/rsa"
)

type JWTAuthenticationBackend struct {
	privateKey *rsa.PrivateKey
	PublicKey *rsa.PublicKey
}

const (
	tokenDuration = 72
	expireOffset = 3600
)

var authBackendInstance *JWTAuthenticationBackend = nil

func InitJWTAuthenticationBackend() *JWTAuthenticationBackend{
	if authBackendInstance == nil{
		authBackendInstance = &JWTAuthenticationBackend{
			privateKey: 
		}
	}
}

func getPrivateKey() *rsa.PrivateKey{
	privateKeyFile, err := os.Open(settings.Get().)
}
package v0

import (
	"crypto/rsa"
	"github.com/golang-jwt/jwt/v4"
	"os"
	"time"
)

type UserClaims struct {
	*jwt.RegisteredClaims

	ID int64
}

var signKey *rsa.PrivateKey

func init() {
	var err error

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(os.Getenv("TOKEN_PRIVATE_KEY")))

	if err != nil {
		panic("Unable to read RSA private key")
	}
}

func createNewToken(id int64) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)

	token.Claims = &UserClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{
				Time: time.Now().Add(time.Minute * 5),
			},
		},
		ID: id,
	}

	return token.SignedString(signKey)
}

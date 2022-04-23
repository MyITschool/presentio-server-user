package v0

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"os"
	"time"
)

type UserClaims struct {
	*jwt.StandardClaims

	ID int64
}

var signKey = os.Getenv("TOKEN_PRIVATE_KEY")

func createNewToken(id int64) (string, error) {
	fmt.Println(signKey)
	token := jwt.New(jwt.GetSigningMethod("RS256"))

	token.Claims = &UserClaims{
		&jwt.StandardClaims{

			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
		},
		id,
	}

	return token.SignedString(signKey)
}

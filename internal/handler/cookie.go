package handler

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const COOKIE_NAME = "user_id"
const TOKEN_EXP = time.Hour * 3
const SECRET_KEY = "supersecretkey"

func BuildJWTString(user_id int, secret_key string, exp_time time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp_time)),
		},
		UserID: user_id,
	})
	JWT, err := token.SignedString([]byte(secret_key))
	if err != nil {
		return "", err
	}
	return JWT, nil
}

func GetUserID(tokenString string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SECRET_KEY), nil
		})
	if err != nil {
		return -1
	}
	if !token.Valid {
		// fmt.Println("Token is not valid")
		return -1
	}
	// fmt.Println("Token os valid")
	return claims.UserID
}

func generateUserID() int {
	var b [8]byte
	if _, err := rand.Read(b[:]); err == nil {
		n := binary.LittleEndian.Uint64(b[:])
		id := int(n & 0x7fffffff) // ensure positive 31-bit value
		if id == 0 {
			id = 1
		}
		return id
	}
	return 1
}

func CookieMiddleware(wrapped http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie(COOKIE_NAME)
		// llll := GetUserID(cookie.Value)
		// log.Println("Cooooooookie: %s|| %s", err, llll)
		if err != nil || GetUserID(cookie.Value) <= 0 {
			user_id := generateUserID()
			// fmt.Printf("################# user_id =  %d\n", user_id)
			cookie_text, err := BuildJWTString(user_id, SECRET_KEY, TOKEN_EXP)
			if err == nil {
				http.SetCookie(res, &http.Cookie{
					Name:  COOKIE_NAME,
					Value: cookie_text,
				})
			} else {
				log.Println("Error generating JWT token: ", err)
			}

		}
		wrapped(res, req)
	}
}

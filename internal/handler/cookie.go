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

const cookieName = "userId"
const tokenExp = time.Hour * 3
const secretKey = "supersecretkey"

func BuildJWTString(userId int, secretKey string, expTime time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expTime)),
		},
		UserID: userId,
	})
	JWT, err := token.SignedString([]byte(secretKey))
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
			return []byte(secretKey), nil
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
		cookie, err := req.Cookie(cookieName)
		// llll := GetUserID(cookie.Value)
		// log.Println("Cooooooookie: %s|| %s", err, llll)
		if err != nil || GetUserID(cookie.Value) <= 0 {
			userId := generateUserID()
			// fmt.Printf("################# user_id =  %d\n", user_id)
			cookieText, err := BuildJWTString(userId, secretKey, tokenExp)
			if err == nil {
				http.SetCookie(res, &http.Cookie{
					Name:  cookieName,
					Value: cookieText,
				})
			} else {
				log.Println("Error generating JWT token: ", err)
			}

		}
		wrapped(res, req)
	}
}

package handler

import (
	"context"
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

type contextKey string

const ctxUserID contextKey = "userID"

func BuildJWTString(userID int, secretKey string, expTime time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expTime)),
		},
		UserID: userID,
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
		return -1
	}
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
		// Выдавать пользователю симметрично подписанную куку,
		// содержащую уникальный идентификатор пользователя,
		// если такой куки не существует или она не проходит проверку подлинности.
		var userID int
		cookie, err := req.Cookie(cookieName)
		if err != nil || GetUserID(cookie.Value) <= 0 {
			userID = generateUserID()
			cookieText, err := BuildJWTString(userID, secretKey, tokenExp)
			if err == nil {
				http.SetCookie(res, &http.Cookie{
					Name:  cookieName,
					Value: cookieText,
				})
			} else {
				log.Println("Error generating JWT token: ", err)
			}
		} else {
			userID = GetUserID(cookie.Value)
		}
		ctx := context.WithValue(req.Context(), ctxUserID, userID)
		wrapped(res, req.WithContext(ctx))
	}
}

func UserIDFromRequest(req *http.Request) (int, bool) {
	// Parse context first, if middleware was used
	ctxValue := req.Context().Value(ctxUserID) 
	if ctxValue != nil {
		userID, ok := ctxValue.(int)
		if ok && userID > 0 {
			return userID, true
		}
	}
	// Parse cookie directly if middleware not used
	cookie, err := req.Cookie(cookieName)
	if err == nil {
		userID := GetUserID(cookie.Value)
		if userID > 0 {
			return userID, true
		}
	}
	return 0, false
}

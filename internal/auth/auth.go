package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.xubinbest.com/go-game-server/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type UserClaims struct {
	UserID uint64 `json:"userId"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int64, role, email string, secretKey string, tokenExpire time.Duration) (string, time.Time, error) {
	utils.Info("GenerateToken", zap.String("secretKey", secretKey))
	expiresAt := time.Now().Add(tokenExpire)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &UserClaims{
		UserID: uint64(userID),
		Role:   role,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	return tokenString, expiresAt, err
}

func ParseToken(tokenStr string, secretKey string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		utils.Error("Error parsing token", zap.Error(err))
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

func VerifyPassword(inputPassword, storedHash, salt string) bool {
	hashedInput := HashPassword(inputPassword, salt)
	return hashedInput == storedHash
}

func GenerateSalt() string {
	const saltLength = 16
	b := make([]byte, saltLength)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func HashPassword(password, salt string) string {
	h := sha256.New()
	h.Write([]byte(password + salt))
	return hex.EncodeToString(h.Sum(nil))
}

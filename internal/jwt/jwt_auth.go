package jwt

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"io"
	"strconv"
	"time"
	"wh-hard01.kol.wb.ru/wh-tech/wh-tech-back/wh_tech_discord_api/internal/models"
	"wh-hard01.kol.wb.ru/wh_core/gocore_service_configs/configs"
)

type JwtAuth struct {
	jwt         models.Jwt
	cipherBlock cipher.Block
}

func NewJwtAuth() *JwtAuth {
	return &JwtAuth{}
}

func (s *JwtAuth) Configure(_ context.Context, config configs.Config) {
	err := jsoniter.Unmarshal(config.GetByServiceKeyRequired("jwt_auth"), &s.jwt)
	if err != nil {
		logrus.Panicf("Error while parsing jwt_auth configs - %s", err.Error())
	}
	s.cipherBlock, err = aes.NewCipher([]byte(s.jwt.Salt))
	if err != nil {
		logrus.Panicf("aes newCipher error - %v", err)
	}
}

type token struct {
	jwt.RegisteredClaims
}

func (s *JwtAuth) GenerateToken(employeeID int64) (string, error) {
	hashEmployeeID, err := s.encrypt(strconv.Itoa(int(employeeID))) // дополнительно шифруем полученные данные
	if err != nil {
		return "", fmt.Errorf("[GenerateToken] encrypt employeeID error: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &token{
		RegisteredClaims: jwt.RegisteredClaims{
			ID: hashEmployeeID,
		},
	})

	tokenString, err := token.SignedString([]byte(s.jwt.Salt))
	if err != nil {
		return "", fmt.Errorf("[GenerateToken] token.SignedString error: %w", err)
	}

	return tokenString, nil
}

func (s *JwtAuth) ParseToken(tokenString string) (string, error) {
	jwtToken, err := jwt.ParseWithClaims(tokenString, &token{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwt.Salt), nil
	}, jwt.WithLeeway(time.Second))

	if err != nil {
		return "", fmt.Errorf("[ParseToken] jwt.Parse error: %w", err)
	}

	if tokenInfo, ok := jwtToken.Claims.(*token); ok {
		decryptEmployeeID, err := s.decrypt(tokenInfo.ID) // расшифровываем полученные данные
		if err != nil {
			return "", fmt.Errorf("[ParseToken] decrypt employeeID error: %w", err)
		}
		if _, err = strconv.Atoi(decryptEmployeeID); err != nil { // доп. проверки при дэшифровании токена
			return "", fmt.Errorf("[ParseToken] strconv.Atoi error: %w", err)
		}

		return decryptEmployeeID, nil
	}
	return "", errors.New(("[ParseToken] jwt.Parse: token is incorrect"))
}

// Encrypt зашифровать строку.
func (s *JwtAuth) encrypt(payload string) (string, error) {
	text := []byte(payload)

	cipherText := make([]byte, aes.BlockSize+len(text))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("io.ReadFull: %w", err)
	}

	stream := cipher.NewCFBEncrypter(s.cipherBlock, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], text)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt расшифровать строку.
func (s *JwtAuth) decrypt(payload string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("DecodeString error: %w", err)
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("ciphertext not correct")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(s.cipherBlock, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), err
}

package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"
	
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	TOKEN_PUBLIC_KEY, _ := hex.DecodeString(os.Getenv("TOKEN_PUBLIC_KEY"))
	TOKEN_PRIVATE_KEY, _ := hex.DecodeString(os.Getenv("TOKEN_PRIVATE_KEY"))
	// privateKey, publicKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	// publicKey:=privateKey.PublicKey()
	fmt.Println(hex.EncodeToString(TOKEN_PRIVATE_KEY), hex.EncodeToString(TOKEN_PUBLIC_KEY))
	fmt.Println(jwt.GetAlgorithms())
	// userNameHashed,_:=bcrypt.GenerateFromPassword([]byte("Keshav Goel"),bcrypt.DefaultCost)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &jwt.MapClaims{
		"name": "Keshav Goel",
		"exp":  time.Now().Add(time.Second * 5).Unix(),
	})
	tokenString, err := token.SignedString(ed25519.PrivateKey(TOKEN_PRIVATE_KEY))
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	fmt.Println()
	fmt.Println(tokenString)
	fmt.Println()
	claims := jwt.MapClaims{}
	token, err = jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		// token.Method.Verify()
		return ed25519.PublicKey(TOKEN_PUBLIC_KEY), nil
	})
	fmt.Println(claims)
	if err != nil {
		if err == jwt.ErrEd25519Verification {
			panic("Verification failed")
		} else if err == jwt.ErrTokenInvalidClaims {
			token = jwt.NewWithClaims(jwt.SigningMethodEdDSA, &jwt.MapClaims{
				"name": "Keshav Goel",
				"exp":  time.Now().Add(time.Millisecond * 5).Unix(),
			})

			log.Fatalf("err is %v", err)
		} else {
			panic(err)
		}

	}
	fmt.Println(claims, err, token.Valid)
}

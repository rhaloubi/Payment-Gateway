package middleware

import "fmt"

func AuthMiddleware() string {
	r, err := fmt.Println("AuthMiddleware")
	if err != nil {
		fmt.Println(err)
	}
	return r
}

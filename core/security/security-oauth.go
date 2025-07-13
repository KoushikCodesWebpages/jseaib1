// // auth/google.go
package security

// import (

// 	"golang.org/x/oauth2"
// 	"golang.org/x/oauth2/google"
// )

// var (
// 	GoogleOAuthConfig = &oauth2.Config{
// 		ClientID:     "YOUR_CLIENT_ID",
// 		ClientSecret: "YOUR_CLIENT_SECRET",
// 		RedirectURL:  "http://localhost:3000/auth/google/callback",
// 		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
// 		Endpoint:     google.Endpoint,
// 	}
// )

// // UserInfo holds data from Google's userinfo API
// type UserInfo struct {
// 	Email         string `json:"email"`
// 	VerifiedEmail bool   `json:"verified_email"`
// 	Name          string `json:"name"`
// 	Picture       string `json:"picture"`
// }

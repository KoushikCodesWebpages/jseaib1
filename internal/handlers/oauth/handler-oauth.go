package oauth

// import (
//     "context"
//     "crypto/rand"
//     "encoding/base64"
//     "encoding/json"
//     "net/http"

//     "RAAS/core/config"
//     "RAAS/core/security"
//     "go.mongodb.org/mongo-driver/mongo"

//     "github.com/gin-contrib/sessions"
//     "github.com/gin-gonic/gin"
//     "golang.org/x/oauth2"
//     "golang.org/x/oauth2/google"
// )

// const stateCookie = "google_oauth_state"

// func newGoogleOAuthConfig() *oauth2.Config {
//     return &oauth2.Config{
//         RedirectURL:  config.Cfg.Cloud.GoogleRedirectURL,
//         ClientID:     config.Cfg.Cloud.GoogleClientId,
//         ClientSecret: config.Cfg.Cloud.GoogleClientSecret,
//         Scopes: []string{
//             "openid", "email", "profile",
//             "https://www.googleapis.com/auth/user.phonenumbers.read",
//         },
//         Endpoint: google.Endpoint,
//     }
// }

// func GoogleLoginHandler(c *gin.Context) {
//     b := make([]byte, 16)
//     rand.Read(b)
//     state := base64.URLEncoding.EncodeToString(b)
//     sessions.Default(c).Set(stateCookie, state)
//     sessions.Default(c).Save()
//     redirectURL := newGoogleOAuthConfig().AuthCodeURL(state, oauth2.AccessTypeOffline)
//     c.Redirect(http.StatusTemporaryRedirect, redirectURL)
// }

// func GoogleCallbackHandler(c *gin.Context) {
//     sess := sessions.Default(c)
//     if sess.Get(stateCookie) != c.Query("state") {
//         c.String(http.StatusBadRequest, "Invalid OAuth state")
//         return
//     }

//     conf := newGoogleOAuthConfig()
//     token, err := conf.Exchange(context.Background(), c.Query("code"))
//     if err != nil {
//         c.String(http.StatusInternalServerError, "Token exchange failed: "+err.Error())
//         return
//     }

//     client := conf.Client(context.Background(), token)
//     // Basic userinfo
//     userInfoResp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
//     if err != nil {
//         c.String(http.StatusInternalServerError, "Userinfo fetch failed")
//         return
//     }
//     defer userInfoResp.Body.Close()

//     var gInfo struct {
//         Email string `json:"email"`
//         Name  string `json:"name"`
//         Picture string `json:"picture"`
//     }
//     json.NewDecoder(userInfoResp.Body).Decode(&gInfo)

//     // Fetch phone via People API
//     peopleResp, err := client.Get("https://people.googleapis.com/v1/people/me?personFields=phoneNumbers")
//     if err != nil {
//         c.String(http.StatusInternalServerError, "People API fetch failed")
//         return
//     }
//     defer peopleResp.Body.Close()

//     var peopleData struct {
//         PhoneNumbers []struct {
//             Value string `json:"value"`
//             Type  string `json:"type"`
//         } `json:"phoneNumbers"`
//     }
//     json.NewDecoder(peopleResp.Body).Decode(&peopleData)

//     phone := ""
//     if len(peopleData.PhoneNumbers) > 0 {
//         phone = peopleData.PhoneNumbers[0].Value
//     }

//     db := c.MustGet("db").(*mongo.Database)
//     ur := NewGoogleUserRepo(db)
//     user, err := ur.GetOrCreateFromGoogle(gInfo.Email, gInfo.Name, gInfo.Picture, phone)
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//         return
//     }

//     jwtToken, err := security.GenerateJWT(user.AuthUserID, user.Email, user.Role)
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "token_generation_failed"})
//         return
//     }

//     c.JSON(http.StatusOK, gin.H{"token": jwtToken})
// }

package oauth

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "net/http"

    "RAAS/core/config"
    "github.com/gin-contrib/sessions"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/mongo"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"

    "RAAS/core/security"
)

func newGoogleOAuthConfig() *oauth2.Config {
    return &oauth2.Config{
        RedirectURL:  config.Cfg.Cloud.GoogleRedirectURL,
        ClientID:     config.Cfg.Cloud.GoogleClientId,
        ClientSecret: config.Cfg.Cloud.GoogleClientSecret,
        Scopes:       []string{"openid", "email", "profile"},
        Endpoint:     google.Endpoint,
    }
}


const stateCookie = "google_oauth_state"

func GoogleLoginHandler(c *gin.Context) {
    b := make([]byte, 16)
    rand.Read(b)
    state := base64.URLEncoding.EncodeToString(b)
    sessions.Default(c).Set(stateCookie, state)
    sessions.Default(c).Save()
    conf := newGoogleOAuthConfig()
    redirectURL := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
    c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func GoogleCallbackHandler(c *gin.Context) {
    sess := sessions.Default(c)
    savedState := sess.Get(stateCookie)
    if savedState == nil || c.Query("state") != savedState {
        c.String(http.StatusBadRequest, "Invalid OAuth state")
        return
    }
    conf := newGoogleOAuthConfig()

    token, err := conf.Exchange(context.Background(), c.Query("code"))
    if err != nil {
        c.String(http.StatusInternalServerError, "Token exchange failed: "+err.Error())
        return
    }
    conf = newGoogleOAuthConfig()
    client := conf.Client(context.Background(), token)
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        c.String(http.StatusInternalServerError, "Failed to fetch userinfo")
        return
    }
    defer resp.Body.Close()

    var gUser struct {
        Email         string `json:"email"`
        VerifiedEmail bool   `json:"verified_email"`
        Name          string `json:"name"`
        Picture       string `json:"picture"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&gUser); err != nil {
        c.String(http.StatusInternalServerError, "Failed to decode userinfo")
        return
    }

    db := c.MustGet("db").(*mongo.Database)
    ur := NewGoogleUserRepo(db)
    user, err := ur.GetOrCreateFromGoogle(gUser.Email, gUser.Name, gUser.Picture)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    jwtToken, err := security.GenerateJWT(user.AuthUserID, user.Email, user.Role)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "token_generation_failed"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": jwtToken})
}

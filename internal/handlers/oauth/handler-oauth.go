package oauth

import (
	"context"
	"net/http"
	"net/url"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"RAAS/core/config"
)

var googleOAuthConfig *oauth2.Config
var oauthToken *oauth2.Token // store token in memory (for demo; use DB/Redis in production)

// Initialize Google OAuth config
func InitGoogleOAuth(cfg *config.Config) {
	googleOAuthConfig = &oauth2.Config{
		ClientID:     cfg.Cloud.GoogleClientId,
		ClientSecret: cfg.Cloud.GoogleClientSecret,
		RedirectURL:  cfg.Cloud.GoogleRedirectURL, // must match GCP console (→ https://yourdomain.com/b1/auth/google/callback)
		Scopes: []string{
			gmail.GmailReadonlyScope,
			"email",
			"profile",
		},
		Endpoint: google.Endpoint,
	}
}

func GetGoogleConfig() *oauth2.Config {
	return googleOAuthConfig
}

// STEP 1: /b1/auth/google/login → Generate auth URL
func GoogleLogin(c *gin.Context) {
	state := "random-state" // ⚠️ replace with real per-session state
	url := GetGoogleConfig().AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.JSON(http.StatusOK, gin.H{"auth_url": url})
}

// STEP 2: /b1/auth/google/callback → Exchange code and redirect to frontend
func GoogleCallback(c *gin.Context) {
	ctx := context.Background()

	if c.Query("state") != "random-state" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	token, err := GetGoogleConfig().Exchange(ctx, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token exchange failed", "details": err.Error()})
		return
	}

	// Fetch user profile
	srv, err := gmail.NewService(ctx, option.WithTokenSource(GetGoogleConfig().TokenSource(ctx, token)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gmail service init failed", "details": err.Error()})
		return
	}

	profile, err := srv.Users.GetProfile("me").Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch profile failed", "details": err.Error()})
		return
	}

	// Redirect to frontend with token and email as query params
	frontendURL := "https://preview.arshan.digital/user/mail"
	params := url.Values{}
	params.Add("token", token.AccessToken)
	params.Add("email", profile.EmailAddress)
	redirectURL := frontendURL + "?" + params.Encode()

	c.Redirect(http.StatusFound, redirectURL)
}

// STEP 3: /b1/auth/google/mails → Fetch mails using token from query param
func GoogleRecentMails(c *gin.Context) {
	accessToken := c.Query("token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing access token"})
		return
	}

	ctx := context.Background()
	token := &oauth2.Token{AccessToken: accessToken}
	srv, err := gmail.NewService(ctx, option.WithTokenSource(GetGoogleConfig().TokenSource(ctx, token)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gmail service init failed", "details": err.Error()})
		return
	}

	msgs, err := srv.Users.Messages.List("me").MaxResults(10).Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch messages failed", "details": err.Error()})
		return
	}

	var emails []gin.H
	for _, m := range msgs.Messages {
		msg, err := srv.Users.Messages.Get("me", m.Id).Format("metadata").MetadataHeaders("Subject", "From").Do()
		if err != nil {
			continue
		}

		subject, from := "", ""
		for _, h := range msg.Payload.Headers {
			if h.Name == "Subject" {
				subject = h.Value
			} else if h.Name == "From" {
				from = h.Value
			}
		}

		emails = append(emails, gin.H{
			"id":      m.Id,
			"from":    from,
			"subject": subject,
			"snippet": msg.Snippet,
		})
	}

	c.JSON(http.StatusOK, gin.H{"messages": emails})
}

package controllers

import (
	"context"
	"fmt"
	"encoding/base64"
	"net/http"
	"os"

	"sci-stock-api/config"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func GoogleLogin(c *gin.Context) {
	url := config.GmailOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no code in request"})
		return
	}

	token, err := config.GmailOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot exchange code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expiry":        token.Expiry,
	})
}

func SendGmail(to, subject, htmlBody string, token *oauth2.Token) error {
	ctx := context.Background()
	srv, err := gmail.NewService(ctx, option.WithTokenSource(config.GmailOAuthConfig.TokenSource(ctx, token)))
	if err != nil {
		return err
	}

	from := os.Getenv("SUPPORT_EMAIL")
	messageStr := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", from, to, subject, htmlBody)
	message := &gmail.Message{
		Raw: encodeWeb64String([]byte(messageStr)),
	}

	_, err = srv.Users.Messages.Send("me", message).Do()
	return err
}

func encodeWeb64String(b []byte) string {
	return base64.URLEncoding.EncodeToString(b)
}
package config

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"golang.org/x/oauth2"
)

var GmailOAuthConfig *oauth2.Config

func GetGmailService(token *oauth2.Token) (*gmail.Service, error) {
	if GmailOAuthConfig == nil {
		return nil, fmt.Errorf("GmailOAuthConfig is not initialized")
	}

	ctx := context.Background()
	srv, err := gmail.NewService(ctx, option.WithTokenSource(GmailOAuthConfig.TokenSource(ctx, token)))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %v", err)
	}
	return srv, nil
}

func SendGmail(token *oauth2.Token, to, subject, htmlBody, plainText string) error {
	srv, err := GetGmailService(token)
	if err != nil {
		return err
	}

	raw := fmt.Sprintf("From: me\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", to, subject, htmlBody)
	msg := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(raw)),
	}

	_, err = srv.Users.Messages.Send("me", msg).Do()
	if err != nil {
		log.Printf("[ERROR] SendGmail failed: %v", err)
		return err
	}
	return nil
}
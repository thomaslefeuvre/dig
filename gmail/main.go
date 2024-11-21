package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func NewService(ctx context.Context, secretsDir string) (*gmail.Service, error) {
	b, err := os.ReadFile(fmt.Sprintf("%s/credentials.json", secretsDir))
	if err != nil {
		return nil, fmt.Errorf("read client secret file: %w", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("parse client secret file: %w", err)
	}

	tokenFile := fmt.Sprintf("%s/token.json", secretsDir)

	fmt.Printf("Attempting to read token file at %s\n", tokenFile)
	token, err := getTokenFromFile(tokenFile)
	if err != nil {
		fmt.Println("Unable to get token from file:", err)
		fmt.Println("Attempting to get token from web...")

		token, err = getTokenFromWeb(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("get token from web: %w", err)
		}

		err = saveToken(tokenFile, token)
		if err != nil {
			return nil, fmt.Errorf("save token to file: %w", err)
		}
	}

	fmt.Println("success")

	client := config.Client(ctx, token)

	fmt.Println("created client")

	return gmail.NewService(ctx, option.WithHTTPClient(client))
}

func getTokenFromFile(filename string) (*oauth2.Token, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var token oauth2.Token
	err = json.NewDecoder(f).Decode(&token)
	return &token, err
}

func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf("Click here to authorise: \n%v\n\n", authURL)
	fmt.Printf("Enter the authorisation code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("read authorisation code: %s", err)
	}

	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("exchange token: %s", err)
	}

	return token, nil
}

func saveToken(filename string, token *oauth2.Token) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

package main

import (
	"fmt"
	"os"
	"strings"
)

// Config
type Config struct {
	AWSRegion             string
	CryptoCodes           []string
	TwitterAccessToken    string
	TwitterAccessSecret   string
	TwitterConsumerKey    string
	TwitterConsumerSecret string
	SlackWebhookURL       string
}

// NewConfig initialises a new config
func NewConfig() (*Config, error) {
	awsRegion, err := getEnv("AWS_REGION")
	if err != nil {
		return nil, err
	}

	twitterAccessToken, err := getEnv("TWITTER_ACCESS_TOKEN")
	if err != nil {
		return nil, err
	}

	twitterAccessSecret, err := getEnv("TWITTER_ACCESS_SECRET")
	if err != nil {
		return nil, err
	}

	twitterConsumerKey, err := getEnv("TWITTER_CONSUMER_KEY")
	if err != nil {
		return nil, err
	}

	twitterConsumerSecret, err := getEnv("TWITTER_CONSUMER_SECRET")
	if err != nil {
		return nil, err
	}

	slackWebhookURL, err := getEnv("SLACK_WEBHOOK_URL")
	if err != nil {
		return nil, err
	}

	cryptoCodesStr, err := getEnv("CRYPTO_CODES")
	if err != nil {
		return nil, err
	}

	cryptoCodes := strings.Split(cryptoCodesStr, ",")

	return &Config{
		AWSRegion:             awsRegion,
		SlackWebhookURL:       slackWebhookURL,
		CryptoCodes:           cryptoCodes,
		TwitterAccessToken:    twitterAccessToken,
		TwitterAccessSecret:   twitterAccessSecret,
		TwitterConsumerKey:    twitterConsumerKey,
		TwitterConsumerSecret: twitterConsumerSecret,
	}, nil
}

func getEnv(key string) (string, error) {
	v := os.Getenv(key)

	if v == "" {
		return "", fmt.Errorf("%s environment variable missing", key)
	}

	return v, nil
}

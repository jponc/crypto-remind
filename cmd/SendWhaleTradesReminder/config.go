package main

import (
	"fmt"
	"os"
	"strings"
)

// Config
type Config struct {
	PhoneNumbers []string
	CryptoCodes  []string

	TwitterAccessToken    string
	TwitterAccessSecret   string
	TwitterConsumerKey    string
	TwitterConsumerSecret string
}

// NewConfig initialises a new config
func NewConfig() (*Config, error) {
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

	phoneNumbersStr, err := getEnv("PHONE_NUMBERS")
	if err != nil {
		return nil, err
	}

	phoneNumbers := strings.Split(phoneNumbersStr, ",")

	cryptoCodesStr, err := getEnv("CRYPTO_CODES")
	if err != nil {
		return nil, err
	}

	cryptoCodes := strings.Split(cryptoCodesStr, ",")

	return &Config{
		PhoneNumbers:          phoneNumbers,
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

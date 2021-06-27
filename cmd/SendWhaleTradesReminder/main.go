package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/jponc/crypto-remind/internal/reminder"

	log "github.com/sirupsen/logrus"
)

func main() {
	config, err := NewConfig()
	if err != nil {
		log.Fatalf("cannot initialise config %v", err)
	}

	twitterConfig := oauth1.NewConfig(config.TwitterConsumerKey, config.TwitterConsumerSecret)
	token := oauth1.NewToken(config.TwitterAccessToken, config.TwitterAccessSecret)
	httpClient := twitterConfig.Client(oauth1.NoContext, token)
	twitterClient := twitter.NewClient(httpClient)

	reminderService := reminder.NewService(twitterClient, config.PhoneNumbers, config.CryptoCodes)
	lambda.Start(reminderService.SendWhaleTradesReminder)
}

package reminder

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/jponc/crypto-remind/pkg/slack"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var cryptoCodeOverride = map[string]string{
	"#bitcoin": "$BTC",
}

const (
	whaleTradesTwitterScreenName = "WhaleTrades"
)

type Transaction struct {
	Amount     float64
	CryptoCode string
	Position   string
}

type Summary struct {
	Amount     float64
	CryptoCode string
}

// Service holds all dependency of this service
type Service struct {
	twitterClient   *twitter.Client
	slackClient     *slack.Client
	slackWebhookURL string
	cryptoCodes     []string
}

// NewService instantiates a new reminder service
func NewService(twitterClient *twitter.Client, slackClient *slack.Client, slackWebhookURL string, cryptoCodes []string) *Service {
	s := &Service{
		twitterClient:   twitterClient,
		slackClient:     slackClient,
		slackWebhookURL: slackWebhookURL,
		cryptoCodes:     cryptoCodes,
	}

	return s
}

func (s *Service) SendWhaleTradesReminder(ctx context.Context, snsEvent events.SNSEvent) {
	if s.twitterClient == nil {
		log.Fatalf("snsClient not initialised")
	}

	if s.slackClient == nil {
		log.Fatalf("slackClient not initialised")
	}

	if len(s.cryptoCodes) == 0 {
		log.Fatalf("cryptoCodes not initialised")
	}

	if s.slackWebhookURL == "" {
		log.Fatalf("slackWebhookURL not initialised")
	}

	params := &twitter.UserTimelineParams{
		ScreenName: whaleTradesTwitterScreenName,
		Count:      100,
	}

	tweets, _, err := s.twitterClient.Timelines.UserTimeline(params)
	if err != nil {
		log.Fatalf("failed to get user timeline of WhaleTrades: %v", err)
	}

	twoHoursAgo := time.Now().Add(time.Hour * -2)

	transactions, err := parseTweets(tweets, twoHoursAgo)
	if err != nil {
		log.Fatalf("failed to parse tweets: %v", err)
	}

	includedSummary, othersSummary := s.getTransactionsSummary(*transactions)

	includedTextArr := []string{}
	othersTextArr := []string{}
	p := message.NewPrinter(language.English)

	for _, s := range includedSummary {
		includedTextArr = append(includedTextArr, p.Sprintf("%s: %.2f", s.CryptoCode, s.Amount))
	}

	for _, s := range othersSummary {
		othersTextArr = append(othersTextArr, p.Sprintf("%s: %.2f", s.CryptoCode, s.Amount))
	}

	msg := fmt.Sprintf(
		"2H WhaleTrades Summary:\n%s\n\nOthers:\n%s",
		strings.Join(includedTextArr, "\n"),
		strings.Join(othersTextArr, "\n"),
	)

	err = s.slackClient.SendMessageToWebhook(ctx, msg, s.slackWebhookURL)
	if err != nil {
		log.Fatalf("failed to send message to slack: %v", err)
	}

	return
}

func (s *Service) getTransactionsSummary(transactions []Transaction) (includedSummary []Summary, othersSummary []Summary) {
	includedMap := map[string]float64{}
	othersMap := map[string]float64{}
	cryptoCodesMap := map[string]bool{}

	// create set for cryptoCodes we're interested in
	for _, c := range s.cryptoCodes {
		cryptoCodesMap[c] = true
	}

	// generate included or others map depending on the cryptoCode
	for _, t := range transactions {
		amount := t.Amount
		if strings.Contains(t.Position, "SHORTED") {
			amount = amount * -1
		}

		if _, found := cryptoCodesMap[t.CryptoCode]; found {
			includedMap[t.CryptoCode] = includedMap[t.CryptoCode] + amount
		} else {
			othersMap[t.CryptoCode] = othersMap[t.CryptoCode] + amount
		}
	}

	for code, total := range includedMap {
		includedSummary = append(includedSummary, Summary{
			CryptoCode: code,
			Amount:     total,
		})
	}

	for code, total := range othersMap {
		othersSummary = append(othersSummary, Summary{
			CryptoCode: code,
			Amount:     total,
		})

	}

	return includedSummary, othersSummary
}

func parseTweets(tweets []twitter.Tweet, fromTime time.Time) (*[]Transaction, error) {
	transactions := []Transaction{}

	for _, t := range tweets {
		createdAt, err := time.Parse(time.RubyDate, t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time: %v", t.CreatedAt)
		}

		if createdAt.After(fromTime) {
			if strings.Contains(t.Text, "SHORTED") || strings.Contains(t.Text, "LONGED") {
				transaction, err := parseShortLongTweet(t)
				if err == nil {
					transactions = append(transactions, *transaction)
				} else {
					log.Errorf("failed to parse long/short tweet: %s", t.Text)
				}
			} else {
				log.Errorf("unknown: %s", t.Text)
			}
		}
	}

	return &transactions, nil
}

func parseShortLongTweet(t twitter.Tweet) (*Transaction, error) {
	// Get all relevant information
	re, err := regexp.Compile(`(\$[\d,?\.]+) ([\$|#]\w+) (#?\w+)`)
	if err != nil {
		return nil, fmt.Errorf("can't compile regex: %v", err)
	}

	submatches := re.FindStringSubmatch(t.Text)
	if len(submatches) == 0 {
		return nil, fmt.Errorf("failed to find match: %s", t.Text)
	}

	// Get amount
	amountStr := strings.ReplaceAll(
		strings.ReplaceAll(submatches[1], "$", ""),
		",",
		"",
	)
	f, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert %s to float64", amountStr)
	}

	// Get crypto code
	code := submatches[2]
	cryptoCode := strings.ToUpper(code)
	if override, found := cryptoCodeOverride[code]; found {
		cryptoCode = override
	}

	// Get position
	position := submatches[3]

	return &Transaction{
		Amount:     f,
		CryptoCode: cryptoCode,
		Position:   position,
	}, nil
}

func parseLiquidatedTweet(t twitter.Tweet) (*Transaction, error) {
	// Get all relevant information
	re, err := regexp.Compile(`(\$[\d,?\.]+) #LIQUIDATED (\$[\w]+) `)
	if err != nil {
		return nil, fmt.Errorf("can't compile regex: %v", err)
	}

	submatches := re.FindStringSubmatch(t.Text)
	if len(submatches) == 0 {
		return nil, fmt.Errorf("failed to find match: %s", t.Text)
	}

	// Get amount
	amountStr := strings.ReplaceAll(
		strings.ReplaceAll(submatches[1], "$", ""),
		",",
		"",
	)
	f, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert %s to float64", amountStr)
	}

	// Get crypto code
	code := submatches[2]
	cryptoCode := strings.ToUpper(code)
	if override, found := cryptoCodeOverride[code]; found {
		cryptoCode = override
	}

	// Get position
	return &Transaction{
		Amount:     f,
		CryptoCode: cryptoCode,
		Position:   "SHORTED",
	}, nil
}

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
	log "github.com/sirupsen/logrus"
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
	twitterClient *twitter.Client
	phoneNumbers  []string
	cryptoCodes   []string
}

// NewService instantiates a new reminder service
func NewService(twitterClient *twitter.Client, phoneNumbers []string, cryptoCodes []string) *Service {
	s := &Service{
		twitterClient: twitterClient,
		phoneNumbers:  phoneNumbers,
		cryptoCodes:   cryptoCodes,
	}

	return s
}

func (s *Service) SendWhaleTradesReminder(ctx context.Context, snsEvent events.SNSEvent) {
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

	included, others := s.getTransactionsSummary(*transactions)

	log.Infof("included: %v", included)
	log.Infof("others: %v", others)

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
			// Get all relevant information
			re, err := regexp.Compile(`(\$[\d,?]+) ([\$|#]\w+) (#?\w+)`)
			if err != nil {
				log.Infof("failed to compile regex")
				continue
			}

			submatches := re.FindStringSubmatch(t.Text)
			if len(submatches) == 0 {
				log.Errorf("failed to find match: %s", t.Text)
				continue
			}

			// Get amount
			amountStr := strings.ReplaceAll(
				strings.ReplaceAll(submatches[1], "$", ""),
				",",
				"",
			)
			f, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				log.Errorf("failed to convert %s to float64", amountStr)
				continue
			}

			// Get crypto code
			code := submatches[2]
			cryptoCode := code
			if override, found := cryptoCodeOverride[code]; found {
				cryptoCode = override
			}

			// Get position
			position := submatches[3]

			transactions = append(transactions, Transaction{
				Amount:     f,
				CryptoCode: cryptoCode,
				Position:   position,
			})
		}
	}

	return &transactions, nil
}

package sns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	awsSns "github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-xray-sdk-go/xray"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	awsSnsClient *awsSns.SNS
}

// NewClient instantiates a SNS client
func NewClient(awsRegion string) (*Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})

	if err != nil {
		return nil, fmt.Errorf("cannot create aws session: %v", err)
	}

	awsSnsClient := sns.New(sess)
	xray.AWS(awsSnsClient.Client)

	c := &Client{
		awsSnsClient: awsSnsClient,
	}

	return c, nil
}

func (c *Client) SendSMS(ctx context.Context, message, phoneNumber string) error {
	input := &awsSns.PublishInput{
		Message:     aws.String(message),
		PhoneNumber: aws.String(phoneNumber),
	}

	out, err := c.awsSnsClient.PublishWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to publish to sns: %v", err)
	}

	log.Infof("SNS: %v", out)

	return nil
}

package modules

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Client struct {
	Downloader *s3manager.Downloader
	Bucket     string
}

func (c *Client) S3Init() {
	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String("ap-southeast-1"),
		Credentials: credentials.NewEnvCredentials(),
	})
	c.Downloader = s3manager.NewDownloader(sess)
}

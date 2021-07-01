package okutil

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	DefaultAwsRegion = "ap-northeast-1"
	AwsKeyEnvName    = "AWS_ACCESS_KEY_ID"
	AwsSecretEnvName = "AWS_SECRET_ACCESS_KEY"
	AwsProfileName   = "AWS_PROFILE"
	Ec2MetaUrl       = "http://169.254.169.254/latest/dynamic/instance-identity/document"
)

func CreateAWSSession(prod bool) (*session.Session, error) {
	var creds *credentials.Credentials
	if prod {
		creds = credentials.NewCredentials(&ec2rolecreds.EC2RoleProvider{})
	} else {
		// dev mode using profile name or API secret/key env variable
		awsProfile := os.Getenv(AwsProfileName)
		if len(awsProfile) > 0 {
			// prefer to aws profile
			creds = credentials.NewSharedCredentials("", awsProfile)
		} else {
			creds = credentials.NewStaticCredentials(
				os.Getenv(AwsKeyEnvName), os.Getenv(AwsSecretEnvName), "")
		}
	}

	return session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(DefaultAwsRegion),
	})
}

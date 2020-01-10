package clients

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/sirupsen/logrus"
)

// AwsConfig represents the main aws.Config with the secret names for Elastic and Redis attached
type AwsConfig struct {
	Main        aws.Config
	Secrets     []Secrets
	Credentials []Credentials
}

// Secrets represents a mapping of Secret string and type for use in the SecretsManagerClient
type Secrets struct {
	Type   string // This type should follow the resolved credentials around, to make it easy to figure out which client these creds belong to
	Secret string
}

// CredentialsWrapper wraps the credentials struct in the type passed from Secrets
type CredentialsWrapper struct {
	Type        string
	Credentials Credentials
}

// SecretInput matches the secretsmanager.GetSecretValueInput struct but alows us to mock the service
type SecretInput struct {
	Client secretsmanageriface.SecretsManagerAPI
	Input  *secretsmanager.GetSecretValueInput
}

// Credentials represents the unmarshalled JSON object from SecretsManager
type Credentials struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password"`
	Endpoint string `json:"endpoint"`
	Database string `json:"database,omitempty"`
}

// Secret is a concrete representation of the SecretsManager response
type Secret struct {
	ARN           string     `json:"ARN"`
	CreatedDate   *time.Time `json:"CreatedDate"`
	Name          string     `json:"Name"`
	Secret        string     `json:"SecretString"`
	VersionID     string     `json:"VersionId"`
	VersionStages []string   `json:"VersionStages"`
}

// SecretsManagerClient represents the AWS Secrets Manager Client
func SecretsManagerClient(awsConfig AwsConfig, logger *logrus.Logger) (c []CredentialsWrapper, err error) {
	// Create a new AWS Session
	sess := session.Must(session.NewSession(&awsConfig.Main))

	// Get the AWS credentials from the environment or Shared Credentials file where the function is running
	_, err = sess.Config.Credentials.Get()
	if err != nil {
		logger.Error(err)
	}

	for _, s := range awsConfig.Secrets {
		var creds CredentialsWrapper
		// Create the request object for SecretsManager using the secretName from the environment
		input := SecretInput{
			Client: secretsmanager.New(sess),
			Input: &secretsmanager.GetSecretValueInput{
				SecretId:     aws.String(s.Secret),
				VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
			},
		}

		result, err := input.Client.GetSecretValue(input.Input)
		if err != nil {
			return c, err
		}

		sec := Secret{}

		// Determine if SecretString is a string pointer or binary to decode
		if result.SecretString != nil {
			sec.Secret = *result.SecretString

			// Unmarshall Secret string
			err = json.Unmarshal([]byte(sec.Secret), &creds.Credentials)
			if err != nil {
				return c, err
			}

			creds.Type = s.Type
			c = append(c, creds)
		}
	}

	return c, nil
}

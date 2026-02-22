package internal

type Environment struct {
	OIDCClientID     string `env:"OIDC_CLIENT_ID"`
	OIDCClientSecret string `env:"OIDC_CLIENT_SECRET"`
	SessionSecret    string `env:"SESSION_SECRET"`
	SessionState     string `env:"SESSION_STATE"`
	BaseUri          string `env:"BASE_URI"`
	DatabaseUri      string `env:"DATABASE_URI"`
	BucketName       string `env:"BUCKET_NAME"`
}

var env Environment

func GetEnv() *Environment {
	return &env
}

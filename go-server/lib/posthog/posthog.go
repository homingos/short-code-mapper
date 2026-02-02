package common

import (
	"github.com/homingos/campaign-svc/config"
	"github.com/posthog/posthog-go"
)

var PostHogClient posthog.Client

func InitializePostHog() posthog.Client {
	client, _ := posthog.NewWithConfig(
		config.NewPosthogConfig().PostHogApiKey,
		posthog.Config{
			PersonalApiKey: config.NewPosthogConfig().PostHogApiKey,
			Endpoint:       config.NewPosthogConfig().Endpoint,
		},
	)
	PostHogClient = client
	return PostHogClient
}

func GetPostHogClient() posthog.Client {
	if PostHogClient == nil {
		InitializePostHog()
	}

	return PostHogClient
}

package features

import (
	"context"
	"log"
	"os"
	"time"

	flagsmith "github.com/Flagsmith/flagsmith-go-client/v3"
)

type FeatureManager struct {
	client *flagsmith.Client
}

var Instance *FeatureManager

// InitFlagsmith sets up the server-side Flagsmith client
func InitFlagsmith() {
	apiKey := os.Getenv("FLAGSMITH_ENVIRONMENT_KEY")
	if apiKey == "" {
		log.Println("FLAGSMITH_ENVIRONMENT_KEY not provided. Feature flagging running in fallback mode.")
	}

	// Configuring client with optimal local evaluation polling for low latency
	client := flagsmith.NewClient(apiKey,
		flagsmith.WithLocalEvaluation(context.Background()),
		flagsmith.WithEnvironmentRefreshInterval(1*time.Minute),
	)

	Instance = &FeatureManager{client: client}
	log.Println("Flagsmith Engine initialized successfully.")
}

// IsFeatureEnabled checks if a specific feature flag is turned on global-wide
func (m *FeatureManager) IsFeatureEnabled(ctx context.Context, featureKey string) bool {
	if m.client == nil {
		return false // Secure default fallback
	}

	flags, err := m.client.GetEnvironmentFlags(ctx)
	if err != nil {
		log.Printf("❌ Failed to resolve environment flags: %v. Using defaults.", err)
		return false
	}

	enabled, _ := flags.IsFeatureEnabled(featureKey)
	return enabled
}
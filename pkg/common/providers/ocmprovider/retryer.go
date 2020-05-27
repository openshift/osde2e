package ocmprovider

import (
	"log"
	"sync"
	"time"

	"github.com/adamliesko/retry"
	"github.com/spf13/viper"
)

var ocmOnce = sync.Once{}
var ocmRetryer *retry.Retryer

// Retryer returns a retryer meant for OCM interactions.
func retryer() *retry.Retryer {
	ocmOnce.Do(func() {
		ocmRetryer = retry.New(retry.SleepFn(func(attempts int) {
			time.Sleep(time.Duration(2^attempts) * time.Second)
		}))
		ocmRetryer.Tries = viper.GetInt(NumRetries)
		ocmRetryer.AfterEachFailFn = func(err error) {
			log.Printf("error during OCM attempt: %v", err)
		}
	})

	return ocmRetryer
}

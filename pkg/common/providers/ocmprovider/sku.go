package ocmprovider

import (
	"log"
	"math/rand"
	"strings"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
)

func getSKU() string {
	var skuID string
	// Retrieve SKU based on config
	// If multiple flavours are supplied in a comma-delimited list
	// select one at random and will override the config with the
	// flavor selected for all future calls
	skus := strings.Split(viper.GetString(Sku), ",")
	skuLength := len(skus)
	switch skuLength {
	case 0:
		skuID = ""
	case 1:
		skuID = skus[0]
	default:
		skuID = skus[rand.Intn(skuLength)]
	}
	log.Printf("Using SKU: %s", skuID)

	return skuID
}

package ocmprovider

import (
	"log"
	"math/rand"
	"strings"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
)

func getFlavour() string {
	var flavourID string
	// Retrieve flavour based on config
	// If multiple flavours are supplied in a comma-delimited list
	// select one at random and will override the config with the
	// flavor selected for all future calls
	flavours := strings.Split(viper.GetString(Flavour), ",")
	flavourLength := len(flavours)
	switch flavourLength {
	case 0:
		flavourID = ""
	case 1:
		flavourID = flavours[0]
	default:
		flavourID = flavours[rand.Intn(flavourLength)]
	}
	log.Printf("Using flavour: %s", flavourID)

	return flavourID
}

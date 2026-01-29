package load

import (
	"os"
	"testing"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

func TestLoadPassthruSecrets_WithSecretVarPrefix(t *testing.T) {
	// Setup: Set environment variables with SecretVarPrefix
	testEnvVars := map[string]string{
		"EXTERNAL_SECRET_MY_SECRET":   "secret_value_1",
		"EXTERNAL_SECRET_API_TOKEN":   "token_123",
		"EXTERNAL_SECRET_DB_PASSWORD": "db_pass_456",
		"EXTERNAL_SECRET_WITH_EQUALS": "value=with=equals",
		"REGULAR_ENV_VAR":             "should_not_be_included",
		"ANOTHER_REGULAR_VAR":         "also_not_included",
	}

	// Set the environment variables
	for key, value := range testEnvVars {
		err := os.Setenv(key, value)
		if err != nil {
			t.Fatalf("Failed to set environment variable %s: %v", key, err)
		}
	}

	// Cleanup: Unset environment variables after test
	defer func() {
		for key := range testEnvVars {
			os.Unsetenv(key)
		}
	}()

	// Initialize viper with an empty map for NonOSDe2eSecrets
	viper.Set(config.NonOSDe2eSecrets, map[string]string{})

	// Call the function under test
	loadPassthruSecrets([]string{})

	// Retrieve the passthrough secrets from viper
	result := viper.GetStringMapString(config.NonOSDe2eSecrets)

	// Verify that environment variables with SecretVarPrefix were loaded
	expectedSecrets := map[string]string{
		"EXTERNAL_SECRET_MY_SECRET":   "secret_value_1",
		"EXTERNAL_SECRET_API_TOKEN":   "token_123",
		"EXTERNAL_SECRET_DB_PASSWORD": "db_pass_456",
		"EXTERNAL_SECRET_WITH_EQUALS": "value=with=equals",
	}

	for key, expectedValue := range expectedSecrets {
		actualValue, exists := result[key]
		if !exists {
			t.Errorf("Expected environment variable %s to be loaded, but it was not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected value for %s to be %q, but got %q", key, expectedValue, actualValue)
		}
	}

	// Verify that regular environment variables (without the prefix) are NOT loaded
	notExpectedVars := []string{"REGULAR_ENV_VAR", "ANOTHER_REGULAR_VAR"}
	for _, key := range notExpectedVars {
		if _, exists := result[key]; exists {
			t.Errorf("Environment variable %s should not have been loaded (missing SecretVarPrefix)", key)
		}
	}
}

func TestLoadPassthruSecrets_HandlesEqualsInValue(t *testing.T) {
	// Test that values containing '=' are handled correctly
	testKey := "EXTERNAL_SECRET_TEST_EQUALS"
	testValue := "key1=value1;key2=value2"

	err := os.Setenv(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer os.Unsetenv(testKey)

	// Initialize viper
	viper.Set(config.NonOSDe2eSecrets, map[string]string{})

	// Call the function
	loadPassthruSecrets([]string{})

	// Verify
	result := viper.GetStringMapString(config.NonOSDe2eSecrets)
	actualValue, exists := result[testKey]

	if !exists {
		t.Errorf("Expected environment variable %s to be loaded", testKey)
	} else if actualValue != testValue {
		t.Errorf("Expected value %q, but got %q", testValue, actualValue)
	}
}

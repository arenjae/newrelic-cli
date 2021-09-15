package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/newrelic/newrelic-cli/internal/config"
	configAPI "github.com/newrelic/newrelic-cli/internal/config/api"
	"github.com/newrelic/newrelic-cli/internal/utils"
	"github.com/newrelic/newrelic-client-go/pkg/region"
)

type PlatformLinkGenerator struct {
	httpClient utils.HTTPClientInterface
}

var nrPlatformHostnames = struct {
	Staging string
	US      string
	EU      string
}{
	Staging: "staging-one.newrelic.com",
	US:      "one.newrelic.com",
	EU:      "one.eu.newrelic.com",
}

func NewPlatformLinkGenerator(httpClient utils.HTTPClientInterface) *PlatformLinkGenerator {
	return &PlatformLinkGenerator{
		httpClient: httpClient,
	}
}

func (g *PlatformLinkGenerator) GenerateExplorerLink(status InstallStatus) string {
	return generateExplorerLink(status)
}

func (g *PlatformLinkGenerator) GenerateEntityLink(entityGUID string) string {
	return generateEntityLink(entityGUID)
}

// GenerateRedirectURL creates a URL for the user to navigate to after running
// through an installation. The URL is displayed in the CLI out as well and is
// also provided in the nerdstorage document. This provides the user two options
// to see their data - click from the CLI output or from the frontend.
func (g *PlatformLinkGenerator) GenerateRedirectURL(status InstallStatus) string {
	if status.AllSelectedRecipesInstalled() {
		return g.GenerateEntityLink(status.HostEntityGUID())
	}

	return g.GenerateExplorerLink(status)
}

type referrerParamValue struct {
	NerdletID  string `json:"nerdletId,omitempty"`
	Referrer   string `json:"referrer,omitempty"`
	EntityGUID string `json:"entityGuid,omitempty"`
}

// The CLI URL referrer param is a JSON string containing information
// the UI can use to understand how/where the URL was generated. This allows the
// UI to return to its previous state in the case of a user closing the browser
// and then clicking a redirect URL in the CLI's output.
func generateReferrerParam(entityGUID string) string {
	p := referrerParamValue{
		NerdletID: "nr1-install-newrelic.installation-plan",
		Referrer:  "newrelic-cli",
	}

	if entityGUID != "" {
		p.EntityGUID = entityGUID
	}

	stringifiedParam, err := json.Marshal(p)
	if err != nil {
		log.Debugf("error marshaling referrer param: %s", err)
		return ""
	}

	return string(stringifiedParam)
}

func generateExplorerLink(status InstallStatus) string {
	longURL := generateExplorerURL(status)
	shortURL, err := generateShortNewRelicURL(longURL)
	if err != nil || shortURL == "" {
		return longURL
	}

	return shortURL
}

func generateExplorerURL(status InstallStatus) string {
	return fmt.Sprintf("https://%s/launcher/nr1-core.explorer?platform[filters]=%s&platform[accountId]=%d&cards[0]=%s",
		nrPlatformHostname(),
		utils.Base64Encode(status.successLinkConfig.Filter),
		configAPI.GetActiveProfileAccountID(),
		utils.Base64Encode(generateReferrerParam(status.HostEntityGUID())),
	)
}

func generateEntityLink(entityGUID string) string {
	longURL := generateEntityURL(entityGUID)
	shortURL, err := generateShortNewRelicURL(longURL)
	if err != nil || shortURL == "" {
		return longURL
	}

	return shortURL
}

func generateEntityURL(entityGUID string) string {
	return fmt.Sprintf("https://%s/redirect/entity/%s", nrPlatformHostname(), entityGUID)
}

// The generateShortNewRelicURL function utilizes a New Relic service to
// convert a New Relic URL to a shortened version of the provided URL.
// The shortened URL redirects to the original URL.
//
// If an error occurs while attempting to shorten the URL, the original
// long URL is returned along with the error.
//
// Note: This API only works in a production environment.
func generateShortNewRelicURL(longURL string) (string, error) {
	const shortURLServiceURL = "https://urly.service.newrelic.com/"

	httpClient := utils.NewHTTPClient(os.Getenv("NEW_RELIC_API_KEY"))

	reqBody := []byte(fmt.Sprintf(`{"url": "%s"}`, longURL))
	respBytes, err := httpClient.Post(context.Background(), shortURLServiceURL, reqBody)
	if err != nil {
		log.WithFields(log.Fields{
			"longURL":      longURL,
			"errorMessage": err.Error(),
		}).Debugf("error creating short URL")

		return longURL, err
	}

	var resp struct {
		URL string `json:"url"`
	}

	err = json.Unmarshal(respBytes, &resp)
	if err != nil {
		log.Debugf("error unmarshaling short URL API response: %s", err)
		return "", err
	}

	return resp.URL, nil
}

// nrPlatformHostname returns the host for the platform based on the region set.
func nrPlatformHostname() string {
	r := configAPI.GetActiveProfileString(config.Region)
	if strings.EqualFold(r, region.Staging.String()) {
		return nrPlatformHostnames.Staging
	}

	if strings.EqualFold(r, region.EU.String()) {
		return nrPlatformHostnames.EU
	}

	return nrPlatformHostnames.US
}

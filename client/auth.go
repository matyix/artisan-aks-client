package client

import (
	"os"

	"github.com/Azure/azure-sdk-for-go/arm/examples/helpers"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/banzaicloud/azure-aks-client/cluster"
	log "github.com/sirupsen/logrus"
	"encoding/json"
)

var sdk cluster.Sdk

func init() {
	// Log as JSON
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

const azureClientId = "AZURE_CLIENT_ID"
const azureClientSecret = "AZURE_CLIENT_SECRET"
const azureSubscriptionId = "AZURE_SUBSCRIPTION_ID"
const azureTenantId = "AZURE_TENANT_ID"

func Authenticate() (*cluster.Sdk, *InitErrorResponse) {
	clientId := os.Getenv(azureClientId)
	clientSecret := os.Getenv(azureClientSecret)
	subscriptionId := os.Getenv(azureSubscriptionId)
	tenantId := os.Getenv(azureTenantId)

	// ---- [Check Environmental variables] ---- //
	if len(clientId) == 0 {
		return nil, createEnvErrorResponse(azureClientId)
	}

	if len(clientSecret) == 0 {
		return nil, createEnvErrorResponse(azureClientSecret)
	}

	if len(subscriptionId) == 0 {
		return nil, createEnvErrorResponse(azureSubscriptionId)
	}

	if len(tenantId) == 0 {
		return nil, createEnvErrorResponse(azureTenantId)
	}

	sdk = cluster.Sdk{
		ServicePrincipal: &cluster.ServicePrincipal{
			ClientID:       clientId,
			ClientSecret:   clientSecret,
			SubscriptionID: subscriptionId,
			TenantId:       tenantId,
			HashMap: map[string]string{
				"AZURE_CLIENT_ID":       clientId,
				"AZURE_CLIENT_SECRET":   clientSecret,
				"AZURE_SUBSCRIPTION_ID": subscriptionId,
				"AZURE_TENANT_ID":       tenantId,
			},
		},
	}

	authenticatedToken, err := helpers.NewServicePrincipalTokenFromCredentials(sdk.ServicePrincipal.HashMap, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return nil, createAuthErrorResponse(err)
	}

	sdk.ServicePrincipal.AuthenticatedToken = authenticatedToken

	resourceGroup := resources.NewGroupsClient(sdk.ServicePrincipal.SubscriptionID)
	resourceGroup.Authorizer = autorest.NewBearerAuthorizer(sdk.ServicePrincipal.AuthenticatedToken)
	sdk.ResourceGroup = &resourceGroup

	return &sdk, nil
}

func GetSdk() *cluster.Sdk {
	return &sdk
}

type InitErrorResponse struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (e InitErrorResponse) toString() string {
	jsonResponse, _ := json.Marshal(e)
	return string(jsonResponse)
}

func createEnvErrorResponse(env string) *InitErrorResponse {
	message := "Environmental variable is empty: " + env
	log.WithFields(log.Fields{"error": "environmental_error"}).Error(message)
	return &InitErrorResponse{StatusCode: internalErrorCode, Message: message}
}

func createAuthErrorResponse(err error) *InitErrorResponse {
	errMsg := "Failed to authenticate with Azure"
	log.WithFields(log.Fields{"Authentication error": err}).Error(errMsg)
	return &InitErrorResponse{StatusCode: internalErrorCode, Message: errMsg}
}

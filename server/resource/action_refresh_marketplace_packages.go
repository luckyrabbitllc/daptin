package resource

import (
	//"golang.org/x/oauth2"
	log "github.com/sirupsen/logrus"
)

type RefreshMarketplacePackagelistPerformer struct {
	cruds     map[string]*DbResource
	cmsConfig *CmsConfig
}

func (d *RefreshMarketplacePackagelistPerformer) Name() string {
	return "marketplace.package.refresh"
}

func (d *RefreshMarketplacePackagelistPerformer) DoAction(request ActionRequest, inFieldMap map[string]interface{}) ([]ActionResponse, []error) {

	marketReferenceId := inFieldMap["marketplace_id"].(string)
	marketplaceHandler, ok := d.cmsConfig.MarketplaceHandlers[marketReferenceId]

	if !ok {

		marketPlace, err := d.cruds["marketplace"].GetMarketplaceByReferenceId(marketReferenceId)
		if err != nil {
			return nil, []error{err}
		}

		handler, err := NewMarketplaceService(marketPlace)
		if err != nil {
			log.Errorf("Failed to create new market place service")
		}
		d.cmsConfig.MarketplaceHandlers[marketReferenceId] = handler
		go handler.RefreshRepository()
		return marketResfreshSuccessResponse, nil
	}
	err := marketplaceHandler.RefreshRepository()
	return marketResfreshSuccessResponse, []error{err}
}

var marketResfreshSuccessResponse = []ActionResponse{
	NewActionResponse("client.notify", map[string]interface{}{
		"type":    "success",
		"message": "Marketplace package list is being refreshed",
		"title":   "Success",
	}),
}

func NewRefreshMarketplacePackagelistPerformer(initConfig *CmsConfig, cruds map[string]*DbResource) (ActionPerformerInterface, error) {

	handler := RefreshMarketplacePackagelistPerformer{
		cruds:     cruds,
		cmsConfig: initConfig,
	}
	return &handler, nil

}

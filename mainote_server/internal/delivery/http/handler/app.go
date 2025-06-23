package handler

import (
	"context"

	"mainote-server/internal/usecase"
	api "mainote-server/pkg/generated/api"
)

// AppHandler handles HTTP requests for apps and implements AppsAPIServicer.
type AppHandler struct {
	appUsecase usecase.AppUsecase
}

// NewAppHandler creates a new instance of AppHandler.
func NewAppHandler(appUsecase usecase.AppUsecase) *AppHandler {
	return &AppHandler{appUsecase: appUsecase}
}

// GetApps implements AppsAPIServicer interface.
func (h *AppHandler) GetApps(ctx context.Context) (api.ImplResponse, error) {
	apps, err := h.appUsecase.GetAllApps(ctx)
	if err != nil {
		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}

	// Convert domain apps to API response
	var appResponses []api.AppResponse
	for _, app := range apps {
		appResponse := api.AppResponse{
			AppId:       app.AppID.String(),
			Name:        app.Name,
			Description: app.Description,
		}
		appResponses = append(appResponses, appResponse)
	}

	response := api.AppsListResponse{
		Apps:  appResponses,
		Count: int32(len(appResponses)),
	}

	return api.Response(200, response), nil
}

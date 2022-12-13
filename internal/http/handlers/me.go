package handlers

import (
	"errors"
	"net/http"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/get_user_by_session_token"
	"remindme/internal/http/serializers"
)

type Me struct {
	service services.Service[service.Input, service.Result]
}

func NewMe(
	service services.Service[service.Input, service.Result],
) *Me {
	return &Me{service: service}
}

type MeResult struct {
	User serializers.User `json:"user"`
}

func (s *Me) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	token, ok := GetAuthToken(r)
	if !ok {
		renderUnauthorizedResponse(rw)
		return
	}

	result, err := s.service.Run(
		r.Context(),
		service.Input{Token: token},
	)
	if errors.Is(err, user.ErrUserDoesNotExist) {
		renderUnauthorizedResponse(rw)
		return
	}
	if err != nil {
		renderErrorResponse(rw, "internal error", http.StatusInternalServerError)
		return
	}

	user := serializers.User{}
	user.FromDomainUser(result.User)
	renderResponse(rw, MeResult{User: user}, http.StatusOK)
}

package httptransport

type Handler struct {
	authService AuthService
	tokenParser AccessTokenParser
	refreshCookieSecure bool
}

func NewHandler(authService AuthService, tokenParser AccessTokenParser, refreshCookieSecure bool) *Handler {
	return &Handler{
		authService:         authService,
		tokenParser: tokenParser,
		refreshCookieSecure: refreshCookieSecure,
	}
}

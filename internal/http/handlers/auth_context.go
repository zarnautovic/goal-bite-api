package handlers

import (
	"net/http"

	httpmiddleware "nutrition/internal/http/middleware"
)

func requireAuthUserID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	userID, ok := httpmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return 0, false
	}
	return userID, true
}

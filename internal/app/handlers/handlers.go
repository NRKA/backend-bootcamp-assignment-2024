package handlers

import (
	"net/http"
)

// TODO: No Auth
func DummyLoginHandler(w http.ResponseWriter, r *http.Request) {
	userType := r.URL.Query().Get("user_type")

	switch userType {
	case "client":
		w.Write([]byte("auth_token_client"))
	case "moderator":
		w.Write([]byte("auth_token_moderator"))
	default:
		http.Error(w, "Invalid role: Invalid request or missing user_type", http.StatusBadRequest)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

}

// TODO: Auth Only
func HouseHandler(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")
	w.Write([]byte(id))
}

func SubscribeHouseHandler(w http.ResponseWriter, r *http.Request) {

}

func CreateFlatHandler(w http.ResponseWriter, r *http.Request) {

}

// TODO: Moderations Only
func CreateHouseHandler(w http.ResponseWriter, r *http.Request) {

}

func UpdateFlatHandler(w http.ResponseWriter, r *http.Request) {

}

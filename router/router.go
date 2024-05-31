package routes

import (
	"banking/controllers"
	"github.com/gorilla/mux"
	"net/http"
)

const password = "mohit123"

// Middleware function to check if the password is correct
func passwordMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := r.URL.Query().Get("password")
		if pass != password {
			http.Error(w, "Incorrect password", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Routers() *mux.Router {
	router := mux.NewRouter()

	router.Use(passwordMiddleware)
	
	router.HandleFunc("/account", controllers.CreateAccountHandler).Methods("POST")
	router.HandleFunc("/accounts", controllers.GetAllAccountHandler).Methods("GET")
	router.HandleFunc("/account/{id}", controllers.GetOneAccountHandler).Methods("GET")
	router.HandleFunc("/account/{id}", controllers.DeleteAccountHandler).Methods("DELETE")
	router.HandleFunc("/accounts", controllers.DeleteAllAccountsHandler).Methods("DELETE")
	return router
}

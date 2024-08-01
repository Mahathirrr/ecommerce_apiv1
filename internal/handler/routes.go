package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

var r *mux.Router

func RegisterRoutes(h *handler) *mux.Router {
	r = mux.NewRouter()
	tokenMaker := h.TokenMaker

	// Products
	r.HandleFunc("/products", h.Listproducts).Methods("GET")
	r.HandleFunc("/products/{id}", h.getProduct).Methods("GET")

	// Admin Product routes
	adminProductRouter := r.PathPrefix("/products").Subrouter()
	adminProductRouter.Use(GetAdminMiddlewareFunc(tokenMaker))
	adminProductRouter.HandleFunc("", h.createProduct).Methods("POST")
	adminProductRouter.HandleFunc("/{id}", h.updateProducts).Methods("PATCH")
	adminProductRouter.HandleFunc("/{id}", h.DeleteProduct).Methods("DELETE")

	// Auth required routes
	authRouter := r.PathPrefix("").Subrouter()
	authRouter.Use(GetAuthMiddlewareFunc(tokenMaker))

	// Orders
	authRouter.HandleFunc("/myorder", h.getOrder).Methods("GET")
	authRouter.HandleFunc("/orders", h.createOrder).Methods("POST")
	authRouter.HandleFunc("/orders/{id}", h.deleteOrder).Methods("DELETE")

	// Admin Order routes
	adminOrderRouter := authRouter.PathPrefix("/orders").Subrouter()
	adminOrderRouter.Use(GetAdminMiddlewareFunc(tokenMaker))
	adminOrderRouter.HandleFunc("", h.listOrders).Methods("GET")

	// Users
	r.HandleFunc("/users", h.createUser).Methods("POST")
	r.HandleFunc("/users/login", h.loginUser).Methods("POST")

	authRouter.HandleFunc("/users", h.updateUser).Methods("PATCH")
	authRouter.HandleFunc("/users/logout", h.logoutUser).Methods("POST")

	// Admin User routes
	adminUserRouter := r.PathPrefix("/users").Subrouter()
	adminUserRouter.Use(GetAdminMiddlewareFunc(tokenMaker))
	adminUserRouter.HandleFunc("", h.listUsers).Methods("GET")
	adminUserRouter.HandleFunc("/{id}", h.deleteUser).Methods("DELETE")

	// Tokens
	authRouter.HandleFunc("/tokens/renew", h.renewAccessToken).Methods("POST")
	authRouter.HandleFunc("/tokens/revoke", h.revokeSession).Methods("POST")

	return r
}

func Start(addr string) error {
	return http.ListenAndServe(addr, r)
}

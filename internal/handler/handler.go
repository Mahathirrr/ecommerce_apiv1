package handler

import (
	"context"
	"ecom_apiv1/internal/server"
	"ecom_apiv1/internal/storer"
	"ecom_apiv1/token"
	"ecom_apiv1/util"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type handler struct {
	Ctx        context.Context
	server     *server.Server
	TokenMaker *token.JWTMaker
}

func NewHandler(server *server.Server, secretKey string) *handler {
	return &handler{
		Ctx:        context.Background(),
		server:     server,
		TokenMaker: token.NewJWTMaker(secretKey),
	}
}

func (h *handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var productReq ProductReq
	err := json.NewDecoder(r.Body).Decode(&productReq)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusInternalServerError)
		return
	}
	p, err := h.server.CreateProduct(h.Ctx, toStorerProduct(productReq))
	if err != nil {
		http.Error(w, "error creating product", http.StatusInternalServerError)
		return
	}
	res := toProductRes(p)

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idString := vars["id"]
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Error parsing id", http.StatusBadRequest)
		return
	}
	p, err := h.server.GetProduct(h.Ctx, id)
	if err != nil {
		http.Error(w, "Error get product", http.StatusInternalServerError)
		return
	}
	res := toProductRes(p)

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) Listproducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.server.ListProducts(h.Ctx)
	if err != nil {
		http.Error(w, "error get list product", http.StatusInternalServerError)
		return
	}
	var res []ProductRes
	for _, product := range products {
		res = append(res, toProductRes(&product))
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) updateProducts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idString := vars["id"]
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Error parsing id", http.StatusBadRequest)
		return
	}
	var productReq ProductReq
	err = json.NewDecoder(r.Body).Decode(&productReq)
	if err != nil {
		http.Error(w, "error decoding body", http.StatusBadRequest)
		return
	}

	p, err := h.server.GetProduct(h.Ctx, id)
	if err != nil {
		http.Error(w, "Error get product", http.StatusInternalServerError)
		return
	}

	patchProductReq(p, productReq)

	updatedProduct, err := h.server.UpdateProduct(h.Ctx, p)
	if err != nil {
		http.Error(w, "error update product", http.StatusInternalServerError)
		return
	}
	res := toProductRes(updatedProduct)

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idString := vars["id"]
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Error parsing id", http.StatusBadRequest)
		return
	}
	err = h.server.DeleteProduct(h.Ctx, id)
	if err != nil {
		http.Error(w, "error deleting product", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) createOrder(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(authKey{}).(*token.UserClaims)
	var orderReq OrderReq
	err := json.NewDecoder(r.Body).Decode(&orderReq)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	so := toStorerOrder(orderReq)
	so.UserId = claims.ID

	created, err := h.server.CreateOrder(h.Ctx, so)
	if err != nil {
		http.Error(w, "error creating order", http.StatusInternalServerError)
		return
	}
	res := toOrderRes(created)

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) getOrder(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	o, err := h.server.GetOrder(h.Ctx, claims.ID)
	if err != nil {
		http.Error(w, "error getting order", http.StatusInternalServerError)
		return
	}
	res := toOrderRes(o)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) listOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.server.ListOrders(h.Ctx)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var res []OrderRes
	for _, o := range orders {
		res = append(res, toOrderRes(&o))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *handler) deleteOrder(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(authKey{}).(*token.UserClaims)
	err := h.server.DeleteOrder(h.Ctx, claims.ID)
	if err != nil {
		http.Error(w, "error deleting order", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) createUser(w http.ResponseWriter, r *http.Request) {
	var userReq UserReq
	err := json.NewDecoder(r.Body).Decode(&userReq)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusInternalServerError)
		return
	}

	hashedPass, err := util.HashPassword(userReq.Password)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	userReq.Password = hashedPass

	created, err := h.server.CreateUser(h.Ctx, toStorerUser(userReq))
	if err != nil {
		http.Error(w, "error creating user", http.StatusInternalServerError)
		return
	}

	res := toUserRes(created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.server.ListUsers(h.Ctx)
	if err != nil {
		http.Error(w, "error getting list users", http.StatusInternalServerError)
		return
	}

	// var res []UserRes
	// for _, user := range users {
	// 	res = append(res, toUserRes(&user))
	// }

	var res ListUserRes
	for _, u := range users {
		res.Users = append(res.Users, toUserRes(&u))
	}

	w.Header().Set("Content-Type", "application-json")
	json.NewEncoder(w).Encode(res)
}

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(authKey{}).(*token.UserClaims)
	var userReq UserReq
	err := json.NewDecoder(r.Body).Decode(&userReq)
	if err != nil {
		http.Error(w, "error decoding request body", http.StatusInternalServerError)
		return
	}
	u, err := h.server.GetUser(h.Ctx, claims.Email)
	if err != nil {
		http.Error(w, "error get user", http.StatusInternalServerError)
		return
	}

	// Patch our user request
	patchUserReq(userReq, u)
	if u.Email == "" {
		u.Email = claims.Email
	}

	updated, err := h.server.UpdateUser(h.Ctx, u)
	if err != nil {
		http.Error(w, "error update user", http.StatusInternalServerError)
		return
	}
	res := toUserRes(updated)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idString := vars["id"]
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "error parsing id", http.StatusBadRequest)
		return
	}
	err = h.server.DeleteUser(h.Ctx, id)
	if err != nil {
		http.Error(w, "error deleting order", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func patchUserReq(userReq UserReq, u *storer.User) {
	if userReq.Email != "" {
		u.Email = userReq.Email
	}
	if userReq.Name != "" {
		u.Name = userReq.Name
	}
	if userReq.IsAdmin {
		u.IsAdmin = userReq.IsAdmin
	}
	if userReq.Password != "" {
		hashed, err := util.HashPassword(userReq.Password)
		if err != nil {
			panic(err)
		}
		u.Password = hashed
	}
	u.UpdatedAt = toTimePtr(time.Now())
}

func toStorerUser(u UserReq) *storer.User {
	return &storer.User{
		Name:     u.Name,
		Email:    u.Email,
		IsAdmin:  u.IsAdmin,
		Password: u.Password,
	}
}

func toUserRes(u *storer.User) UserRes {
	return UserRes{
		Name:    u.Name,
		Email:   u.Email,
		IsAdmin: u.IsAdmin,
	}
}

func (h *handler) loginUser(w http.ResponseWriter, r *http.Request) {
	var loginReq LoginUserReq
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}
	u, err := h.server.GetUser(h.Ctx, loginReq.Email)
	if err != nil {
		http.Error(w, "error getting user", http.StatusInternalServerError)
		return
	}
	if err = util.CheckPassword(loginReq.Password, u.Password); err != nil {
		http.Error(w, "wrong password", http.StatusBadRequest)
		return
	}
	log.Printf("User data: ID=%v, Email=%s, IsAdmin=%v", u.ID, u.Email, u.IsAdmin)
	// Create JWT and return it as response
	accessToken, ATclaims, err := h.TokenMaker.CreateToken(u.ID, u.Email, u.IsAdmin, 60*time.Minute)
	if err != nil {
		log.Printf("Error creating accesstoken: %v", err)
		http.Error(w, "error creating token", http.StatusInternalServerError)
		return
	}

	refreshToken, RTclaims, err := h.TokenMaker.CreateToken(u.ID, u.Email, u.IsAdmin, 24*time.Hour)
	if err != nil {
		log.Printf("Error creating refreshtoken: %v", err)
		http.Error(w, "error creating token", http.StatusInternalServerError)
		return
	}
	session, err := h.server.CreateSession(h.Ctx, &storer.Session{
		ID:           RTclaims.RegisteredClaims.ID,
		UserEmail:    u.Email,
		RefreshToken: refreshToken,
		IsRevoked:    false,
		ExpiresAt:    RTclaims.RegisteredClaims.ExpiresAt.Time,
	})
	if err != nil {
		http.Error(w, "error create session", http.StatusInternalServerError)
		return
	}
	res := LoginUserRes{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  ATclaims.RegisteredClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: RTclaims.RegisteredClaims.ExpiresAt.Time,
		User:                  toUserRes(u),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) logoutUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(authKey{}).(*token.UserClaims)
	err := h.server.DeleteSession(h.Ctx, claims.RegisteredClaims.ID)
	if err != nil {
		http.Error(w, "error deleting session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) renewAccessToken(w http.ResponseWriter, r *http.Request) {
	var req RenewAccessTokenReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}
	claims, err := h.TokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "error verifying token", http.StatusUnauthorized)
		return
	}

	session, err := h.server.GetSession(h.Ctx, claims.RegisteredClaims.ID)
	if err != nil {
		http.Error(w, "error get session", http.StatusInternalServerError)
		return
	}

	// 2 Validation
	if session.IsRevoked {
		http.Error(w, "session revoked", http.StatusUnauthorized)
		return
	}
	if session.UserEmail != claims.Email {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// Generate New AccessToken
	accessToken, ATClaims, err := h.TokenMaker.CreateToken(claims.ID, claims.Email, claims.IsAdmin, 60*time.Minute)
	if err != nil {
		http.Error(w, "error creating token", http.StatusInternalServerError)
		return
	}
	res := RenewAccessTokenRes{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: ATClaims.RegisteredClaims.ExpiresAt.Time,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) revokeSession(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(authKey{}).(*token.UserClaims)
	err := h.server.RevokeSession(h.Ctx, claims.RegisteredClaims.ID)
	if err != nil {
		http.Error(w, "error revoke session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func toStorerOrder(o OrderReq) *storer.Order {
	return &storer.Order{
		TaxPrice:      o.TaxPrice,
		TotalPrice:    o.TotalPrice,
		PaymentMethod: o.PaymentMethod,
		ShippingPrice: o.ShippingPrice,
		Items:         toStorerOrderItem(o.Items),
	}
}

func toStorerOrderItem(items []OrderItem) []storer.OrderItem {
	var res []storer.OrderItem
	for _, item := range items {
		res = append(res, storer.OrderItem{
			Image:     item.Image,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  item.Quantity,
			ProductID: item.ProductID,
		})
	}
	return res
}

func toOrderRes(o *storer.Order) OrderRes {
	return OrderRes{
		ID:            o.ID,
		ShippingPrice: o.ShippingPrice,
		PaymentMethod: o.PaymentMethod,
		TotalPrice:    o.TotalPrice,
		TaxPrice:      o.TaxPrice,
		CreatedAt:     o.CreatedAt,
		UpdatedAt:     o.UpdatedAt,
		Items:         toOrderItem(o.Items),
	}
}

func toOrderItem(items []storer.OrderItem) []OrderItem {
	var res []OrderItem
	for _, item := range items {
		res = append(res, OrderItem{
			Name:      item.Name,
			Quantity:  item.Quantity,
			Image:     item.Image,
			Price:     item.Price,
			ProductID: item.ProductID,
		})
	}
	return res
}

func patchProductReq(product *storer.Product, p ProductReq) {
	if p.Name != "" {
		product.Name = p.Name
	}
	if p.Image != "" {
		product.Image = p.Image
	}
	if p.Category != "" {
		product.Category = p.Category
	}
	if p.Description != "" {
		product.Description = p.Description
	}
	if p.Rating != 0 {
		product.Rating = p.Rating
	}
	if p.NumReviews != 0 {
		product.NumReviews = p.NumReviews
	}
	if p.Price != 0 {
		product.Price = p.Price
	}
	if p.CountInStock != 0 {
		product.CountInStock = p.CountInStock
	}
	product.UpdatedAt = toTimePtr(time.Now())
}

func toTimePtr(t time.Time) *time.Time {
	return &t
}

func toStorerProduct(p ProductReq) *storer.Product {
	return &storer.Product{
		Name:         p.Name,
		CountInStock: p.CountInStock,
		Image:        p.Image,
		Category:     p.Category,
		Description:  p.Description,
		Rating:       p.Rating,
		NumReviews:   p.NumReviews,
		Price:        p.Price,
	}
}

func toProductRes(p *storer.Product) ProductRes {
	return ProductRes{
		ID:           p.ID,
		Name:         p.Name,
		Image:        p.Image,
		Category:     p.Category,
		Description:  p.Description,
		Rating:       p.Rating,
		NumReviews:   p.NumReviews,
		Price:        p.Price,
		CountInStock: p.CountInStock,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

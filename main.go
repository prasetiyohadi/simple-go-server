package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

func main() {
	// Initialize chi new router
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Handle root path
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!\n"))
	})

	// Create new route
	r.Route("/products", func(r chi.Router) {
		r.Get("/", ListProducts)

		r.Route("/{productID}", func(r chi.Router) {
			r.Use(ProductCtx)      // Load the *Product on the request context
			r.Get("/", GetProduct) // GET /products/123
		})
	})

	// Create the server
	s := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	err := s.ListenAndServe()
	if err != nil {
		log.Printf("Cannot start server: %s\n", err)
	}
}

// ListProducts lists all products in database
func ListProducts(w http.ResponseWriter, r *http.Request) {
	if err := render.RenderList(w, r, NewProductListResponse(products)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// Define a separate type for context key
type key string

const (
	ctxKey key = "product"
)

// ProductCtx middleware is used to load an Product object from
// the URL parameters passed through as the request.
func ProductCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var product *Product
		var err error

		if productID := chi.URLParam(r, "productID"); productID != "" {
			// Conversion string to integer if ID in Product data model is integer
			// v, e := strconv.Atoi(productID)
			// if e != nil {
			// 	fmt.Printf("Panic: %s", e)
			// }
			// product, err = dbGetProduct(v)
			product, err = dbGetProduct(productID)
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKey, product)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetProduct returns the specific Product.
func GetProduct(w http.ResponseWriter, r *http.Request) {
	product := r.Context().Value(ctxKey).(*Product)

	if err := render.Render(w, r, NewProductResponse(product)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// This is entirely optional, but I wanted to demonstrate how you could easily
// add your own logic to the render.Respond method.
func init() {
	render.Respond = func(w http.ResponseWriter, r *http.Request, v interface{}) {
		if err, ok := v.(error); ok {

			// We set a default error status response code if one hasn't been set.
			if _, ok := r.Context().Value(render.StatusCtxKey).(int); !ok {
				w.WriteHeader(400)
			}

			// We log the error
			fmt.Printf("Logging err: %s\n", err.Error())

			// We change the response to not reveal the actual error message,
			// instead we can transform the message something more friendly or mapped
			// to some code / language, etc.
			render.DefaultResponder(w, r, render.M{"status": "error"})
			return
		}

		render.DefaultResponder(w, r, v)
	}
}

// --
// Request and Response payloads for the REST api.
// --

// ProductResponse is the response payload for the Product data model.
type ProductResponse struct {
	*Product
	Elapsed int64 `json:"elapsed"`
}

// NewProductResponse returns a product
func NewProductResponse(product *Product) *ProductResponse {
	resp := &ProductResponse{Product: product}
	return resp
}

// Render function for ProductResponse to use render.Renderer interface
func (rd *ProductResponse) Render(w http.ResponseWriter, r *http.Request) error {
	rd.Elapsed = 10
	return nil
}

// NewProductListResponse returns list of products
func NewProductListResponse(products []*Product) []render.Renderer {
	list := []render.Renderer{}
	for _, product := range products {
		list = append(list, NewProductResponse(product))
	}
	return list
}

// --
// Error response payloads & renderer
// --

// ErrResponse renderer type for handling all sorts of errors.
type ErrResponse struct {
	Err            error  `json:"-"`               // low-level runtime error
	HTTPStatusCode int    `json:"-"`               // http response status code
	StatusText     string `json:"status"`          // user-level status message
	AppCode        int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText      string `json:"error,omitempty"` // application-level error message, for debugging
}

// Render function for ErrResponse to use render.Renderer interface
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrInvalidRequest renders invalid request error
// func ErrInvalidRequest(err error) render.Renderer {
// 	return &ErrResponse{
// 		Err:            err,
// 		HTTPStatusCode: 400,
// 		StatusText:     "Invalid request.",
// 		ErrorText:      err.Error(),
// 	}
// }

// ErrRender renders rendering error
func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}

// --
// Data model objects and persistence mocks
// --

// Product data model
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	SKU         string  `json:"sku"`
}

// Product fixture data
var products = []*Product{
	{ID: "1", Name: "Latte", Description: "Frothy milky coffe", Price: 2.45, SKU: "abc123"},
	{ID: "2", Name: "Esspresso", Description: "Short and strong coffee without milk", Price: 1.99, SKU: "def456"},
	{ID: "3", Name: "Affogato", Description: "Coffee in the realms of dessert", Price: 3.14, SKU: "ghi789"},
	{ID: "4", Name: "Cappucino", Description: "The gateway into coffee", Price: 2.34, SKU: "jkl135"},
	{ID: "5", Name: "Americano", Description: "Simple coffee topped up with hot water", Price: 2.12, SKU: "mno246"},
}

func dbGetProduct(id string) (*Product, error) {
	for _, p := range products {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, errors.New("product not found")
}

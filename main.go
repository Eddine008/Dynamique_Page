package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
	ImageURL    string
	InStock     int
	DiscountPct int
}

var (
	products  []Product
	nextID    int
	templates *template.Template
)

func init() {
	products = []Product{
		{ID: 1, Name: "Palace Pull à capuche", Price: 146, ImageURL: "/static/img/products/19A.webp", InStock: 5, DiscountPct: 0},
		{ID: 2, Name: "Palace Pull à capuchon marine", Price: 138, ImageURL: "/static/img/products/16A.webp", InStock: 3, DiscountPct: 10},
		{ID: 3, Name: "Palace pull crew", Price: 128, ImageURL: "/static/img/products/18A.webp", InStock: 2, DiscountPct: 0},
		{ID: 4, Name: "Palace washed terry 1/4 Placket hood mojito", Price: 168, ImageURL: "/static/img/products/21A.webp", InStock: 4, DiscountPct: 5},
		{ID: 5, Name: "palace pantalon bossy jean stone", Price: 125, ImageURL: "/static/img/products/22A.webp", InStock: 6, DiscountPct: 0},
		{ID: 5, Name: "palace pantalon cargo gore-tex r-tek noir", Price: 110, ImageURL: "/static/img/products/34B.webp", InStock: 6, DiscountPct: 0},
	}
	nextID = 6
}

func fmtPrice(p float64) string {
	return fmt.Sprintf("%.2f €", p)
}

func discountedPrice(p float64, pct int) float64 {
	if pct <= 0 {
		return p
	}
	return p * (1.0 - float64(pct)/100.0)
}

func findProductByID(id int) (Product, error) {
	for _, p := range products {
		if p.ID == id {
			return p, nil
		}
	}
	return Product{}, errors.New("not found")
}

func addProduct(p Product) int {
	p.ID = nextID
	nextID++
	products = append(products, p)
	return p.ID
}

func mustLoadTemplates() {
	funcMap := template.FuncMap{
		"fmtPrice":        fmtPrice,
		"discountedPrice": discountedPrice,
	}
	var err error
	templates, err = template.New("base").Funcs(funcMap).ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		log.Fatalf("Erreur chargement templates : %v", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("indexHandler appelé pour", r.URL.Path)
	if r.URL.Path != "/" {
		log.Println("indexHandler : page non trouvée")
		http.NotFound(w, r)
		return
	}
	data := map[string]interface{}{"Products": products}
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Println("Erreur execution index.html :", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("productHandler appelé pour", r.URL.Path)
	idStr := r.URL.Path[len("/product/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println("productHandler : ID invalide")
		http.NotFound(w, r)
		return
	}
	p, err := findProductByID(id)
	if err != nil {
		log.Println("productHandler : produit non trouvé")
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{"Product": p}
	if err := templates.ExecuteTemplate(w, "detail.html", data); err != nil {
		log.Println("Erreur execution detail.html :", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("addHandler appelé pour", r.Method)
	switch r.Method {
	case http.MethodGet:
		log.Println("Execution de add.html")
		if err := templates.ExecuteTemplate(w, "add.html", nil); err != nil {
			log.Println("Erreur execution add.html :", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			log.Println("Erreur ParseForm :", err)
			http.Error(w, "formulaire invalide", http.StatusBadRequest)
			return
		}

		name := r.PostForm.Get("name")
		desc := r.PostForm.Get("description")
		priceStr := r.PostForm.Get("price")
		stockStr := r.PostForm.Get("stock")
		discountStr := r.PostForm.Get("discount")

		price, _ := strconv.ParseFloat(priceStr, 64)
		stock, _ := strconv.Atoi(stockStr)
		discount := 0
		if discountStr != "" {
			discount, _ = strconv.Atoi(discountStr)
		}

		p := Product{
			Name:        name,
			Description: desc,
			Price:       price,
			ImageURL:    "/static/img/products/16A.webp",
			InStock:     stock,
			DiscountPct: discount,
		}
		newID := addProduct(p)
		log.Println("Produit ajouté avec ID :", newID)
		http.Redirect(w, r, fmt.Sprintf("/product/%d", newID), http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	mustLoadTemplates()
	fmt.Println("Templates chargés :", templates.DefinedTemplates())

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/product/", productHandler)
	http.HandleFunc("/add", addHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Serveur démarré sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

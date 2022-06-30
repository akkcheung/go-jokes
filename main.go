package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"
)

var templates *template.Template
var DB *gorm.DB

var (
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

func main() {

	r := mux.NewRouter()

	//db, err := gorm.Open(sqlite.Open("my-jokes.db"), &gorm.Config{})
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}
	//defer db.Close()
	DB = db

	db.AutoMigrate(&Joke{}, &User{})

	fs := http.FileServer(http.Dir("assets/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	templates = template.Must(template.ParseGlob("views/*.html"))

	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
	r.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Helo you've requested: %s\n", r.URL.Path)
	})

	r.HandleFunc("/", indexHandler).Methods("GET")

	r.HandleFunc("/signup", signUpHandler)

	r.HandleFunc("/signin", signInHandler)

	r.HandleFunc("/signout", signOutHandler)

	r.HandleFunc("/jokes", jokesGetHandler).Methods("GET")

	r.HandleFunc("/jokes/{id:[0-9]+}", jokeGetHandler).Methods("GET")

	r.HandleFunc("/jokes/random", jokeGetRandomHandler).Methods("GET")

	//below routes need authentication

	r.HandleFunc("/jokes/new", authMiddleware(jokeGetNewHandler)).Methods("GET")

	r.HandleFunc("/jokes", authMiddleware(jokePostHandler)).Methods("POST")
	r.HandleFunc("/jokes/{id}", authMiddleware(jokeDeleteHandler)).Methods("POST")

	//http.ListenAndServe(":5000", nil)
	port := os.Getenv("PORT")

	if port == "" {
		port = "5000"
	}

	http.ListenAndServe(":"+port, r)
}

type Joke struct {
	gorm.Model
	Name    string `json:"name"`
	Content string `json:"content"`
	UserID  uint   //foreign key
}

type User struct {
	gorm.Model
	Login    string `json:"login"`
	Password string `json:"password gorm:type:"varchar(100)"`
	Jokes    []Joke
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Helo you've requested: %s\n", r.URL.Path)
}

func signInHandler(w http.ResponseWriter, r *http.Request) {

	data := map[string]interface{}{}
	data["Jokes"] = nil
	data["Login"] = nil
	data["Password"] = nil

	data["Success"] = false
	data["Errors"] = nil

	var user User
	var tmpl *template.Template
	var err error

	var errors map[string]string

	errors = make(map[string]string)

	if r.Method == http.MethodPost {

		session, _ := store.Get(r, "cookie-name")

		r.ParseForm()

		//debug
		//fmt.Println("login:", r.FormValue("login"))
		//fmt.Println("password:", r.FormValue("password"))

		login := r.FormValue("login")
		password := r.FormValue("password")

		if login == "" || password == "" || len(password) < 6 {

			if login == "" {
				errors["login"] = "Empty login "
			}

			if password == "" {
				errors["password"] = "Empty password "
			}

			if len(password) < 6 {
				errors["password"] = "Password too short "
			}

			data["Errors"] = errors
			data["Login"] = login
			data["Password"] = password

		} else {
			//result := DB.Where(&User{Login: login, Password: ""}).First(&user )
			result := DB.First(&user, "login = ?", login)
			err = result.Error
			if err != nil {
				//fmt.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//log.Println("login success\n")
			data["Success"] = true

			session.Values["authenticated"] = true
			session.Values["login"] = login
			session.Save(r, w)

			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/jokes")
			} else {
				http.Redirect(w, r, "/jokes", http.StatusFound)
			}

			return
		}
	}

	//method get or post with error

	if r.Header.Get("HX-Request") == "true" {
		tmpl, err = template.ParseFiles("views/login.html")
	} else {
		tmpl, err = template.ParseFiles("views/base.html", "views/header.html", "views/sidebar.html", "views/jokes.html", "views/login.html")
	}

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		err = tmpl.ExecuteTemplate(w, "content", data)
	} else {
		err = tmpl.ExecuteTemplate(w, "base", data)
	}

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func signOutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	session.Values["authenticated"] = false
	session.Save(r, w)

	tmpl, err := template.ParseFiles("views/signout.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	tmpl.Execute(w, nil)

}

func signUpHandler(w http.ResponseWriter, r *http.Request) {

	var tmpl *template.Template
	var err error
	var errors map[string]string

	data := map[string]interface{}{}

	data["Jokes"] = nil
	data["Login"] = ""

	data["Errors"] = nil
	data["Success"] = false

	errors = make(map[string]string)

	session, _ := store.Get(r, "cookie-name")

	if r.Method == http.MethodPost {
		r.ParseForm()

		//debug
		//fmt.Println("login:", r.FormValue("login"))
		//fmt.Println("password:", r.FormValue("password"))
		//fmt.Println("repeatPassword:", r.FormValue("repeatPassword"))

		login := r.FormValue("login")
		password := r.FormValue("password")
		repeatPassword := r.FormValue("repeatPassword")

		if login == "" || password != repeatPassword || len(password) < 6 {

			if login == "" {
				errors["login"] = "Empty login "
			}

			if password == "" {
				errors["password"] = "Empty password "
			}

			if len(password) < 6 {
				errors["password"] = "Password too short "
			}

			if repeatPassword != password {
				errors["repeatPassword"] = "Repeat password not match "
			}

			data["Success"] = false
			data["Login"] = r.FormValue("login")
			data["Errors"] = errors

		} else {

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), 8)
			if err != nil {
				fmt.Println(err)
			}

			user := User{Login: r.FormValue("login"), Password: string(hashedPassword)}
			result := DB.Create(&user)

			err = result.Error
			if err != nil {
				fmt.Println(err)
			}

			data["Success"] = true

			session.Values["authenticated"] = true
			session.Values["login"] = user.Login
			session.Save(r, w)
		}
	}

	if r.Header.Get("HX-Request") == "true" {
		tmpl, err = template.ParseFiles("views/signup.html")
	} else {

		tmpl, err = template.ParseFiles("views/base.html", "views/header.html", "views/sidebar.html", "views/signup.html")
	}

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if r.Header.Get("HX-Request") == "true" {
		if r.Method == http.MethodGet {
			err = tmpl.ExecuteTemplate(w, "content", data)
		} else {
			if len(errors) == 0 {
				w.Header().Set("HX-Redirect", "/jokes")
				return
			} else {
				err = tmpl.ExecuteTemplate(w, "content", data)
			}
		}

	} else {
		err = tmpl.ExecuteTemplate(w, "base", data)
	}

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func jokesGetHandler(w http.ResponseWriter, r *http.Request) {

	data := map[string]interface{}{}

	var (
		jokes       []Joke
		currentJoke Joke

		err  error
		tmpl *template.Template
	)

	result := DB.Find(&jokes)
	err = result.Error

	if err != nil {
		fmt.Println(err)
	}

	result2 := DB.First(&currentJoke)
	err = result2.Error
	if err != nil {
		fmt.Println(err)
	}

	data["PageTitle"] = "Page Title"
	data["Jokes"] = jokes
	data["CurrentJoke"] = currentJoke
	data["IsSignIn"] = isSignIn(r)
	data["Login"] = getLogin(r)

	if r.Header.Get("HX-Request") == "true" {
		tmpl, err = template.ParseFiles("views/sidebar.html", "views/jokes.html")
	} else {
		tmpl, err = template.ParseFiles("views/base.html", "views/header.html", "views/sidebar.html", "views/jokes.html", "views/show.html")
	}

	if r.Header.Get("HX-Request") == "true" {
		err = tmpl.ExecuteTemplate(w, "jokes", data)
	} else {
		err = tmpl.ExecuteTemplate(w, "base", data)
	}

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func jokePostHandler(w http.ResponseWriter, r *http.Request) {

	var (
		name    string
		content string
		errors  map[string]string

		err  error
		tmpl *template.Template
	)

	data := map[string]interface{}{}
	errors = make(map[string]string)

	r.ParseForm()
	//fmt.Println("name:", r.FormValue("name"))
	//fmt.Println("content:", r.FormValue("content"))
	name = r.FormValue("name")
	content = r.FormValue("content")

	if name == "" {
		errors["name"] = "Empty Name "
	}

	if content == "" {
		errors["content"] = "Empty Content "
	}

	data["Success"] = false
	data["Errors"] = errors

	if len(errors) != 0 {
		if r.Header.Get("HX-Request") == "true" {
			tmpl, err = template.ParseFiles("views/new.html")
		} else {
			tmpl, err = template.ParseFiles("views/base.html", "views/header.html", "views/sidebar.html", "views/new.html")
		}

		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if r.Header.Get("HX-Request") == "true" {
			err = tmpl.ExecuteTemplate(w, "content", data)
		} else {
			err = tmpl.ExecuteTemplate(w, "base", data)
		}

	} else {

		joke := Joke{Name: r.FormValue("name"), Content: r.FormValue("content")}
		result := DB.Create(&joke)

		lastInsertId := joke.ID

		err := result.Error
		if err != nil {
			fmt.Println(err)
		}

		data["LastInsertId"] = lastInsertId
		data["Success"] = true

		//w.Header().Set("HX-Trigger", "newJoke") // will convert to Canonical case
		w.Header()["HX-Trigger"] = []string{"newJoke"}

		//debug
		//fmt.Println("lastInsertId:" + strconv.Itoa(int(lastInsertId)))

		if r.Header.Get("HX-Request") == "true" {
			tmpl, err = template.ParseFiles("views/new.html")
		} else {
			tmpl, err = template.ParseFiles("views/base.html", "views/header.html", "views/sidebar.html", "views/jokes.html", "views/show.html")
		}

		if r.Header.Get("HX-Request") == "true" {
			err = tmpl.ExecuteTemplate(w, "content", data)
		} else {
			err = tmpl.ExecuteTemplate(w, "base", data)
		}

		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func jokeGetHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]

	//debug
	//fmt.Printf("%s\n", id)

	data := map[string]interface{}{}

	var jokes []Joke
	var currentJoke Joke
	var tmpl *template.Template

	result := DB.Find(&jokes)
	err := result.Error

	if err != nil {
		fmt.Println(err)
	}

	data["PageTitle"] = "Page Title"
	data["Jokes"] = jokes
	data["CurrentJoke"] = nil
	data["IsSignIn"] = isSignIn(r)

	var s uint64

	if id == "random" {
		var random_id int
		DB.Raw("select id from jokes where deleted_at is null order by random() limit 1").Scan(&random_id)

		//debug
		//fmt.Printf("random_id : %v\n", random_id)

		s = uint64(random_id)
	} else {
		s, err = strconv.ParseUint(id, 10, 64)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	result = DB.First(&currentJoke, s)

	err = result.Error
	if err != nil {
		fmt.Println(err)
	}

	//debug
	//fmt.Printf("%v\n", s)

	data["CurrentJoke"] = currentJoke

	//debug
	//log.Println("header : " + r.Header.Get("HX-Request"))

	if r.Header.Get("HX-Request") == "true" {
		tmpl, err = template.ParseFiles("views/show.html")
		err = tmpl.ExecuteTemplate(w, "content", data)

	} else {
		tmpl, err = template.ParseFiles("views/base.html", "views/header.html", "views/sidebar.html", "views/show.html")

		err = tmpl.ExecuteTemplate(w, "base", data)
	}

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func jokeGetRandomHandler(w http.ResponseWriter, r *http.Request) {

	var random_id int
	DB.Raw("select id from jokes where deleted_at is null order by random() limit 1").Scan(&random_id)

	//debug
	//fmt.Printf("random_id : %v\n", random_id)

	http.Redirect(w, r, "/jokes/"+strconv.Itoa(random_id), http.StatusSeeOther)

}

func jokeGetNewHandler(w http.ResponseWriter, r *http.Request) {

	var (
		jokes []Joke
		tmpl  *template.Template
		err   error
	)

	data := map[string]interface{}{}

	result := DB.Find(&jokes)
	err = result.Error

	if err != nil {
		fmt.Println(err)
	}

	data["PageTitle"] = "Page Title"
	data["Jokes"] = jokes
	data["CurrentJoke"] = nil
	data["IsSignIn"] = isSignIn(r)
	data["login"] = getLogin(r)

	if r.Header.Get("HX-Request") == "true" {
		tmpl, err = template.ParseFiles("views/new.html")
	} else {
		tmpl, err = template.ParseFiles("views/base.html", "views/header.html", "views/sidebar.html", "views/new.html")
	}

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		err = tmpl.ExecuteTemplate(w, "content", data)
	} else {
		err = tmpl.ExecuteTemplate(w, "base", data)
	}

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func jokeDeleteHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]

	r.ParseForm()

	//debug
	//fmt.Println("_method:", r.FormValue("_method"))

	method := r.FormValue("_method")
	if method == "delete" {

		s, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		result := DB.Delete(&Joke{}, s) //soft delete

		err = result.Error
		if err != nil {
			fmt.Println(err)
		}

		//data["Success"] = true
		w.Header()["HX-Trigger"] = []string{"deleteJoke"}

		if r.Header.Get("HX-Request") == "true" {
			w.Header()["HX-Trigger"] = []string{"deleteJoke"}
			fmt.Fprintf(w, "<p style='color:blue'>Record is removed.</p>")
		} else {
			http.Redirect(w, r, "/jokes", http.StatusSeeOther)
		}

	}

}

func authMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if !isSignIn(r) {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/signout")
				return
			}

			http.Redirect(w, r, "/signout", http.StatusFound)
		}

		f(w, r)
	}
}

func isSignIn(r *http.Request) bool {
	session, _ := store.Get(r, "cookie-name")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		return false
	}
	return true
}

func getLogin(r *http.Request) string {
	session, _ := store.Get(r, "cookie-name")
	if login, ok := session.Values["login"].(string); !ok || login == "" {
		return ""
	} else {
		return login
	}
	return ""
}

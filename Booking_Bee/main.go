package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var tpl *template.Template
var tpl2 *template.Template
var db *sql.DB

type MoviesStruct struct {
	MovieID      int
	MovieName    string
	MovieDesc    string
	MoviePrice   float64
	CinemaHallID int
	PremierDate  string
	ImageUrl     string
}

type CinemaStruct struct {
	CinemaID       int
	CinemaName     string
	CinemaLocation string
	CinemaCapacity int
}

type DashboardData struct {
	Movies []MoviesStruct
	Cinema []CinemaStruct
}

func main() {
	tpl, _ = template.ParseGlob("views/*.html")
	tpl2, _ = template.ParseGlob("views/screens/*.html")
	fs := http.FileServer(http.Dir("assets"))
	var err error
	db, err = sql.Open("mysql", "owen:ichimaruGin@tcp(localhost:3306)/Booking_Bee")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	http.Handle("/assets/", http.StripPrefix("/assets", fs))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/dashboard", dashHandler)
	http.HandleFunc("/payment", payHandler)
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "home.html", nil)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Parse form data
		fname := r.FormValue("registerFname")
		lname := r.FormValue("registerLname")
		email := r.FormValue("registerEmail")
		password := r.FormValue("registerPassword")

		// check email availability
		stmt := "select count(*) from Customers where Customers_email = ?"
		row := db.QueryRow(stmt, email)
		var count int
		err := row.Scan(&count)
		if err != sql.ErrNoRows {
			fmt.Println("Email exists, error: ", err)
			tpl.ExecuteTemplate(w, "signup.html", "Email already taken, try again!!")
			return
		}

		// Hash password
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println("bcrypt err: ", err)
			tpl.ExecuteTemplate(w, "signup.html", "There was a problem registering account")
			return
		}

		insert, err := db.Prepare("INSERT INTO Customers(Customers_fname, Customers_lname, Customers_email, Customers_password) VALUES(?, ?, ?, ?)")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		insert.Exec(fname, lname, email, string(hash))

		fmt.Fprint(w, "<div class=\"container\"><h2 class=\"mbr-section-title mb-0 display-1\"><strong>congrats, your account has been created successfully.<br>You may proceed to LOG IN</strong></h2></div> ")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		tpl.ExecuteTemplate(w, "signup.html", nil)
	}

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tpl.ExecuteTemplate(w, "login.html", nil)
	} else {
		email := r.FormValue("loginName")
		password := r.FormValue("loginPassword")

		var dbEmail string
		var dbPassword string
		err := db.QueryRow("SELECT Customers_email, Customers_password FROM Customers WHERE Customers_email=?", email).Scan(&dbEmail, &dbPassword)
		if err != nil {
			fmt.Println(err)
			tpl.ExecuteTemplate(w, "login.html", "There was a problem logging in, please try again!!")
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password))
		if err != nil {
			tpl.ExecuteTemplate(w, "login.html", "Incorrect email or password, please try again!!")
			return
		}

		// Login successful, redirect to dashboard page
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func dashHandler(w http.ResponseWriter, r *http.Request) {
	//SELECTING MOVIE ROWS
	movieRows, err := db.Query("SELECT * FROM Movies")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer movieRows.Close()

	// Process rows into a slice of struct or map
	var movieDetails []MoviesStruct
	for movieRows.Next() {
		var movie = MoviesStruct{}
		err := movieRows.Scan(&movie.MovieID, &movie.MovieName, &movie.MovieDesc, &movie.MoviePrice, &movie.CinemaHallID, &movie.PremierDate, &movie.ImageUrl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		movieDetails = append(movieDetails, movie)
	}
	if err = movieRows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//SELECTING CINEMA ROWS
	cinemaRows, err := db.Query("SELECT * FROM Cinema_halls")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cinemaRows.Close()

	// Process rows into a slice of struct or map
	var cinemaDetails []CinemaStruct
	for cinemaRows.Next() {
		var cinema = CinemaStruct{}
		err := cinemaRows.Scan(&cinema.CinemaID, &cinema.CinemaName, &cinema.CinemaLocation, &cinema.CinemaCapacity)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cinemaDetails = append(cinemaDetails, cinema)

	}
	if err = cinemaRows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	DashData := DashboardData{Movies: movieDetails, Cinema: cinemaDetails}

	tpl2, err := template.ParseFiles("views/screens/dashboard.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tpl2.Execute(w, DashData)

}

func payHandler(w http.ResponseWriter, r *http.Request) {
	tpl2.ExecuteTemplate(w, "payment.html", nil)
}

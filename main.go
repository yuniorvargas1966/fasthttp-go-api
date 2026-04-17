package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttprouter"
)

var db *sql.DB

// ==========================
// MODELO
// ==========================
type Servicio struct {
	ID          int    `json:"id"`
	Nombre      string `json:"nombre"`
	Correo      string `json:"correo"`
	Telefono    string `json:"telefono"`
	Equipo      string `json:"equipo"`
	Diagnostico string `json:"diagnostico"`
	Resultados  string `json:"resultados"`
	Decision    string `json:"decision"`
	Taller      string `json:"taller"`
	Servicio    string `json:"servicio"`
	Entrega     string `json:"entrega"`
	Fecha       string `json:"fecha"`
}

// ==========================
// INIT DB (UNA SOLA VEZ)
// ==========================
func initDB() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	dsn := os.Getenv("Usuario") + ":" +
		os.Getenv("Contrasena") +
		"@tcp(" + os.Getenv("Host") + os.Getenv(":Puerto") + os.Getenv("DBPort") + ")/" +
		os.Getenv("Nombre") + "?parseTime=false"

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// Pool de conexiones (CLAVE para rendimiento)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

// ==========================
// UTILIDADES
// ==========================
func jsonResponse(ctx *fasthttp.RequestCtx, status int, data interface{}) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)

	resp, err := json.Marshal(data)
	if err != nil {
		ctx.Error("Error procesando JSON", 500)
		return
	}

	ctx.SetBody(resp)
}

func enableCORS(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if string(ctx.Method()) == "OPTIONS" {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
	}
}

// ==========================
// HANDLERS
// ==========================
func getServicios(ctx *fasthttp.RequestCtx, _ fasthttprouter.Params) {
	enableCORS(ctx)

	rows, err := db.Query("SELECT id, nombre, correo, telefono, equipo, diagnostico, resultados, decision, taller, servicio, entrega, fecha FROM taller")
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}
	defer rows.Close()

	var servicios []Servicio

	for rows.Next() {
		var s Servicio
		if err := rows.Scan(
			&s.ID, &s.Nombre, &s.Correo, &s.Telefono,
			&s.Equipo, &s.Diagnostico, &s.Resultados,
			&s.Decision, &s.Taller, &s.Servicio,
			&s.Entrega, &s.Fecha,
		); err != nil {
			ctx.Error(err.Error(), 500)
			return
		}
		servicios = append(servicios, s)
	}

	jsonResponse(ctx, 200, servicios)
}

func getServicio(ctx *fasthttp.RequestCtx, ps fasthttprouter.Params) {
	enableCORS(ctx)

	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ctx.Error("ID inválido", 400)
		return
	}

	var s Servicio
	err = db.QueryRow("SELECT id, nombre, correo, telefono, equipo, diagnostico, resultados, decision, taller, servicio, entrega, fecha FROM taller WHERE id=?", id).
		Scan(
			&s.ID, &s.Nombre, &s.Correo, &s.Telefono,
			&s.Equipo, &s.Diagnostico, &s.Resultados,
			&s.Decision, &s.Taller, &s.Servicio,
			&s.Entrega, &s.Fecha,
		)

	if err == sql.ErrNoRows {
		ctx.Error("No encontrado", 404)
		return
	} else if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	jsonResponse(ctx, 200, s)
}

func crearServicio(ctx *fasthttp.RequestCtx, _ fasthttprouter.Params) {
	enableCORS(ctx)

	var s Servicio
	if err := json.Unmarshal(ctx.PostBody(), &s); err != nil {
		ctx.Error("JSON inválido", 400)
		return
	}

	result, err := db.Exec(
		"INSERT INTO taller (nombre, correo, telefono, equipo, diagnostico, resultados, decision, taller, servicio, entrega, fecha) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		s.Nombre, s.Correo, s.Telefono, s.Equipo,
		s.Diagnostico, s.Resultados, s.Decision,
		s.Taller, s.Servicio, s.Entrega, s.Fecha,
	)

	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	id, _ := result.LastInsertId()
	s.ID = int(id)

	jsonResponse(ctx, 201, s)
}

func actualizarServicio(ctx *fasthttp.RequestCtx, ps fasthttprouter.Params) {
	enableCORS(ctx)

	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ctx.Error("ID inválido", 400)
		return
	}

	var s Servicio
	if err := json.Unmarshal(ctx.PostBody(), &s); err != nil {
		ctx.Error("JSON inválido", 400)
		return
	}

	_, err = db.Exec(
		"UPDATE taller SET nombre=?, correo=?, telefono=?, equipo=?, diagnostico=?, resultados=?, decision=?, taller=?, servicio=?, entrega=?, fecha=? WHERE id=?",
		s.Nombre, s.Correo, s.Telefono, s.Equipo,
		s.Diagnostico, s.Resultados, s.Decision,
		s.Taller, s.Servicio, s.Entrega, s.Fecha, id,
	)

	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	s.ID = id
	jsonResponse(ctx, 200, s)
}

func eliminarServicio(ctx *fasthttp.RequestCtx, ps fasthttprouter.Params) {
	enableCORS(ctx)

	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ctx.Error("ID inválido", 400)
		return
	}

	_, err = db.Exec("DELETE FROM taller WHERE id=?", id)
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

// ==========================
// MAIN
// ==========================
func main() {
	initDB()
	defer db.Close()

	port := os.Getenv("Port")
	if port == "" {
		port = "Port"
	}

	router := fasthttprouter.New()

	router.GET("/servicios", getServicios)
	router.GET("/servicios/:id", getServicio)
	router.POST("/servicios", crearServicio)
	router.PUT("/servicios/:id", actualizarServicio)
	router.DELETE("/servicios/:id", eliminarServicio)

	log.Println("Servidor en http://localhost" + port + "/servicios")
	log.Fatal(fasthttp.ListenAndServe("0.0.0.0"+port, router.Handler))
}

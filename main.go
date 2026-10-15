package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"registro-despesas/config"
	"registro-despesas/internal/database"
	"registro-despesas/internal/handler"
	"registro-despesas/internal/middleware"
	"registro-despesas/internal/repository"
	"registro-despesas/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("banco: %v", err)
	}
	defer db.Close()

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	// Sem proxy confiável, ClientIP ignora X-Forwarded-For (evita burlar o rate limit).
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Fatalf("proxies: %v", err)
	}
	r.Use(
		middleware.SecurityHeaders(),
		middleware.CORS(cfg.CORSOrigin),
		middleware.NewRateLimiter(60).Middleware(),
	)

	// Redirecionamentos sempre 302 (301 é cacheado pelo navegador e causa loops).
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/index.html")
	})

	authSvc, err := service.NewAuthService(repository.NewAuthRepository(db), cfg.JWTSecret)
	if err != nil {
		log.Fatalf("auth: %v", err)
	}
	authHandler := handler.NewAuthHandler(authSvc)

	api := r.Group("/api")
	api.POST("/auth/login", middleware.NewRateLimiter(10).Middleware(), authHandler.Login)

	protected := api.Group("", middleware.Auth(cfg.JWTSecret))
	protected.POST("/auth/alterar-senha", authHandler.AlterarSenha)

	// Rotas de escrita (e consultas restritas) exigem ADMIN; READ recebe 403.
	admin := protected.Group("", middleware.RequireAdmin())

	categoriaHandler := handler.NewCategoriaHandler(service.NewCategoriaService(repository.NewCategoriaRepository(db)))
	protected.GET("/categorias", categoriaHandler.Listar)
	admin.POST("/categorias", categoriaHandler.Salvar)
	admin.DELETE("/categorias/:id", categoriaHandler.Excluir)

	usuarioHandler := handler.NewUsuarioHandler(service.NewUsuarioService(repository.NewUsuarioRepository(db)))
	admin.GET("/usuarios", usuarioHandler.Listar)
	admin.POST("/usuarios", usuarioHandler.Salvar)
	admin.DELETE("/usuarios/:id", usuarioHandler.Excluir)

	despesaHandler := handler.NewDespesaHandler(service.NewDespesaService(repository.NewDespesaRepository(db)))
	admin.POST("/despesas", despesaHandler.Salvar)
	admin.DELETE("/despesas/:id", despesaHandler.Excluir)

	r.Static("/css", "./public/css")
	r.Static("/js", "./public/js")
	for _, page := range []string{"login", "index", "categoria", "despesa", "usuario"} {
		r.GET("/"+page+".html", servePage("public/"+page+".html"))
	}

	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("servidor em http://localhost:%s", cfg.ServerPort)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("servidor: %v", err)
	}
}

// servePage usa ServeContent porque ServeFile redireciona /index.html com 301.
func servePage(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		f, err := os.Open(path)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Header("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(c.Writer, c.Request, "", fi.ModTime(), f)
	}
}

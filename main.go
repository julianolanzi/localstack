package main

import (
	"fmt"
	"log"

	"localstackdemo/config"
	"localstackdemo/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Configurar AWS
	cfg, err := config.GetAWSConfig()
	if err != nil {
		log.Fatalf("Erro ao carregar configuração AWS: %v", err)
	}

	// Configurar Gin
	r := gin.Default()

	// Configurar rotas
	routes.SetupRoutes(r, cfg)

	// Iniciar servidor
	fmt.Println("Servidor rodando na porta 6000")
	r.Run(":6000")
}

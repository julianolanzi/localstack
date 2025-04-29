package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/gin-gonic/gin"
)

type LambdaController struct {
	client *lambda.Client
}

func NewLambdaController(cfg aws.Config) *LambdaController {
	client := lambda.NewFromConfig(cfg)
	return &LambdaController{
		client: client,
	}
}

type CreateFunctionRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (l *LambdaController) CreateFunction(c *gin.Context) {
	var req CreateFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nome da função é obrigatório"})
		return
	}

	// Ler o arquivo ZIP da função
	zipFile, err := os.ReadFile("lambda/function.zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao ler arquivo ZIP: %v", err)})
		return
	}

	// Criar a função Lambda
	createFunctionOutput, err := l.client.CreateFunction(context.TODO(), &lambda.CreateFunctionInput{
		FunctionName: aws.String(req.Name),
		Description:  aws.String(req.Description),
		Runtime:      types.RuntimeGo1x,
		Handler:      aws.String("main"),
		Role:         aws.String("arn:aws:iam::000000000000:role/lambda-role"),
		Code: &types.FunctionCode{
			ZipFile: zipFile,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao criar função: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Função criada com sucesso",
		"arn":     *createFunctionOutput.FunctionArn,
	})
}

func (l *LambdaController) InvokeFunction(c *gin.Context) {
	functionName := c.Param("name")
	if functionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nome da função é obrigatório"})
		return
	}

	// Verificar o estado da função
	var functionState string
	for i := 0; i < 5; i++ {
		getFunctionOutput, err := l.client.GetFunction(context.TODO(), &lambda.GetFunctionInput{
			FunctionName: aws.String(functionName),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao verificar estado da função: %v", err)})
			return
		}

		functionState = string(getFunctionOutput.Configuration.State)
		if functionState == string(types.StateActive) {
			break
		}

		// Esperar 1 segundo antes de tentar novamente
		time.Sleep(time.Second)
	}

	if functionState != string(types.StateActive) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Função ainda não está ativa"})
		return
	}

	// Invocar a função
	invokeOutput, err := l.client.Invoke(context.TODO(), &lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		Payload:      []byte(`{"message": "Hello from API Gateway!"}`),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao invocar função: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  invokeOutput.StatusCode,
		"payload": string(invokeOutput.Payload),
	})
}

func (l *LambdaController) ListFunctions(c *gin.Context) {
	result, err := l.client.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao listar funções: %v", err)})
		return
	}

	functions := make([]map[string]string, 0)
	for _, function := range result.Functions {
		functions = append(functions, map[string]string{
			"name":        *function.FunctionName,
			"description": *function.Description,
			"arn":         *function.FunctionArn,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"functions": functions,
	})
}

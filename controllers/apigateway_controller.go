package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/gin-gonic/gin"
)

type APIGatewayController struct {
	client *apigateway.Client
}

func NewAPIGatewayController(cfg aws.Config) *APIGatewayController {
	client := apigateway.NewFromConfig(cfg)
	return &APIGatewayController{
		client: client,
	}
}

type CreateAPIRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (a *APIGatewayController) CreateAPI(c *gin.Context) {
	var req CreateAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nome da API é obrigatório"})
		return
	}

	// Criar a API
	createAPIOutput, err := a.client.CreateRestApi(context.TODO(), &apigateway.CreateRestApiInput{
		Name:        aws.String(req.Name),
		Description: aws.String(req.Description),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao criar API: %v", err)})
		return
	}

	// Obter o ID do recurso raiz
	rootResourceOutput, err := a.client.GetResources(context.TODO(), &apigateway.GetResourcesInput{
		RestApiId: createAPIOutput.Id,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao obter recursos: %v", err)})
		return
	}

	// Criar um recurso para o endpoint
	resourceOutput, err := a.client.CreateResource(context.TODO(), &apigateway.CreateResourceInput{
		RestApiId: createAPIOutput.Id,
		ParentId:  rootResourceOutput.Items[0].Id,
		PathPart:  aws.String("test"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao criar recurso: %v", err)})
		return
	}

	// Criar método GET
	_, err = a.client.PutMethod(context.TODO(), &apigateway.PutMethodInput{
		RestApiId:         createAPIOutput.Id,
		ResourceId:        resourceOutput.Id,
		HttpMethod:        aws.String("GET"),
		AuthorizationType: aws.String("NONE"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao criar método: %v", err)})
		return
	}

	// Configurar integração
	_, err = a.client.PutIntegration(context.TODO(), &apigateway.PutIntegrationInput{
		RestApiId:             createAPIOutput.Id,
		ResourceId:            resourceOutput.Id,
		HttpMethod:            aws.String("GET"),
		Type:                  types.IntegrationTypeAwsProxy,
		IntegrationHttpMethod: aws.String("POST"),
		Uri:                   aws.String("arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:000000000000:function:minha-funcao/invocations"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao configurar integração: %v", err)})
		return
	}

	// Configurar resposta do método
	_, err = a.client.PutMethodResponse(context.TODO(), &apigateway.PutMethodResponseInput{
		RestApiId:  createAPIOutput.Id,
		ResourceId: resourceOutput.Id,
		HttpMethod: aws.String("GET"),
		StatusCode: aws.String("200"),
		ResponseModels: map[string]string{
			"application/json": "Empty",
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao configurar resposta: %v", err)})
		return
	}

	// Configurar resposta da integração
	_, err = a.client.PutIntegrationResponse(context.TODO(), &apigateway.PutIntegrationResponseInput{
		RestApiId:  createAPIOutput.Id,
		ResourceId: resourceOutput.Id,
		HttpMethod: aws.String("GET"),
		StatusCode: aws.String("200"),
		ResponseTemplates: map[string]string{
			"application/json": `{"message": "Hello from API Gateway!"}`,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao configurar resposta da integração: %v", err)})
		return
	}

	// Criar deployment
	_, err = a.client.CreateDeployment(context.TODO(), &apigateway.CreateDeploymentInput{
		RestApiId: createAPIOutput.Id,
		StageName: aws.String("test"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao criar deployment: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API criada com sucesso",
		"api_id":  *createAPIOutput.Id,
		"url":     fmt.Sprintf("http://localhost:4566/restapis/%s/test/stages/test", *createAPIOutput.Id),
	})
}

func (a *APIGatewayController) ListAPIs(c *gin.Context) {
	result, err := a.client.GetRestApis(context.TODO(), &apigateway.GetRestApisInput{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao listar APIs: %v", err)})
		return
	}

	apis := make([]map[string]string, 0)
	for _, api := range result.Items {
		apis = append(apis, map[string]string{
			"id":          *api.Id,
			"name":        *api.Name,
			"description": *api.Description,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"apis": apis,
	})
}

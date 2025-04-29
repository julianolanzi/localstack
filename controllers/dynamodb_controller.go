package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DynamoDBController struct {
	client *dynamodb.Client
}

func NewDynamoDBController(cfg aws.Config) *DynamoDBController {
	client := dynamodb.NewFromConfig(cfg)
	return &DynamoDBController{
		client: client,
	}
}

func (d *DynamoDBController) setupTable() error {
	// Verificar se a tabela já existe
	_, err := d.client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String("users"),
	})
	if err == nil {
		return nil // Tabela já existe
	}

	// Criar tabela se não existir
	_, err = d.client.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		TableName: aws.String("users"),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	})
	if err != nil {
		return fmt.Errorf("erro ao criar tabela: %v", err)
	}

	// Esperar a tabela ficar ativa
	waiter := dynamodb.NewTableExistsWaiter(d.client)
	err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String("users"),
	}, 30*time.Second)
	if err != nil {
		return fmt.Errorf("erro ao esperar tabela ficar ativa: %v", err)
	}

	return nil
}

type User struct {
	ID             string `json:"id"`
	Name           string `json:"name" binding:"required"`
	Email          string `json:"email" binding:"required"`
	EmployeeNumber string `json:"employee_number" binding:"required"`
	CreatedAt      string `json:"created_at"`
}

func (d *DynamoDBController) CreateUser(c *gin.Context) {
	// Configurar tabela na primeira chamada
	if err := d.setupTable(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Gerar ID único
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now().Format(time.RFC3339)

	// Criar item no DynamoDB
	_, err := d.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("users"),
		Item: map[string]types.AttributeValue{
			"id":              &types.AttributeValueMemberS{Value: user.ID},
			"name":            &types.AttributeValueMemberS{Value: user.Name},
			"email":           &types.AttributeValueMemberS{Value: user.Email},
			"employee_number": &types.AttributeValueMemberS{Value: user.EmployeeNumber},
			"created_at":      &types.AttributeValueMemberS{Value: user.CreatedAt},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao criar usuário: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (d *DynamoDBController) GetUser(c *gin.Context) {
	// Configurar tabela na primeira chamada
	if err := d.setupTable(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID é obrigatório"})
		return
	}

	// Buscar usuário no DynamoDB
	result, err := d.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String("users"),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao buscar usuário: %v", err)})
		return
	}

	if result.Item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	user := User{
		ID:             result.Item["id"].(*types.AttributeValueMemberS).Value,
		Name:           result.Item["name"].(*types.AttributeValueMemberS).Value,
		Email:          result.Item["email"].(*types.AttributeValueMemberS).Value,
		EmployeeNumber: result.Item["employee_number"].(*types.AttributeValueMemberS).Value,
		CreatedAt:      result.Item["created_at"].(*types.AttributeValueMemberS).Value,
	}

	c.JSON(http.StatusOK, user)
}

func (d *DynamoDBController) UpdateUser(c *gin.Context) {
	// Configurar tabela na primeira chamada
	if err := d.setupTable(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID é obrigatório"})
		return
	}

	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Atualizar usuário no DynamoDB
	_, err := d.client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String("users"),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String("SET #name = :name, #email = :email, #employee_number = :employee_number"),
		ExpressionAttributeNames: map[string]string{
			"#name":            "name",
			"#email":           "email",
			"#employee_number": "employee_number",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":name":            &types.AttributeValueMemberS{Value: user.Name},
			":email":           &types.AttributeValueMemberS{Value: user.Email},
			":employee_number": &types.AttributeValueMemberS{Value: user.EmployeeNumber},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao atualizar usuário: %v", err)})
		return
	}

	// Buscar usuário atualizado
	result, err := d.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String("users"),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao buscar usuário atualizado: %v", err)})
		return
	}

	updatedUser := User{
		ID:             result.Item["id"].(*types.AttributeValueMemberS).Value,
		Name:           result.Item["name"].(*types.AttributeValueMemberS).Value,
		Email:          result.Item["email"].(*types.AttributeValueMemberS).Value,
		EmployeeNumber: result.Item["employee_number"].(*types.AttributeValueMemberS).Value,
		CreatedAt:      result.Item["created_at"].(*types.AttributeValueMemberS).Value,
	}

	c.JSON(http.StatusOK, updatedUser)
}

func (d *DynamoDBController) DeleteUser(c *gin.Context) {
	// Configurar tabela na primeira chamada
	if err := d.setupTable(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID é obrigatório"})
		return
	}

	// Deletar usuário do DynamoDB
	_, err := d.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String("users"),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao deletar usuário: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuário deletado com sucesso"})
}

func (d *DynamoDBController) ListUsers(c *gin.Context) {
	// Configurar tabela na primeira chamada
	if err := d.setupTable(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Listar todos os usuários do DynamoDB
	result, err := d.client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String("users"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao listar usuários: %v", err)})
		return
	}

	users := make([]User, 0)
	for _, item := range result.Items {
		users = append(users, User{
			ID:             item["id"].(*types.AttributeValueMemberS).Value,
			Name:           item["name"].(*types.AttributeValueMemberS).Value,
			Email:          item["email"].(*types.AttributeValueMemberS).Value,
			EmployeeNumber: item["employee_number"].(*types.AttributeValueMemberS).Value,
			CreatedAt:      item["created_at"].(*types.AttributeValueMemberS).Value,
		})
	}

	c.JSON(http.StatusOK, users)
}

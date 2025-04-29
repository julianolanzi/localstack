package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/gin-gonic/gin"
)

type SNSController struct {
	client    *sns.Client
	topicARN  string
	topicName string
}

func NewSNSController(cfg aws.Config) *SNSController {
	client := sns.NewFromConfig(cfg)
	return &SNSController{
		client:    client,
		topicName: "demo-topic",
	}
}

func (s *SNSController) setupTopic() error {
	// Criar tópico se não existir
	createTopicOutput, err := s.client.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String(s.topicName),
	})
	if err != nil {
		return fmt.Errorf("erro ao criar tópico SNS: %v", err)
	}

	s.topicARN = *createTopicOutput.TopicArn

	// Criar uma inscrição padrão
	_, err = s.client.Subscribe(context.TODO(), &sns.SubscribeInput{
		TopicArn: aws.String(s.topicARN),
		Protocol: aws.String("email"),
		Endpoint: aws.String("test@example.com"),
	})
	if err != nil {
		return fmt.Errorf("erro ao criar inscrição padrão: %v", err)
	}

	return nil
}

type PublishMessageRequest struct {
	Message string `json:"message" binding:"required"`
	Subject string `json:"subject" binding:"required"`
}

func (s *SNSController) PublishMessage(c *gin.Context) {
	// Configurar tópico na primeira chamada
	if s.topicARN == "" {
		if err := s.setupTopic(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	var req PublishMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mensagem e assunto são obrigatórios"})
		return
	}

	_, err := s.client.Publish(context.TODO(), &sns.PublishInput{
		TopicArn: aws.String(s.topicARN),
		Message:  aws.String(req.Message),
		Subject:  aws.String(req.Subject),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao publicar mensagem: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mensagem publicada com sucesso",
		"topic":   s.topicName,
	})
}

type SubscribeRequest struct {
	Protocol string `json:"protocol" binding:"required"`
	Endpoint string `json:"endpoint" binding:"required"`
}

func (s *SNSController) Subscribe(c *gin.Context) {
	// Configurar tópico na primeira chamada
	if s.topicARN == "" {
		if err := s.setupTopic(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Protocolo e endpoint são obrigatórios"})
		return
	}

	_, err := s.client.Subscribe(context.TODO(), &sns.SubscribeInput{
		TopicArn: aws.String(s.topicARN),
		Protocol: aws.String(req.Protocol),
		Endpoint: aws.String(req.Endpoint),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao criar inscrição: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Inscrição criada com sucesso",
		"topic":   s.topicName,
	})
}

func (s *SNSController) ListSubscriptions(c *gin.Context) {
	// Configurar tópico na primeira chamada
	if s.topicARN == "" {
		if err := s.setupTopic(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Listar inscrições do tópico
	result, err := s.client.ListSubscriptionsByTopic(context.TODO(), &sns.ListSubscriptionsByTopicInput{
		TopicArn: aws.String(s.topicARN),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao listar inscrições: %v", err)})
		return
	}

	subscriptions := make([]map[string]string, 0)
	for _, sub := range result.Subscriptions {
		subscriptions = append(subscriptions, map[string]string{
			"endpoint": *sub.Endpoint,
			"protocol": *sub.Protocol,
			"arn":      *sub.SubscriptionArn,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"topic":         s.topicName,
		"subscriptions": subscriptions,
	})
}

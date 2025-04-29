package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"
)

type SQSController struct {
	client    *sqs.Client
	queueURL  string
	queueName string
}

func NewSQSController(cfg aws.Config) *SQSController {
	client := sqs.NewFromConfig(cfg)
	return &SQSController{
		client:    client,
		queueName: "demo-queue",
	}
}

func (s *SQSController) setupQueue() error {
	// Criar fila se não existir
	createQueueOutput, err := s.client.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: aws.String(s.queueName),
	})
	if err != nil {
		return fmt.Errorf("erro ao criar fila SQS: %v", err)
	}

	s.queueURL = *createQueueOutput.QueueUrl
	return nil
}

type SendMessageRequest struct {
	Message string `json:"message" binding:"required"`
}

func (s *SQSController) SendMessage(c *gin.Context) {
	// Configurar fila na primeira chamada
	if s.queueURL == "" {
		if err := s.setupQueue(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mensagem é obrigatória"})
		return
	}

	_, err := s.client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueURL),
		MessageBody: aws.String(req.Message),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao enviar mensagem: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mensagem enviada com sucesso",
	})
}

func (s *SQSController) ReceiveMessage(c *gin.Context) {
	// Configurar fila na primeira chamada
	if s.queueURL == "" {
		if err := s.setupQueue(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Receber mensagem
	result, err := s.client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20, // Long polling
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao receber mensagem: %v", err)})
		return
	}

	if len(result.Messages) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Nenhuma mensagem na fila",
		})
		return
	}

	// Deletar mensagem após receber
	_, err = s.client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL),
		ReceiptHandle: result.Messages[0].ReceiptHandle,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao deletar mensagem: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": *result.Messages[0].Body,
	})
}

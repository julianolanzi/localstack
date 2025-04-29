package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
)

type S3Controller struct {
	client *s3.Client
}

func NewS3Controller(cfg aws.Config) *S3Controller {
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	return &S3Controller{client: client}
}

func (s *S3Controller) setupBucket() error {
	_, err := s.client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String("demo-bucket"),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraintSaEast1,
		},
	})
	if err != nil {
		if !isBucketAlreadyExistsError(err) {
			return fmt.Errorf("erro ao criar bucket S3: %v", err)
		}
	}
	return nil
}

func isBucketAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == "BucketAlreadyOwnedByYou" {
		return true
	}
	return false
}

func (s *S3Controller) UploadFile(c *gin.Context) {
	// Configurar bucket na primeira chamada
	if err := s.setupBucket(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não encontrado no formulário"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao abrir arquivo"})
		return
	}
	defer src.Close()

	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String("demo-bucket"),
		Key:    aws.String(file.Filename),
		Body:   src,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao fazer upload: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Arquivo %s enviado com sucesso", file.Filename),
	})
}

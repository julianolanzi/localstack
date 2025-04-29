package routes

import (
	"localstackdemo/controllers"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, cfg aws.Config) {
	s3Controller := controllers.NewS3Controller(cfg)
	sqsController := controllers.NewSQSController(cfg)

	// Grupo de rotas S3
	s3 := r.Group("/s3")
	{
		s3.POST("/upload", s3Controller.UploadFile)
	}

	// Grupo de rotas SQS
	sqs := r.Group("/sqs")
	{
		sqs.POST("/send", sqsController.SendMessage)
		sqs.GET("/receive", sqsController.ReceiveMessage)
	}

	// Grupo de rotas SNS
	snsController := controllers.NewSNSController(cfg)
	sns := r.Group("/sns")
	{
		sns.POST("/publish", snsController.PublishMessage)
		sns.POST("/subscribe", snsController.Subscribe)
		sns.GET("/subscriptions", snsController.ListSubscriptions)
	}

	// Grupo de rotas API Gateway
	apiGatewayController := controllers.NewAPIGatewayController(cfg)
	api := r.Group("/api-gateway")
	{
		api.POST("/create", apiGatewayController.CreateAPI)
		api.GET("/list", apiGatewayController.ListAPIs)
	}

	// Grupo de rotas Lambda
	lambdaController := controllers.NewLambdaController(cfg)
	lambda := r.Group("/lambda")
	{
		lambda.POST("/create", lambdaController.CreateFunction)
		lambda.GET("/list", lambdaController.ListFunctions)
		lambda.POST("/invoke/:name", lambdaController.InvokeFunction)
	}
}

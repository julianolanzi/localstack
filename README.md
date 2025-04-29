# LocalStack Demo

Este projeto demonstra o uso do LocalStack para simular serviços AWS localmente, incluindo S3, SQS, SNS, API Gateway e Lambda.

## Pré-requisitos

- Docker
- Docker Compose
- Go 1.16+
- Make (opcional)

## Configuração

1. Clone o repositório

2. Configure o perfil AWS local:
```bash
mkdir -p ~/.aws
cat > ~/.aws/credentials << EOL
[default]
aws_access_key_id = test
aws_secret_access_key = test
region = sa-east-1
EOL
```

3. Instale as dependências:
```bash
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/service/sqs
go get github.com/aws/aws-sdk-go-v2/service/sns
go get github.com/aws/aws-sdk-go-v2/service/apigateway
go get github.com/aws/aws-sdk-go-v2/service/lambda
go get github.com/aws/aws-lambda-go/lambda
go get github.com/gin-gonic/gin
```

4. Inicie o LocalStack:
```bash
docker-compose up -d
```

5. Compile a função Lambda:
```bash
cd lambda
GOOS=linux GOARCH=amd64 go build -o main main.go
zip function.zip main
cd ..
```

6. Inicie a aplicação:
```bash
go run main.go
```

## Endpoints Disponíveis

### S3

1. Upload de arquivo:
```bash
curl -X POST http://localhost:6000/s3/upload \
  -H "Content-Type: multipart/form-data" \
  -F "file=@/caminho/para/seu/arquivo.txt"
```

### SQS

1. Enviar mensagem:
```bash
curl -X POST http://localhost:6000/sqs/send \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Hello from SQS!"
  }'
```

2. Receber mensagem:
```bash
curl http://localhost:6000/sqs/receive
```

### SNS

1. Publicar mensagem:
```bash
curl -X POST http://localhost:6000/sns/publish \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Hello from SNS!",
    "subject": "Test Message"
  }'
```

2. Criar inscrição:
```bash
curl -X POST http://localhost:6000/sns/subscribe \
  -H "Content-Type: application/json" \
  -d '{
    "protocol": "email",
    "endpoint": "seu-email@exemplo.com"
  }'
```

3. Listar inscrições:
```bash
curl http://localhost:6000/sns/subscriptions
```

### API Gateway

1. Criar API:
```bash
curl -X POST http://localhost:6000/api-gateway/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Minha API",
    "description": "API de teste"
  }'
```

2. Listar APIs:
```bash
curl http://localhost:6000/api-gateway/list
```

3. Testar endpoint da API (substitua {api-id} pelo ID retornado na criação):
```bash
curl http://localhost:4566/restapis/{api-id}/test/stages/test/test
```

### Lambda

1. Criar função:
```bash
curl -X POST http://localhost:6000/lambda/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "minha-funcao",
    "description": "Função de teste"
  }'
```

2. Listar funções:
```bash
curl http://localhost:6000/lambda/list
```

3. Invocar função:
```bash
curl -X POST http://localhost:6000/lambda/invoke/minha-funcao
```

## Fluxo Completo de Exemplo

1. Criar função Lambda:
```bash
curl -X POST http://localhost:6000/lambda/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "minha-funcao",
    "description": "Função de teste"
  }'
```

2. Criar API Gateway:
```bash
curl -X POST http://localhost:6000/api-gateway/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Minha API",
    "description": "API com integração Lambda"
  }'
```

3. Testar endpoint (substitua {api-id} pelo ID retornado):
```bash
curl http://localhost:4566/restapis/{api-id}/test/stages/test/test
```

## Estrutura do Projeto

```
.
├── controllers/
│   ├── s3_controller.go
│   ├── sqs_controller.go
│   ├── sns_controller.go
│   ├── apigateway_controller.go
│   └── lambda_controller.go
├── lambda/
│   └── main.go
├── routes/
│   └── routes.go
├── config/
│   └── aws_config.go
├── main.go
├── docker-compose.yml
└── README.md
```

## Limpeza

Para parar e remover os containers:
```bash
docker-compose down
```

## Observações

- O LocalStack está configurado para rodar na porta 4566
- A aplicação Go está configurada para rodar na porta 6000
- Todos os serviços AWS são simulados localmente
- As credenciais AWS são configuradas automaticamente para o ambiente local
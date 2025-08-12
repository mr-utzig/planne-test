# API de Baldes de Frutas
Esta é uma API RESTful simples desenvolvida em Go para gerenciar baldes de frutas, conforme as especificações do desafio.

## Tecnologias Utilizadas
- Linguagem: Go (Golang)
- Roteador HTTP: Chi v5
- Banco de Dados: SQLite 3
- Driver do Banco: mattn/go-sqlite3

## Funcionalidades
- Criação e exclusão de Baldes.
- Criação e exclusão de Frutas.
- Depósito e remoção de Frutas de Baldes.
- Listagem de Baldes com detalhes (valor total, ocupação) e ordenação.
- Persistência de dados em um arquivo SQLite (fruit_buckets.db).
- Remoção automática de frutas expiradas através de uma rotina em background.

## Pré-requisitos
- Go 1.24 ou superior instalado.
- Um terminal com curl para testar os endpoints (ou qualquer cliente de API como Postman/Insomnia).

## Como Executar os Testes
Com os arquivos de teste no lugar, navegue até o diretório raiz do projeto no seu terminal e execute o seguinte comando:
```bash
go test ./... -v
```
Você deverá ver uma saída indicando que todos os testes passaram com sucesso (--- PASS).

## Como Executar a Aplicação
### 1. Baixar Dependências:
Navegue até a pasta raiz do projeto e execute o comando abaixo para baixar as dependências (chi e go-sqlite3):
    ```bash
    go mod tidy
    ```
### 2. Iniciar o Servidor:
Ainda na pasta raiz, execute o comando para compilar e iniciar a aplicação:
    ```bash
    go run .
    ```
    Você verá a seguinte mensagem no console, indicando que o servidor está no ar:
    ```bash
    Servidor iniciado na porta :8080
    ```
    Um arquivo __fruit_buckets.db__ será criado no diretório raiz para armazenar os dados.

## Endpoints da API
Aqui estão os endpoints disponíveis e exemplos de como usá-los com curl.

### 1. Baldes (/buckets)
__POST__ /buckets - Criar um novo balde
Cria um balde com a capacidade especificada.

Exemplo:
```bash
curl -X POST http://localhost:8080/buckets -d '{"capacity": 5}'
```
Resposta:
```json
{"id":1,"capacity":5}
```
__GET__ /buckets - Listar todos os baldes
Retorna uma lista de todos os baldes, com detalhes sobre as frutas contidas, o valor total e a porcentagem de ocupação. A lista é ordenada de forma decrescente pela ocupação.

Exemplo:
```bash
curl http://localhost:8080/buckets
```
Resposta:
```json
[
    {
        "id": 2,
        "capacity": 3,
        "fruits": [
            {"id": 3, "name": "Pera", "price": 2.5, "expiration_time": 1723494540, "bucket_id": {"Int64": 2, "Valid": true}},
            {"id": 4, "name": "Uva", "price": 7.8, "expiration_time": 1723494600, "bucket_id": {"Int64": 2, "Valid": true}}
        ],
        "total_value": 10.3,
        "occupancy_percentage": 66.66666666666667
    },
    {
        "id": 1,
        "capacity": 5,
        "fruits": [
            {"id": 1, "name": "Maçã", "price": 1.5, "expiration_time": 1723494480, "bucket_id": {"Int64": 1, "Valid": true}}
        ],
        "total_value": 1.5,
        "occupancy_percentage": 20
    }
]
```
__DELETE__ /buckets/{bucketID} - Excluir um balde
Exclui um balde. A operação só é permitida se o balde estiver vazio.

Exemplo:
```bash
curl -X DELETE http://localhost:8080/buckets/3
```
Resposta:
```bash
204 No Content se for bem-sucedido.
```
```bash
400 Bad Request se o balde não estiver vazio.
```
### 2. Frutas (/fruits)
__POST__ /fruits - Criar uma nova fruta
Cria uma fruta com nome, preço e tempo de expiração em segundos a partir do momento da criação.

Exemplo (fruta que expira em 1 hora):
```bash
curl -X POST http://localhost:8080/fruits -d '{"name": "Banana", "price": 0.75, "expires_in_seconds": 3600}'
```
Resposta:
```json
{"id":5,"name":"Banana","price":0.75,"expiration_time":1723497965,"bucket_id":{"Int64":0,"Valid":false}}
```
__DELETE__ /fruits/{fruitID} - Excluir uma fruta
Exclui uma fruta permanentemente do sistema, independentemente de estar em um balde ou não.

Exemplo:
```bash
curl -X DELETE http://localhost:8080/fruits/5
```
Resposta:
```bash
204 No Content.
```
### 3. Operações entre Baldes e Frutas
__POST__ /buckets/{bucketID}/fruits - Depositar uma fruta em um balde
Move uma fruta existente (que não está em nenhum balde) para dentro de um balde específico.

Exemplo (depositar a fruta com ID 5 no balde com ID 1):
```bash
curl -X POST http://localhost:8080/buckets/1/fruits -d '{"fruit_id": 5}'
```
Resposta:
```json
{"message":"Fruta depositada com sucesso"}
```
Casos de erro:
- A capacidade do balde foi excedida.
- A fruta já está em outro balde.
- A fruta ou o balde não existem.

__DELETE__ /buckets/{bucketID}/fruits/{fruitID} - Remover uma fruta de um balde

Exemplo (remover a fruta 5 do balde 1):
```bash
curl -X DELETE http://localhost:8080/buckets/1/fruits/5
```
Resposta:
```json
{"message":"Fruta removida com sucesso"}
```
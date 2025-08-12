package models

import (
	"database/sql"
	"log"
	"time"

	"github.com/mr-utzig/planne-test/database"
)

// Fruit representa a estrutura de uma fruta no banco de dados.
type Fruit struct {
	ID             int           `json:"id"`
	Name           string        `json:"name"`
	Price          float64       `json:"price"`
	ExpirationTime int64         `json:"expiration_time"`
	BucketID       sql.NullInt64 `json:"bucket_id"`
}

// CreateFruitRequest é a estrutura do corpo da requisição para criar uma nova fruta.
// Usa `ExpiresInSeconds` para facilitar a entrada do usuário.
type CreateFruitRequest struct {
	Name             string  `json:"name"`
	Price            float64 `json:"price"`
	ExpiresInSeconds int64   `json:"expires_in_seconds"`
}

func (f *Fruit) GetByID(id int) error {
	row := database.DB.QueryRow("SELECT id, name, price, expiration_time, bucket_id FROM fruits WHERE id = ?", id)

	if err := row.Scan(&f.ID, &f.Name, &f.Price, &f.ExpirationTime, &f.BucketID); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (f Fruit) GetFruitsInBucket(bucketID int) ([]Fruit, error) {
	rows, err := database.DB.Query("SELECT id, name, price, expiration_time, bucket_id FROM fruits WHERE bucket_id = ?", bucketID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var fruits []Fruit
	for rows.Next() {
		var fruit Fruit
		if err := rows.Scan(&fruit.ID, &fruit.Name, &fruit.Price, &fruit.ExpirationTime, &fruit.BucketID); err != nil {
			log.Println(err)
			return nil, err
		}

		fruits = append(fruits, fruit)
	}

	return fruits, nil
}

func (f *Fruit) AddToBucket(bucketID int) (int64, error) {
	result, err := database.DB.Exec("UPDATE fruits SET bucket_id = ? WHERE id = ?", bucketID, f.ID)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return result.RowsAffected()

}

func (f Fruit) RemoveFromBucket(fruitID, bucketID int) (int64, error) {
	result, err := database.DB.Exec("UPDATE fruits SET bucket_id = NULL WHERE id = ? AND bucket_id = ?", fruitID, bucketID)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return result.RowsAffected()
}

func (f Fruit) DeleteByID(id int) error {
	_, err := database.DB.Exec("DELETE FROM fruits WHERE id = ?", id)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (f Fruit) DeleteExpireds() int64 {
	now := time.Now().Unix()

	result, err := database.DB.Exec("DELETE FROM fruits WHERE expiration_time <= ?", now)
	if err != nil {
		log.Println("Erro ao limpar frutas expiradas:", err)
		return 0
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Erro ao obter linhas afetadas pela limpeza:", err)
		return 0
	}

	return rowsAffected
}

func (f *CreateFruitRequest) InsertFruitFromPayload() (*Fruit, error) {
	expirationTime := time.Now().Add(time.Duration(f.ExpiresInSeconds) * time.Second).Unix()

	result, err := database.DB.Exec(
		"INSERT INTO fruits (name, price, expiration_time) VALUES (?, ?, ?)",
		f.Name, f.Price, expirationTime,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	id, _ := result.LastInsertId()
	fruit := &Fruit{
		ID:             int(id),
		Name:           f.Name,
		Price:          f.Price,
		ExpirationTime: expirationTime,
	}

	return fruit, nil
}

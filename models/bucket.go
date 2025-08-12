package models

import (
	"log"

	"github.com/mr-utzig/planne-test/database"
)

// Bucket representa a estrutura de um balde no banco de dados.
type Bucket struct {
	ID       int `json:"id"`
	Capacity int `json:"capacity"`
}

// BucketDetails é uma estrutura mais completa usada para a listagem,
// incluindo informações sobre as frutas contidas.
type BucketDetails struct {
	ID         int     `json:"id"`
	Capacity   int     `json:"capacity"`
	Fruits     []Fruit `json:"fruits"`
	TotalValue float64 `json:"total_value"`
	Occupancy  float64 `json:"occupancy_percentage"`
}

func (b *Bucket) Insert() error {
	result, err := database.DB.Exec("INSERT INTO buckets (capacity) VALUES (?)", b.Capacity)
	if err != nil {
		log.Println(err)
		return err
	}

	id, _ := result.LastInsertId()
	b.ID = int(id)

	return nil
}

func (b *Bucket) GetByID(id int) error {
	row := database.DB.QueryRow("SELECT id, capacity FROM buckets WHERE id = ?", id)

	if err := row.Scan(&b.ID, &b.Capacity); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (b Bucket) GetAll() ([]Bucket, error) {
	rows, err := database.DB.Query("SELECT id, capacity FROM buckets")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var buckets []Bucket
	for rows.Next() {
		var bucket Bucket
		if err := rows.Scan(&bucket.ID, &bucket.Capacity); err != nil {
			log.Println(err)
			return nil, err
		}

		buckets = append(buckets, bucket)
	}

	return buckets, nil
}

func (b Bucket) DeleteByID(id int) error {
	_, err := database.DB.Exec("DELETE FROM buckets WHERE id = ?", id)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (d *BucketDetails) CalcTotalValue() {
	total := 0.0
	for _, fruit := range d.Fruits {
		total += fruit.Price
	}

	d.TotalValue = total
}

func (d *BucketDetails) CalcOccupancyPercentage() {
	if d.Capacity > 0 {
		d.Occupancy = (float64(len(d.Fruits)) / float64(d.Capacity)) * 100
	}
}

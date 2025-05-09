package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DealerSourceCode represents the lookup table for dealer codes & floors
type DealerSourceCode struct {
	ID            uint64 `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	Source        string `gorm:"column:source;type:varchar(255);not null" json:"source"`
	Dealership    string `gorm:"column:dealership;type:varchar(255);not null" json:"dealership"`
	DealerCode    string `gorm:"column:dealer_code;type:varchar(255);not null" json:"dealer_code"`
	FloorCodeNew  string `gorm:"column:floor_code_new;type:varchar(255)" json:"floor_code_new"`
	FloorCodeUsed string `gorm:"column:floor_code_used;type:varchar(255)" json:"floor_code_used"`
	ContactPerson string `gorm:"column:contact_person;type:varchar(255)" json:"contact_person"`
}

// TableName overrides the default pluralized table name
func (DealerSourceCode) TableName() string {
	return "dealer_source_code"
}

// LeadAudit is the Gorm model for auditing requests & responses
type LeadAudit struct {
	ID            uint64          `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	CreatedAt     time.Time       `gorm:"column:created_at;type:datetime;not null" json:"created_at"`
	InputPayload  json.RawMessage `gorm:"column:input_payload;type:json;not null" json:"input_payload"`
	LeadPayload   json.RawMessage `gorm:"column:lead_payload;type:json;not null" json:"lead_payload"`
	LeadReference string          `gorm:"column:lead_reference;type:varchar(255);not null" json:"lead_reference"`
}

// TableName overrides the default pluralized table name
func (LeadAudit) TableName() string {
	return "lead_audit"
}

func setupDB() *gorm.DB {
	ctx := context.Background()

	// Initialize Cloud SQL Connector Dialer
	dialer, err := cloudsqlconn.NewDialer(ctx)
	if err != nil {
		panic("failed to initialize Cloud SQL Connector:  " + err.Error())
	}
	defer dialer.Close()

	// Construct DSN using Cloud SQL Go Connector
	dsn := fmt.Sprintf("%s:%s@unix(/cloudsql/%s)/%s?parseTime=true",
		os.Getenv("CLOUD_MYSQL_USER"),
		os.Getenv("CLOUD_MYSQL_PASSWORD"),
		os.Getenv("CLOUD_MYSQL_CONNECTION_NAME"),
		os.Getenv("CLOUD_MYSQL_SCHEMA"),
	)

	// Open a database connection using the Cloud SQL Connector
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic("failed to open database:  " + err.Error())
	}

	// Apply the Cloud SQL Connector dialer
	db.SetConnMaxIdleTime(0) // Keep connections open indefinitely

	// Use GORM with the established SQL connection
	database, err := gorm.Open(mysql.New(mysql.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		panic("failed to connect to the database:  " + err.Error())
	}

	return database
}

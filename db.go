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
	ID            uint64    `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	Source        string    `gorm:"column:source;type:varchar(255);not null" json:"source"`
	Dealership    string    `gorm:"column:dealership;type:varchar(255);not null" json:"dealership"`
	DealerCode    string    `gorm:"column:dealer_code;type:varchar(255);not null" json:"dealer_code"`
	FloorCodeNew  string    `gorm:"column:floor_code_new;type:varchar(255)" json:"floor_code_new"`
	FloorCodeUsed string    `gorm:"column:floor_code_used;type:varchar(255)" json:"floor_code_used"`
	ContactPerson string    `gorm:"column:contact_person;type:varchar(255)" json:"contact_person"`
	Created       time.Time `gorm:"column:created;type:datetime;default:CURRENT_TIMESTAMP;autoCreateTime" json:"created"`
	Modified      time.Time `gorm:"column:modified;type:datetime;default:CURRENT_TIMESTAMP;autoCreateTime" json:"modified"`
}

// TableName overrides the default pluralized table name
func (DealerSourceCode) TableName() string {
	return "dealer_source_code"
}

// LeadAudit is the Gorm model for auditing requests & responses
type LeadAudit struct {
	ID            uint64          `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	CreatedAt     time.Time       `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;autoCreateTime" json:"created_at"`
	InputPayload  json.RawMessage `gorm:"column:input_payload;type:json;not null" json:"input_payload"`
	LeadPayload   json.RawMessage `gorm:"column:lead_payload;type:json;not null" json:"lead_payload"`
	LeadReference string          `gorm:"column:lead_reference;type:varchar(255);not null" json:"lead_reference"`
	Created       time.Time       `gorm:"column:created;type:datetime;default:CURRENT_TIMESTAMP;autoCreateTime" json:"created"`
	Modified      time.Time       `gorm:"column:modified;type:datetime;default:CURRENT_TIMESTAMP;autoCreateTime" json:"modified"`
}

// TableName overrides the default pluralized table name
func (LeadAudit) TableName() string {
	return "lead_audit"
}

// Contact represents the contact person for a given dealer source
type Contact struct {
	ID                           uint64    `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"-"`
	DealerSourceID               uint64    `gorm:"column:dealer_source_id;type:bigint unsigned;not null;index" json:"-"`
	Title                        string    `gorm:"column:title;type:varchar(255);not null" json:"title"`
	FirstName                    string    `gorm:"column:first_name;type:varchar(255);not null" json:"firstName"`
	Surname                      string    `gorm:"column:surname;type:varchar(255);not null" json:"surname"`
	Email                        string    `gorm:"column:email;type:varchar(255);not null;uniqueIndex" json:"email"`
	OfficePhone                  string    `gorm:"column:office_phone;type:varchar(50)" json:"officePhone"`
	CellPhone                    string    `gorm:"column:cell_phone;type:varchar(50);not null" json:"cellPhone"`
	DriversLicense               string    `gorm:"column:drivers_license;type:varchar(100)" json:"driversLicense"`
	IncomeBracket                string    `gorm:"column:income_bracket;type:varchar(100)" json:"incomeBracket"`
	PreferredContactMethod       string    `gorm:"column:preferred_contact_method;type:varchar(100)" json:"preferredContactMethod"`
	PreferredContactTime         string    `gorm:"column:preferred_contact_time;type:varchar(100)" json:"preferredContactTime"`
	Citizenship                  string    `gorm:"column:citizenship;type:varchar(100);not null" json:"citizenship"`
	IDNo                         string    `gorm:"column:id_no;type:varchar(50)" json:"idNo"`
	BirthDate                    string    `gorm:"column:birth_date;type:date" json:"birthDate"`
	Gender                       string    `gorm:"column:gender;type:varchar(50)" json:"gender"`
	Ethnicity                    string    `gorm:"column:ethnicity;type:varchar(100)" json:"ethnicity"`
	HomeLanguage                 string    `gorm:"column:home_language;type:varchar(100)" json:"homeLanguage"`
	ResidentialAddressLine1      string    `gorm:"column:residential_address_line1;type:varchar(255)" json:"residentialAddressLine1"`
	ResidentialAddressLine2      string    `gorm:"column:residential_address_line2;type:varchar(255)" json:"residentialAddressLine2"`
	ResidentialAddressSuburb     string    `gorm:"column:residential_address_suburb;type:varchar(100)" json:"residentialAddressSuburb"`
	ResidentialAddressCity       string    `gorm:"column:residential_address_city;type:varchar(100)" json:"residentialAddressCity"`
	ResidentialAddressPostalCode string    `gorm:"column:residential_address_postal_code;type:varchar(20)" json:"residentialAddressPostalCode"`
	ResidentialAddressProvince   string    `gorm:"column:residential_address_province;type:varchar(100)" json:"residentialAddressProvince"`
	PostalAddressLine1           string    `gorm:"column:postal_address_line1;type:varchar(255)" json:"postalAddressLine1"`
	PostalAddressLine2           string    `gorm:"column:postal_address_line2;type:varchar(255)" json:"postalAddressLine2"`
	PostalAddressSuburb          string    `gorm:"column:postal_address_suburb;type:varchar(100)" json:"postalAddressSuburb"`
	PostalAddressCity            string    `gorm:"column:postal_address_city;type:varchar(100)" json:"postalAddressCity"`
	PostalAddressCode            string    `gorm:"column:postal_address_code;type:varchar(20)" json:"postalAddressCode"`
	PostalAddressProvince        string    `gorm:"column:postal_address_province;type:varchar(100)" json:"postalAddressProvince"`
	MarketingConsent             string    `gorm:"column:marketing_consent;type:tinyint(1)" json:"marketingConsent"`
	CreditGrading                string    `gorm:"column:credit_grading;type:varchar(50)" json:"creditGrading"`
	CompanyName                  string    `gorm:"column:company_name;type:varchar(255)" json:"companyName"`
	CompanyType                  string    `gorm:"column:company_type;type:varchar(100)" json:"companyType"`
	Created                      time.Time `gorm:"column:created;type:datetime;default:CURRENT_TIMESTAMP;autoCreateTime" json:"created"`
	Modified                     time.Time `gorm:"column:modified;type:datetime;default:CURRENT_TIMESTAMP;autoCreateTime" json:"modified"`
}

// TableName overrides the default table name
func (Contact) TableName() string {
	return "contact"
}

func setupCloudDB() *gorm.DB {
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

	database.AutoMigrate(&DealerSourceCode{}, &LeadAudit{}, &Contact{})

	return database
}

func setupLocalDB() *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("CLOUD_MYSQL_USER"),
		os.Getenv("CLOUD_MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("CLOUD_MYSQL_SCHEMA"),
	)

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	database.AutoMigrate(&DealerSourceCode{}, &LeadAudit{}, &Contact{})

	return database
}

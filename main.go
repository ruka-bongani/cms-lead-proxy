// File: main.go
package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/gin-gonic/gin"
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

// DataFileds represents the minimal incoming payload
type DataFields struct {
	Source string            `json:"source"`
	Fields map[string]string `json:"fields"`
}

// LeadWrapper & nested types for the CMS payload
type LeadWrapper struct {
	Lead Lead `json:"lead"`
}

type Lead struct {
	DealerRef         string            `json:"dealerRef"`
	DealerFloor       string            `json:"dealerFloor"`
	DealerSalesPerson string            `json:"dealerSalesPerson"`
	Region            string            `json:"region"`
	Source            string            `json:"source"`
	TransactionID     string            `json:"transactionID"`
	ExtLeadRef        string            `json:"extLeadRef"`
	PromotionalCode   string            `json:"promotionalCode"`
	UtmParameters     string            `json:"utmParameters"`
	CountryCode       string            `json:"countryCode"`
	Contact           Contact           `json:"contact"`
	Seeks             Seeks             `json:"seeks"`
	Referrer          map[string]string `json:"referrer"`
	Options           map[string]string `json:"options"`
	TradeIns          []TradeIn         `json:"TradeIns"`
}

type Contact struct {
	Title                        string `json:"title"`
	FirstName                    string `json:"firstName"`
	Surname                      string `json:"surname"`
	Email                        string `json:"email"`
	OfficePhone                  string `json:"officePhone"`
	CellPhone                    string `json:"cellPhone"`
	DriversLicense               string `json:"driversLicense"`
	IncomeBracket                string `json:"incomeBracket"`
	PreferredContactMethod       string `json:"preferredContactMethod"`
	PreferredContactTime         string `json:"preferredContactTime"`
	Citizenship                  string `json:"citizenship"`
	IDNo                         string `json:"idNo"`
	BirthDate                    string `json:"birthDate"`
	Gender                       string `json:"gender"`
	Ethnicity                    string `json:"ethnicity"`
	HomeLanguage                 string `json:"homeLanguage"`
	ResidentialAddressLine1      string `json:"residentialAddressLine1"`
	ResidentialAddressLine2      string `json:"residentialAddressLine2"`
	ResidentialAddressSuburb     string `json:"residentialAddressSuburb"`
	ResidentialAddressCity       string `json:"residentialAddressCity"`
	ResidentialAddressPostalCode string `json:"residentialAddressPostalCode"`
	ResidentialAddressProvince   string `json:"residentialAddressProvince"`
	PostalAddressLine1           string `json:"postalAddressLine1"`
	PostalAddressLine2           string `json:"postalAddressLine2"`
	PostalAddressSuburb          string `json:"postalAddressSuburb"`
	PostalAddressCity            string `json:"postalAddressCity"`
	PostalAddressCode            string `json:"postalAddressCode"`
	PostalAddressProvince        string `json:"postalAddressProvince"`
	MarketingConsent             string `json:"marketingConsent"`
	CreditGrading                string `json:"creditGrading"`
	CompanyName                  string `json:"companyName"`
	CompanyType                  string `json:"companyType"`
}

type Seeks struct {
	Used             string `json:"used"`
	Brand            string `json:"brand"`
	Model            string `json:"model"`
	MmCode           string `json:"mmCode"`
	ModelCode        string `json:"modelCode"`
	Kms              string `json:"kms"`
	Year             string `json:"year"`
	Colour           string `json:"colour"`
	StockNr          string `json:"stockNr"`
	Price            string `json:"price"`
	Deposit          string `json:"deposit"`
	TestDrive        string `json:"testDrive"`
	TradeIn          string `json:"tradeIn"`
	Finance          string `json:"finance"`
	Valuation        string `json:"valuation"`
	Registration     string `json:"registration"`
	Special          string `json:"special"`
	SpecialBannerURL string `json:"specialBannerURL"`
	ServiceHistory   string `json:"serviceHistory"`
	Comments         string `json:"comments"`
	Vin              string `json:"vin"`
}

type TradeIn struct {
	Make       string  `json:"Make"`
	Model      string  `json:"Model"`
	Variant    string  `json:"Variant"`
	Year       int     `json:"Year"`
	Mileage    int     `json:"Mileage"`
	MMCode     *int    `json:"MMCode"`
	IsFinanced bool    `json:"IsFinanced"`
	Price      float64 `json:"Price"`
}

type CMSResponse struct {
	Code          string `json:"code"`
	LeadReference string `json:"leadReference"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

func leadHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var dataFields DataFields
		if err := c.ShouldBindJSON(&dataFields); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// lookup dealer codes & floors
		dealership := dataFields.Fields["Dealership"]
		source := dataFields.Source
		var dsc DealerSourceCode
		if err := db.
			Where("source = ? AND dealership = ?", source, dealership).
			First(&dsc).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dealer not found"})
			return
		}

		// choose the correct floor code
		floor := dsc.FloorCodeNew
		if dataFields.Fields["Our cars"] == "Used vehicles" {
			floor = dsc.FloorCodeUsed
		}

		// build CMS lead payload
		lead := Lead{
			DealerRef:         dsc.DealerCode,
			DealerFloor:       floor,
			DealerSalesPerson: dsc.ContactPerson,
			Region:            dealership,
			Source:            source,
			Contact: Contact{
				Title:        "Mr",
				FirstName:    "Dean",
				Surname:      "Kabasa",
				Email:        "dean@ruka.co.za",
				CellPhone:    "0831111111",
				Citizenship:  "South Africa",
				BirthDate:    "1980-08-05",
				Gender:       "Male",
				HomeLanguage: "English",
			},
			Seeks: Seeks{
				Used:     "1",
				Brand:    "Ford",
				Model:    "Fiesta ST",
				Kms:      "15000",
				Year:     "2016",
				StockNr:  "123Nr",
				Price:    "199800",
				Deposit:  "20500.00",
				Comments: "Comment test",
			},
			Referrer: map[string]string{},
			Options:  map[string]string{},
			TradeIns: []TradeIn{
				{Make: "Mazda", Model: "3", Variant: "3 1.6 tdi", Year: 2018, Mileage: 60500, Price: 230500},
				{Make: "Ford", Model: "Figo", Variant: "1.4", Year: 2016, Mileage: 98500, MMCode: ptrInt(12344321), IsFinanced: true, Price: 160500.70},
			},
		}

		wrapper := LeadWrapper{Lead: lead}
		leadJSON, _ := json.Marshal(wrapper)

		// forward to external API
		resp, err := http.Post(
			os.Getenv("CMS_API_URL"),
			"application/json",
			bytes.NewReader(leadJSON),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		var cmsResp CMSResponse
		if err := json.NewDecoder(resp.Body).Decode(&cmsResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// audit to database
		inputJSON, _ := json.Marshal(dataFields)
		audit := LeadAudit{
			InputPayload:  json.RawMessage(inputJSON),
			LeadPayload:   json.RawMessage(leadJSON),
			LeadReference: cmsResp.LeadReference,
		}
		if err := db.Create(&audit).Error; err != nil {
			log.Printf("audit save error: %v", err)
		}

		c.JSON(resp.StatusCode, cmsResp)
	}
}

func ptrInt(i int) *int { return &i }

func main() {
	db := setupDB()

	r := gin.Default()
	r.POST("/lead", leadHandler(db))

	port := os.Getenv("PORT")
	log.Println("Listening on :", port)
	log.Fatal(r.Run(":" + port))
}

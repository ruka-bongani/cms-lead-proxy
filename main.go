// File: main.go
package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DataFileds represents the minimal incoming payload
type DataFields struct {
	Source string            `json:"source"`
	Fields map[string]string `json:"fields"`
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

func main() {
	db := setupDB()

	r := gin.Default()
	r.POST("/lead", leadHandler(db))

	port := os.Getenv("API_PORT")
	log.Println("Listening on :", port)
	log.Fatal(r.Run(":" + port))
}

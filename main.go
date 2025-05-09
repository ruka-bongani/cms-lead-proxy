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
type DataFields map[string]string

func leadHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var dataFields DataFields
		if err := c.ShouldBindJSON(&dataFields); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Println("Received payload:", dataFields)

		// lookup dealer codes & floors
		dealership := dataFields["Dealership"]
		source := dataFields["Source"]
		var dsc DealerSourceCode
		if err := db.
			Where("source = ?", source).
			Where("dealership = ?", dealership).
			First(&dsc).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dealer not found"})
			return
		}

		// choose the correct floor code
		floor := dsc.FloorCodeNew
		if dataFields["Our cars"] == "Used vehicles" {
			floor = dsc.FloorCodeUsed
		}

		var contact Contact
		if err := db.
			Where("dealer_source_id = ?", dsc.ID).
			First(&contact).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "contact not found"})
			return
		}

		// build CMS lead payload
		lead := Lead{
			DealerRef:         dsc.DealerCode,
			DealerFloor:       floor,
			DealerSalesPerson: dsc.ContactPerson,
			Region:            dealership,
			Source:            source,
			Contact:           contact,
		}

		wrapper := LeadWrapper{Lead: lead}
		leadJSON, _ := json.Marshal(wrapper)

		// forward to external API
		req, err := http.NewRequest(
			"POST",
			os.Getenv("CMS_API_URL"),
			bytes.NewReader(leadJSON),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", os.Getenv("CMS_API_KEY"))

		resp, err := http.DefaultClient.Do(req)
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
	db := setupCloudDB()

	r := gin.Default()
	r.POST("/lead", leadHandler(db))

	port := os.Getenv("API_PORT")
	log.Println("Listening on :", port)
	log.Fatal(r.Run(":" + port))
}

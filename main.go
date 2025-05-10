package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
		province := dataFields["Province"]
		dealership := dataFields["Dealership"]
		internal_source_code := dataFields["Source"]
		var dsc DealerSourceCode
		if err := db.
			Where("internal_source_code = ?", internal_source_code).
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

		contact := Contact{
			FirstName:              dataFields["FullName"],
			CellPhone:              dataFields["MSISDN"],
			PreferredContactMethod: "Cellphone",
		}

		if len(dataFields["AlternateMSISDN"]) >= 10 {
			contact.CellPhone = dataFields["MSISDN"]
		}

		date, err := parseAppointment(dataFields["CallBackDate"], dataFields["CallBackTime"])
		if err != nil {
			panic(err)
		}

		// build CMS lead payload
		lead := Lead{
			DealerRef:         dsc.DealerCode,
			DealerFloor:       floor,
			DealerSalesPerson: dsc.ContactPerson,
			Region:            province,
			Source:            dsc.Source,
			Contact:           contact,
			Appointment: Appointment{
				DateOfAppointment: date,
				PartOfTheDay:      dataFields["CallBackTime"],
			},
		}

		wrapper := LeadWrapper{Lead: lead}
		leadJSON, _ := json.Marshal(wrapper)

		log.Printf("Lead JSON: %s", leadJSON)

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

		// read the raw response
		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// unmarshal into your struct for easy field access
		var cmsResp CMSResponse
		if err := json.Unmarshal(respBytes, &cmsResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Printf("CMS API raw response: %s", respBytes)

		// audit to database
		inputJSON, _ := json.Marshal(dataFields)
		audit := LeadAudit{
			InputPayload:       json.RawMessage(inputJSON),
			LeadPayload:        json.RawMessage(leadJSON),
			CMSResponsePayload: json.RawMessage(respBytes),
			LeadReference:      cmsResp.LeadReference,
		}
		if err := db.Create(&audit).Error; err != nil {
			log.Printf("audit save error: %v", err)
		}

		contact.LeadAuditID = audit.ID
		if err := db.Create(&contact).Error; err != nil {
			log.Printf("contact save error: %v", err)
		}

		c.JSON(resp.StatusCode, cmsResp)
	}
}

func parseAppointment(dateStr, timeRange string) (string, error) {
	date, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date: %w", err)
	}

	parts := strings.Split(timeRange, " - ")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid time range")
	}
	t, err := time.Parse("3:04pm", parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid time: %w", err)
	}

	combined := time.Date(
		date.Year(), date.Month(), date.Day(),
		t.Hour(), t.Minute(), 0, 0,
		time.Local,
	)

	return combined.Format("2006-01-02 15:04:05"), nil
}

func main() {
	db := setupCloudDB()

	r := gin.Default()
	r.POST("/lead", leadHandler(db))

	port := os.Getenv("API_PORT")
	log.Println("Listening on :", port)
	log.Fatal(r.Run(":" + port))
}

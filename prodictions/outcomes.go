package prodictions

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/labstack/echo"
	"github.com/marcato15/api.prodictions.com/db"
	uuid "github.com/satori/go.uuid"
)

var (
	outcomeTable = "Outcomes"
)

// Outcome This is a prodictable outcome of a outcome. Stores the current line, but the line must not be used to pay out koins
type Outcome struct {
	Prodiction     *Prodiction `json:"prodiction,omitempty"`
	OutcomeID      string      `json:"outcomeId,omitempty"`
	Description    string      `json:"description"`
	CurrentLine    string      `json:"currentLine,omitempty"`
	Status         string      `json:"status,omitempty"`
	Hash           string      `json:"hash,omitempty"`
	ProdictionHash string      `json:"prodictionHash,omitempty" dynamodbav:"-"`
}

type BaseOutcome struct {
	Description string `json:"description"`
	CurrentLine string `json:"currentLine,omitempty"`
	Status      string `json:"status,omitempty"`
	Hash        string `json:"hash,omitempty"`
}

func init() {
	Server.GET("/outcomes", listOutcomes)
	Server.POST("/outcomes", createOutcome)
	Server.PUT("/outcomes/:id", updateOutcome)
	Server.GET("/outcomes/:id", getOutcome)
}

// MarshalDynamoDBAttributeValue Custom Marshaller
func (outcome *Outcome) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	return db.CustomMarshalKeepEmptyLists(outcome, []string{"outcomes"}, av)
}

func getOutcome(c echo.Context) error {
	keys, err := db.DecodeKeys(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	outcome := Outcome{}
	err = DB.GetItemByKeysFromTable(keys, outcomeTable, &outcome)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, outcome)
}
func createOutcome(c echo.Context) error {
	outcome := Outcome{}

	if err := c.Bind(outcome); err != nil {
		return c.String(http.StatusBadRequest, "Invalid Object")
	}

	//Retrieve Prodiction
	eventKeys, err := db.DecodeKeys(outcome.EventHash)
	if err != nil {
		fmt.Println("Error decoding keys")
		return c.JSON(http.StatusInternalServerError, err)
	}
	err = DB.GetItemByKeysFromTable(eventKeys, eventTable, &prodiction.Event)
	if err != nil {
		fmt.Println("Could find event group")
		return c.JSON(http.StatusInternalServerError, err)
	}

	// Create UUID
	outcome.OutcomeID = uuid.NewV4().String()

	// Set Hash
	keys := []db.Key{}
	keys = append(keys, db.Key{
		Field: "prodictionHash",
		Value: outcome.ProdictionHash,
	})
	keys = append(keys, db.Key{
		Field: "outcomeId",
		Value: outcome.OutcomeID,
	})
	encodedKeys, err := db.EncodeKeys(keys)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	outcome.Hash = encodedKeys

	_, err = DB.PutItemInTable(outcome, outcomeTable)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	//Now we need to add to Prodiction
	//Convert Outcome to BaseOutcome via json.(Un)Marshal
	j, _ := json.Marshal(outcome)
	baseOutcome := BaseOutcome{}
	json.Unmarshal(j, &baseOutcome)

	keys, err = db.DecodeKeys(outcome.Prodiction.Hash)
	if err != nil {
		fmt.Println("error decoding keys", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	err = DB.AddItemToListByKeysInTable(baseOutcome, "prodictions", keys, eventTable)
	if err != nil {
		fmt.Println("error adding list to item", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, outcome)
}

func updateOutcome(c echo.Context) error {
	outcomeId := c.Param("id")
	outcome := Outcome{}
	//See if we can bind the input to an event
	if err := c.Bind(outcome); err != nil {
		return c.String(http.StatusBadRequest, "Invalid Object")
	}
	outcome.OutcomeID = outcomeId
	_, err := DB.PutItemInTable(outcome, outcomeTable)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, outcome)
}

func listOutcomes(c echo.Context) error {
	outcomes := []Outcome{}
	DB.ListItems(outcomeTable, &outcomes)
	return c.JSON(http.StatusOK, outcomes)
}

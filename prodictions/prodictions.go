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
	prodictionTable = "Prodictions"
)

// Prodiction The main item that drives the system
type Prodiction struct {
	EventID      string        `json:"eventId"`
	ProdictionID string        `json:"prodictionId"`
	Event        *BaseEvent    `json:"event,omitempty"`
	EventHash    string        `json:"eventHash,omitempty" dynamodbav:"-"`
	Outcomes     []BaseOutcome `json:"outcomes,omitempty"`
	Description  string        `json:"description"`
	Status       string        `json:"status,omitempty"`
	Hash         string        `json:"hash,omitempty"`
}

type BaseProdiction struct {
	Description string `json:"description"`
	Status      string `json:"status,omitempty"`
	Hash        string `json:"hash,omitempty"`
}

func init() {
	Server.GET("/prodictions", listProdictions)
	Server.POST("/prodictions", createProdiction)
	Server.PUT("/prodictions/:id", updateProdiction)
	Server.GET("/prodictions/:id", getProdiction)
}

// MarshalDynamoDBAttributeValue Custom Marshaller
func (prodiction *Prodiction) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	return db.CustomMarshalKeepEmptyLists(prodiction, []string{"outcomes"}, av)
}

func getProdiction(c echo.Context) error {
	keys, err := db.DecodeKeys(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	prodiction := Prodiction{}
	err = DB.GetItemByKeysFromTable(keys, prodictionTable, &prodiction)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, prodiction)
}
func createProdiction(c echo.Context) error {
	prodiction := new(Prodiction)

	if err := c.Bind(prodiction); err != nil {
		return c.String(http.StatusBadRequest, "Invalid Object")
	}
	//Retrieve Event
	eventKeys, err := db.DecodeKeys(prodiction.EventHash)
	if err != nil {
		fmt.Println("Error decoding keys")
		return c.JSON(http.StatusInternalServerError, err)
	}
	err = DB.GetItemByKeysFromTable(eventKeys, eventTable, &prodiction.Event)
	if err != nil {
		fmt.Println("Could find event")
		return c.JSON(http.StatusInternalServerError, err)
	}

	// Create UUID
	prodiction.ProdictionID = uuid.NewV4().String()
	prodiction.Outcomes = []BaseOutcome{}

	// Set Hash
	keys := []db.Key{}
	keys = append(keys, db.Key{
		Field: "eventId",
		Value: prodiction.EventID,
	})
	keys = append(keys, db.Key{
		Field: "prodictionId",
		Value: prodiction.ProdictionID,
	})
	encodedKeys, err := db.EncodeKeys(keys)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	prodiction.Hash = encodedKeys

	_, err = DB.PutItemInTable(prodiction, prodictionTable)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	//Now we need to add to Event
	//Convert Prodicition to BaseProdiction via json.(Un)Marshal
	j, _ := json.Marshal(prodiction)
	baseProdiction := BaseProdiction{}
	json.Unmarshal(j, &baseProdiction)

	keys, err = db.DecodeKeys(prodiction.Event.Hash)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	err = DB.AddItemToListByKeysInTable(baseProdiction, "prodictions", keys, eventTable)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, prodiction)
}

func updateProdiction(c echo.Context) error {
	prodictionId := c.Param("id")
	prodiction := new(Prodiction)
	//See if we can bind the input to an event
	if err := c.Bind(prodiction); err != nil {
		return c.String(http.StatusBadRequest, "Invalid Object")
	}
	prodiction.ProdictionID = prodictionId
	_, err := DB.PutItemInTable(prodiction, prodictionTable)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, prodiction)
}

func listProdictions(c echo.Context) error {
	prodictions := []Prodiction{}
	DB.ListItems(prodictionTable, &prodictions)
	return c.JSON(http.StatusOK, prodictions)
}

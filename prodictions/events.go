package prodictions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/labstack/echo"
	"github.com/marcato15/api.prodictions.com/db"
	uuid "github.com/satori/go.uuid"
)

var (
	eventTable = "Events"
)

// Event This is a basic container for a specfic real life event to house prodictions
type Event struct {
	EventId        string           `json:"eventId,omitempty"`
	EventGroupID   string           `json:"eventGroupId,omitempty"`
	EventGroup     *BaseEventGroup  `json:"eventGroup,omitempty"`
	Prodictions    []BaseProdiction `json:"prodictions,omitempty"`
	Description    string           `json:"description,omitempty"`
	Date           *time.Time       `json:"date,omitempty"`
	Status         string           `json:"status,omitempty"`
	Hash           string           `json:"hash,omitempty"`
	EventGroupHash string           `json:"eventGroupHash,omitempty" dyanamodbav:"-"`
}

type BaseEvent struct {
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
	Status      string     `json:"status,omitempty"`
	Hash        string     `json:"hash,omitempty"`
}

// MarshalDynamoDBAttributeValue Custom Marshaller
func (event *Event) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	return db.CustomMarshalKeepEmptyLists(event, []string{"prodictions"}, av)
}

func init() {

	//permissions.RegisterCustomPermission("GET", "/events", "eventgroups:list")

	Server.GET("/events", listEvents)
	Server.GET("/events/:id", getEvent)
	Server.PUT("/events/:id", updateEvent)
	Server.POST("/events/:hash", createEvent)
}

func getEvent(c echo.Context) error {
	//Extract and decode encrypted key
	keys, err := db.DecodeKeys(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	event := Event{}
	DB.GetItemByKeysFromTable(keys, eventTable, &event)
	return c.JSON(http.StatusOK, event)
}

func updateEvent(c echo.Context) error {
	eventId := c.Param("id")
	event := new(Event)
	//See if we can bind the input to an event
	if err := c.Bind(event); err != nil {
		return c.String(http.StatusBadRequest, "Invalid Object")
	}
	event.EventId = eventId
	_, err := DB.PutItemInTable(event, eventTable)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, event)
}

func createEvent(c echo.Context) error {
	event := new(Event)
	//See if we can bind the input to an event
	if err := c.Bind(event); err != nil {
		return c.String(http.StatusBadRequest, "Invalid Object")
	}
	//Retrieve Event Group
	eventGroupKeys, err := db.DecodeKeys(event.EventGroupHash)
	if err != nil {
		fmt.Println("Error decoding keys")
		return c.JSON(http.StatusInternalServerError, err)
	}
	err = DB.GetItemByKeysFromTable(eventGroupKeys, eventGroupTable, &event.EventGroup)
	if err != nil {
		fmt.Println("Could find event group")
		return c.JSON(http.StatusInternalServerError, err)
	}

	//Set UUID
	event.EventId = uuid.NewV4().String()

	// Set Hash
	keys := []db.Key{}
	keys = append(keys, db.Key{
		Field: "eventGroupId",
		Value: event.EventGroupID,
	})
	keys = append(keys, db.Key{
		Field: "eventId",
		Value: event.EventId,
	})
	encodedKeys, err := db.EncodeKeys(keys)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	event.Hash = encodedKeys

	_, err = DB.PutItemInTable(event, eventTable)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	//Now we need to add to Event Group
	//Convert Event to MiniEvent via json.(Un)Marshal
	j, _ := json.Marshal(event)
	baseEvent := BaseEvent{}
	json.Unmarshal(j, &baseEvent)

	keys, err = db.DecodeKeys(event.EventGroup.Hash)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	err = DB.AddItemToListByKeysInTable(baseEvent, "events", keys, eventGroupTable)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, event)
}

func listEvents(c echo.Context) error {
	events := []Event{}
	DB.ListItems(eventTable, &events)
	return c.JSON(http.StatusOK, events)
}

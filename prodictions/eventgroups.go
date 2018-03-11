package prodictions

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/labstack/echo"
	"github.com/marcato15/api.prodictions.com/db"
	uuid "github.com/satori/go.uuid"
)

var (
	eventGroupTable = "EventGroups"
)

// Event This is a basic container for a specfic real life event to house prodictions
type EventGroup struct {
	EventGroupID string      `json:"eventGroupId,omitempty"`
	Name         string      `json:"name,omitempty"`
	Description  string      `json:"description,omitempty"`
	Date         *time.Time  `json:"date,omitempty"`
	Status       string      `json:"status,omitempty"`
	Hash         string      `json:"hash,omitempty"`
	Events       []BaseEvent `json:"events,omitempty"`
}

type BaseEventGroup struct {
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
	Status      string     `json:"status,omitempty"`
	Hash        string     `json:"hash,omitempty"`
}

func init() {
	Server.GET("/eventgroups", listEventGroups)
	Server.GET("/eventgroups/:id", getEventGroup)
	Server.PUT("/eventgroups/:id", updateEvent)
	Server.POST("/eventgroups", createEventGroup)
}

// MarshalDynamoDBAttributeValue Custom Marshaller
func (eventGroup *EventGroup) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	return db.CustomMarshalKeepEmptyLists(eventGroup, []string{"events"}, av)
}

func getEventGroup(c echo.Context) error {
	keys, err := db.DecodeKeys(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	eventGroup := EventGroup{}
	err = DB.GetItemByKeysFromTable(keys, eventGroupTable, &eventGroup)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, eventGroup)
}

func createEventGroup(c echo.Context) error {
	eventGroup := new(EventGroup)

	if err := c.Bind(eventGroup); err != nil {
		return c.String(http.StatusBadRequest, "Invalid Object")
	}

	// Create UUID
	eventGroup.EventGroupID = uuid.NewV4().String()
	eventGroup.Events = []BaseEvent{}

	// Create Hash
	encodedKey, err := db.EncodeKey(db.Key{
		Field: "eventGroupId",
		Value: eventGroup.EventGroupID,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	eventGroup.Hash = encodedKey

	_, err = DB.PutItemInTable(eventGroup, eventGroupTable)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, eventGroup)
}

func updateEventGroup(c echo.Context) error {
	eventGroupId := c.Param("id")
	eventGroup := new(EventGroup)
	//See if we can bind the input to an event
	if err := c.Bind(eventGroup); err != nil {
		return c.String(http.StatusBadRequest, "Invalid Object")
	}
	eventGroup.EventGroupID = eventGroupId
	_, err := DB.PutItemInTable(eventGroup, eventGroupTable)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, eventGroup)
}

func listEventGroups(c echo.Context) error {
	eventGroups := []EventGroup{}
	DB.ListItems(eventGroupTable, &eventGroups)
	return c.JSON(http.StatusOK, eventGroups)
}

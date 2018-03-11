package prodictions

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/marcato15/api.prodictions.com/db"
	uuid "github.com/satori/go.uuid"
)

var (
	ticketTable = "Tickets"
)

// Ticket This is a prodictable ticket of a ticket. Stores the current line, but the line must not be used to pay out koins
type Ticket struct {
	Outcome     *Outcome `json:"outcome,omitempty" dynamodbav:"-"`
	CurrentLine string   `json:"currentLine,omitempty"`
	Status      string   `json:"status,omitempty"`
	Amount      int      `json:"amount,omitempty"`
	TicketID    string   `json:"ticketId"`
	UserID      string   `json:"userId,omitempty"`
	OutcomeID   string   `json:"outcomeId,omitempty"`
	User        *User    `json:"user,omitempty" dynamodbav:"-"`
	Hash        string   `json:"hash,omitempty"`
}

type BaseTicket struct {
	Outcome     *Outcome `json:"outcome,omitempty" dynamodbav:"-"`
	CurrentLine string   `json:"currentLine,omitempty"`
	Status      string   `json:"status,omitempty"`
	Amount      int      `json:"amount,omitempty"`
	Hash        string   `json:"hash,omitempty"`
}

func init() {
	Server.GET("/tickets", listTickets)
	Server.POST("/tickets", createTicket)
	//Server.PUT("/tickets", putTicket)
	Server.GET("/tickets/:id", getTicket)
}

func getTicket(c echo.Context) error {
	keys, err := db.DecodeKeys(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	ticket := Ticket{}
	err = DB.GetItemByKeysFromTable(keys, ticketTable, &ticket)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, ticket)
}

func createTicket(c echo.Context) error {
	ticket := new(Ticket)

	if err := c.Bind(ticket); err != nil {
		return c.String(http.StatusBadRequest, "Invalid Object")
	}

	// Create UUID
	ticket.TicketID = uuid.NewV4().String()

	// Set Hash
	keys := []db.Key{}
	keys = append(keys, db.Key{
		Field: "userId",
		Value: ticket.UserID,
	})
	keys = append(keys, db.Key{
		Field: "ticketId",
		Value: ticket.TicketID,
	})
	encodedKeys, err := db.EncodeKeys(keys)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	ticket.Hash = encodedKeys

	_, err = DB.PutItemInTable(ticket, ticketTable)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, ticket)
}

func listTickets(c echo.Context) error {
	tickets := []Ticket{}
	DB.ListItems(ticketTable, &tickets)
	return c.JSON(http.StatusOK, tickets)
}

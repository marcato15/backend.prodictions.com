package prodictions

import (
	"net/http"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/labstack/echo"
	"github.com/marcato15/api.prodictions.com/db"
	uuid "github.com/satori/go.uuid"
)

var (
	userTable = "Users"
)

// User This is a basic represenatation of a user in the system. Only the most basic info is stored here, and the rest is pulled from the identify provider
type User struct {
	UserID    string `json:"id"`
	Nickname  string `json:"nickname,omitempty"`
	KoinKount int    `json:"koinKount,omitempty"`
	Hash      string `json:"hash,omitempty"`
}
type PrivateUser struct {
	UserID    string   `json:"id"`
	Nickname  string   `json:"nickname,omitempty"`
	KoinKount int      `json:"koinKount,omitempty"`
	AuthID    string   `json:"authId,omitempty"`
	Tickets   []Ticket `json:"tickets,omitempty"`
	Hash      string   `json:"hash,omitempty"`
	//Koinlog []Koinlog `json:"koinlog,omitempty"`
}

func init() {
	Server.GET("/users", listUsers)
	Server.POST("/users", createUser)
	//Server.PUT("/users/:id", updateUser)
	Server.GET("/users/:id", getUser)
}

// MarshalDynamoDBAttributeValue Custom Marshaller
func (user *User) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	return db.CustomMarshalKeepEmptyLists(user, []string{"tickets"}, av)
}

func getUser(c echo.Context) error {
	//Extract and decode encrypted key
	keys, err := db.DecodeKeys(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	user := User{}
	DB.GetItemByKeysFromTable(keys, userTable, &user)
	return c.JSON(http.StatusOK, user)
}

func createUser(c echo.Context) error {
	user := new(User)
	//See if we can bind the input to an user
	if err := c.Bind(user); err != nil {
		return c.String(http.StatusBadRequest, "Invalid Object")
	}

	//Set UUID
	user.UserID = uuid.NewV4().String()

	// Set Hash
	keys := []db.Key{}
	keys = append(keys, db.Key{
		Field: "userId",
		Value: user.UserID,
	})
	encodedKeys, err := db.EncodeKeys(keys)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	user.Hash = encodedKeys

	_, err = DB.PutItemInTable(user, userTable)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, user)
}

func listUsers(c echo.Context) error {
	users := []User{}
	DB.ListItems(userTable, &users)
	return c.JSON(http.StatusOK, users)
}

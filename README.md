# mini
Provide easier way to make a micro service with Golang

#### Simple
[![](https://imgur.com/B1VJiNR.png)](https://imgur.com/B1VJiNR.png)

### Chain support
mini support chain with mostly all http methods also get data from params.
```go
m.Echo().GET("/api/users", getUsers)

func getUsers(c echo.Context) error {
	//-------------------------
	records := []*ModelUser{}
	if dbErr, chainErr := mini.With(c).
		WhereStringAdv("full_name", []string{"full_name", "ilike", "?"}, "%", "%").
		WhereInt("gender", []string{"gender", "=", "?"}).
		WhereDate("birthday", []string{"birthday", "=", "?"}).
		WhereStringAdv("email", []string{"email", "ilike", "?"}, "%", "%").
		WhereString("mobile", []string{"mobile", "ilike", "?"}).
		WhereString("type", []string{"type", "ilike", "?"}).
		Order("sort").
		Pagnation(&ModelUser{}).
		Find(&records); dbErr != nil || chainErr != nil {
		if dbErr != nil {
			return c.JSON(http.StatusInternalServerError, mini.DatabaseError(dbErr))
		}
		if chainErr != nil {
			return c.JSON(http.StatusBadRequest, mini.BadRequest(chainErr))
		}
	}
	//-------------------------
	return c.JSON(http.StatusOK, mini.ResponseJSON{
		Code:    1,
		Message: "Get records successfully",
		Data:    records,
	})
}
```

### Sync data support
 Sync data from object to object
```go
func editUser(c echo.Context) error {
	//-------------------------
	recordId, recIdErr := mini.With(c).On("path").UUID("id")
	if recIdErr != nil {
		return c.JSON(http.StatusBadRequest, mini.InvalidInput("id"))
	}
	record := ModelUser{}
	if mini.DB().First(&record, "id = ?", recordId).RecordNotFound() {
		return c.JSON(http.StatusInternalServerError, mini.NotFound())
	}
	//-------------------------
	userInput := new(struct {
		FullName     *string        `json:"full_name"`
		DisplayImage *uuid.UUID     `json:"display_image"`
		Gender       *int           `json:"gender"`
		Birthday     *mini.Datetime `json:"birthday"`
		Email        *string        `json:"email"`
	})
	if err := c.Bind(userInput); err != nil {
		return c.JSON(http.StatusBadRequest, mini.InvalidInput(err))
	}
	//-------------------------
	mini.Sync(*userInput, &record)
	if err := mini.DB().Save(&record).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, mini.DatabaseError(err))
	}
	//-------------------------
	return c.JSON(http.StatusOK, mini.ResponseJSON{
		Code:    1,
		Message: "Get record successfully",
	})
}
```

### Compute fields support for GORM

```go
type ModelUserTest struct {
	mini.ModelBase

	FullName     string `gorm:"not null;" json:"full_name" validate:"required"`
	RandomNumber int    `gorm:"-" json:"random_num" compute:"ComputeRandomNumber"`
}

func (m *ModelUserTest) ComputeRandomNumber() {
	m.RandomNumber = rand.Int()
}

func apiHandleSomething() {
	a := &ModelUserTest{}
	mini.Compute(a)
}

```


### And something else who know

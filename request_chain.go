package mini

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
)

func CurrentUser(c echo.Context) (uuid.UUID, error) {
	userIdStr := c.Request().Header.Get("x-user-id")
	return uuid.FromString(userIdStr)
}

func Pagnation(c echo.Context, query *gorm.DB, model interface{}) (*gorm.DB, error) {
	currentPage, perPage := PagnationQuery(c)
	totalRecords := 0
	if err := DB().Model(model).Count(&totalRecords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, DatabaseError(err))
		return query, err
	}
	lastPage := FloorOf(totalRecords, perPage)
	if lastPage*perPage < totalRecords {
		lastPage++
	}
	lastPage = Max(1, lastPage)
	currentPage = Min(currentPage, lastPage)
	nextPage := currentPage + 1
	if nextPage > lastPage {
		nextPage = 0
	}
	query = query.Limit(perPage).Offset((currentPage - 1) * perPage)
	// set header
	c.Response().Header().Add("X-Page", fmt.Sprint(currentPage))
	c.Response().Header().Add("X-Per-Page", fmt.Sprint(perPage))
	c.Response().Header().Add("X-Next-Page", fmt.Sprint(nextPage))
	c.Response().Header().Add("X-Last-Page", fmt.Sprint(lastPage))
	c.Response().Header().Add("X-Total-Items", fmt.Sprint(totalRecords))
	return query, nil
}

func PagnationQuery(c echo.Context) (int, int) {
	currentPage := 1
	perPage := 30
	if c.Request().Method == http.MethodGet || c.Request().Method == http.MethodDelete {
		currentPage = MustBeInt(strings.Trim(c.QueryParam("page"), " "))
		perPage = MustBeInt(strings.Trim(c.QueryParam("per_page"), " "))
	} else if c.Request().Method == http.MethodPost ||
		c.Request().Method == http.MethodPatch ||
		c.Request().Method == http.MethodPut {
		currentPage = MustBeInt(strings.Trim(c.FormValue("page"), " "))
		perPage = MustBeInt(strings.Trim(c.FormValue("per_page"), " "))
	}
	// validate
	if currentPage < 1 {
		currentPage = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 30
	}
	return currentPage, perPage
}

func HealthStatus(m *Mini, version string) {
	start := time.Now()
	m.Echo().GET("/api/status", func(c echo.Context) error {
		return c.JSON(http.StatusOK, ResponseJSON{
			Code:    1,
			Message: "Service is ok",
			Data: map[string]interface{}{
				"uptime":  time.Now().Sub(start).String(),
				"version": version,
			},
		})
	})
}

//=========================================================================
// Minimal Chain Data Getter For Echo Framework
// @by Namkazt
//=========================================================================
type MiniChain struct {
	ctx     echo.Context
	q       *gorm.DB
	err     error
	checkOn string
	updater map[string]interface{}
}

func With(c echo.Context) *MiniChain {
	return &MiniChain{ctx: c}
}
func (c *MiniChain) On(checkOn string) *MiniChain {
	c.checkOn = checkOn
	return c
}
func (c *MiniChain) Get(name string) string {
	if c.checkOn == "path" {
		return strings.Trim(c.ctx.Param(name), " ")
	} else if c.checkOn == "query" {
		return strings.Trim(c.ctx.QueryParam(name), " ")
	} else if c.checkOn == "form" {
		return strings.Trim(c.ctx.FormValue(name), " ")
	} else {
		if c.ctx.Request().Method == http.MethodGet || c.ctx.Request().Method == http.MethodDelete {
			return strings.Trim(c.ctx.QueryParam(name), " ")
		} else if c.ctx.Request().Method == http.MethodPost ||
			c.ctx.Request().Method == http.MethodPatch ||
			c.ctx.Request().Method == http.MethodPut {
			return strings.Trim(c.ctx.FormValue(name), " ")
		}
	}
	return ""
}
func (c *MiniChain) UUID(name string) (uuid.UUID, error) {
	v := c.Get(name)
	return uuid.FromString(v)
}
func (c *MiniChain) Int(name string) (int, error) {
	v := c.Get(name)
	return strconv.Atoi(v)
}
func (c *MiniChain) Float(name string) (float64, error) {
	v := c.Get(name)
	return strconv.ParseFloat(v, 64)
}
func (c *MiniChain) Date(name string) (time.Time, error) {
	v := c.Get(name)
	return time.Parse(time.RFC3339, v)
}
func (c *MiniChain) Time(name string) (time.Time, error) {
	v := c.Get(name)
	return time.Parse("15:04", v)
}
func (c *MiniChain) Bool(name string) bool {
	v := c.Get(name)
	return strings.ToLower(v) == "true" || v == "1"
}

//---------------------------------------------------------------
// Special methods apply directly to query
//---------------------------------------------------------------
func (c *MiniChain) DB(query *gorm.DB) *MiniChain {
	c.q = query
	return c
}

// ------------------------------------------------
func (c *MiniChain) Query() *gorm.DB {
	return c.q
}

// return DB error and MiniChain error if have
func (c *MiniChain) Preload(table string) *MiniChain {
	c.q = c.q.Preload(table)
	return c
}

// return DB error and MiniChain error if have
func (c *MiniChain) Find(output interface{}) (error, error) {
	return c.q.Find(output).Error, c.err
}

// return DB error and MiniChain error if have
func (c *MiniChain) Update(m interface{}) (error, error) {
	return c.q.Model(m).Update(c.updater).Error, c.err
}

// ------------------------------------------------

func (c *MiniChain) Order(name string) *MiniChain {
	if c.q == nil {
		c.q = DB()
	}
	v := c.Get(name)
	if v != "" {
		c.q = c.q.Order(v, true)
	}
	return c
}

func (c *MiniChain) Pagnation(model interface{}) *MiniChain {
	if c.err != nil {
		return c
	}
	if c.q == nil {
		c.q = DB()
	}
	currentPage, perPage := PagnationQuery(c.ctx)
	totalRecords := 0
	c.q.Model(model).Count(&totalRecords)
	lastPage := FloorOf(totalRecords, perPage)
	if lastPage*perPage < totalRecords {
		lastPage++
	}
	lastPage = Max(1, lastPage)
	currentPage = Min(currentPage, lastPage)
	nextPage := currentPage + 1
	if nextPage > lastPage {
		nextPage = 0
	}
	// add to query
	c.q = c.q.Limit(perPage).Offset((currentPage - 1) * perPage)
	// set header
	c.ctx.Response().Header().Add("X-Page", fmt.Sprint(currentPage))
	c.ctx.Response().Header().Add("X-Per-Page", fmt.Sprint(perPage))
	c.ctx.Response().Header().Add("X-Next-Page", fmt.Sprint(nextPage))
	c.ctx.Response().Header().Add("X-Last-Page", fmt.Sprint(lastPage))
	c.ctx.Response().Header().Add("X-Total-Items", fmt.Sprint(totalRecords))
	return c
}

//---------------------------------------------------------------
// Almighty combine methods
//---------------------------------------------------------------
func (c *MiniChain) ValidateMiniChain() *MiniChain {
	if c.err != nil {
		return c
	}
	if c.q == nil {
		c.q = DB()
	}
	return nil
}

// Example: searchUUID("product_id", []string{"product_id", "=", "?"})
func (c *MiniChain) WhereUUID(name string, domain []string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if c.Get(name) != "" {
		if v, err := c.UUID(name); err != nil {
			c.err = err
		} else {
			c.q = c.q.Where(strings.Join(domain, " "), v)
		}
	}
	return c
}

func (c *MiniChain) WhereString(name string, domain []string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if v := c.Get(name); v != "" {
		c.q = c.q.Where(strings.Join(domain, " "), v)
	}
	return c
}

func (c *MiniChain) WhereStringAdv(name string, domain []string, pre string, sub string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if v := c.Get(name); v != "" {
		c.q = c.q.Where(strings.Join(domain, " "), pre+v+sub)
	}
	return c
}

func (c *MiniChain) WhereDate(name string, domain []string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if c.Get(name) != "" {
		if v, err := c.Date(name); err != nil {
			c.err = err
		} else {
			c.q = c.q.Where(strings.Join(domain, " "), v)
		}
	}
	return c
}

func (c *MiniChain) WhereTime(name string, domain []string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if c.Get(name) != "" {
		if _, err := c.Time(name); err != nil {
			c.err = err
		} else {
			c.q = c.q.Where(strings.Join(domain, " "), c.Get(name))
		}
	}
	return c
}

func (c *MiniChain) WhereInt(name string, domain []string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if c.Get(name) != "" {
		if v, err := c.Int(name); err != nil {
			c.err = err
		} else {
			c.q = c.q.Where(strings.Join(domain, " "), v)
		}
	}
	return c
}

func (c *MiniChain) WhereFloat(name string, domain []string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if c.Get(name) != "" {
		if v, err := c.Float(name); err != nil {
			c.err = err
		} else {
			c.q = c.q.Where(strings.Join(domain, " "), v)
		}
	}
	return c
}

func (c *MiniChain) WhereBool(name string, domain []string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if c.Get(name) != "" {
		c.q = c.q.Where(strings.Join(domain, " "), c.Bool(name))
	}
	return c
}

//---------------------------------------------------------------
// Almighty combine methods
//---------------------------------------------------------------

func (c *MiniChain) UpdateUUIDDirect(name string, v uuid.UUID) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	c.updater[name] = v
	return c
}

func (c *MiniChain) UpdateUUID(name string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if c.Get(name) != "" {
		if v, err := c.UUID(name); err != nil {
			c.err = err
		} else {
			c.updater[name] = v
		}
	}
	return c
}

func (c *MiniChain) UpdateString(name string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if c.updater == nil {
		c.updater = make(map[string]interface{})
	}
	if c.Get(name) != "" {
		if v := c.Get(name); v != "" {
			c.updater[name] = v
		}
	}
	return c
}

func (c *MiniChain) UpdateDate(name string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if c.updater == nil {
		c.updater = make(map[string]interface{})
	}
	if c.Get(name) != "" {
		if v, err := c.Date(name); err != nil {
			c.err = err
		} else {
			c.updater[name] = v
		}
	}
	return c
}

func (c *MiniChain) UpdateTime(name string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if c.updater == nil {
		c.updater = make(map[string]interface{})
	}
	if c.Get(name) != "" {
		if _, err := c.Time(name); err == nil {
			c.updater[name] = c.Get(name)
		} else {
			c.err = err
		}
	}
	return c
}

func (c *MiniChain) UpdateInt(name string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if c.updater == nil {
		c.updater = make(map[string]interface{})
	}
	if c.Get(name) != "" {
		if v, err := c.Int(name); err != nil {
			c.err = err
		} else {
			c.updater[name] = v
		}
	}
	return c
}

func (c *MiniChain) UpdateBool(name string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if c.updater == nil {
		c.updater = make(map[string]interface{})
	}
	if c.Get(name) != "" {
		c.updater[name] = c.Bool(name)
	}
	return c
}

func (c *MiniChain) UpdateJsonB(name string) *MiniChain {
	if c := c.ValidateMiniChain(); c != nil {
		return c
	}
	if c.updater == nil {
		c.updater = make(map[string]interface{})
	}
	if v := c.Get(name); v != "" {
		c.updater[name] = Jsonb(v)
	}
	return c
}

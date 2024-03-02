package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"interview_test/src/app/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/patrickmn/go-cache"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "interview_test"
)

type HistoryController struct {
	Cache *cache.Cache
	DB    *sql.DB
}

func NewHistoryController() *HistoryController {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname))
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	// Test the database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging the database: %v", err)
	}
	return &HistoryController{
		Cache: cache.New(1*time.Second, 24*time.Hour), // Adjust expiration and cleanup intervals as needed
		DB:    db,
	}
}

func (c *HistoryController) CloseDB() {
	c.DB.Close()
}

func (c *HistoryController) GetHistories(ctx *gin.Context) {
	input := models.Input{}
	if err := ctx.BindJSON(&input); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	histories, err := FetchHistoricalPrices(c, input)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err := c.InsertHistory(histories, input); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, histories)
}

func (c *HistoryController) InsertHistory(histories models.PriceHistory, input models.Input) error {
	query := `INSERT INTO market_price_histories (symbol, low, high, open, close, change) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := c.DB.Exec(query, input.Symbol, histories.Low, histories.High, histories.Open, histories.Close, histories.Change)
	if err != nil {
		return err
	}
	return nil
}

func callAPICoingecko(input models.Input) (*models.ResponseData, error) {
	query := `https://api.coingecko.com/api/v3/coins/%v/market_chart/range?vs_currency=%v&from=%v&to=%v`
	url := fmt.Sprintf(query, input.Symbol, "usd", input.Start.Unix(), input.End.Unix())

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var data models.ResponseData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %w", err)
	}
	return &data, nil
}

func FetchHistoricalPrices(c *HistoryController, input models.Input) (models.PriceHistory, error) {
	cacheKey := fmt.Sprintf("%v_%v_%v", input.Symbol, input.Start.Unix(), input.End.Unix())

	if item, found := c.Cache.Get(cacheKey); found {
		result, ok := item.(models.PriceHistory)
		if ok {
			return result, nil
		}
	}

	data, err := callAPICoingecko(input)
	if err != nil {
		return models.PriceHistory{}, err
	}

	var high, low, open, close, closePrevious float64

	query := `
	SELECT mph.id, mph.change
	FROM market_price_histories mph
	JOIN (
		SELECT MAX(id) AS max_id, change
		FROM market_price_histories
		GROUP BY change
	) AS max_ids
	ON mph.id = max_ids.max_id`

	for i, price := range data.Prices {
		if i == 0 {
			high, low, open = price[1], price[1], price[1]
		} else {
			if price[1] > high {
				high = price[1]
			}
			if price[1] < low {
				low = price[1]
			}
		}
		close = price[1]
	}
	c.DB.QueryRow(query).Scan(&closePrevious)
	var percentChange float64
	if closePrevious == 0 {
		closePrevious = close
	}
	percentChange = (closePrevious / close) / close * 100
	timeResp := int64(data.Prices[0][0])

	result := models.PriceHistory{
		Open:   open,
		Close:  close,
		High:   high,
		Low:    low,
		Time:   timeResp,
		Change: percentChange,
	}

	expiration := cache.DefaultExpiration
	if input.Period == "30M" {
		expiration = 30 * time.Minute
	} else if input.Period == "1H" {
		expiration = 1 * time.Hour
	} else if input.Period == "1D" {
		expiration = 24 * time.Hour
	} else if input.Period == "1S" {
		expiration = 1 * time.Second
	}

	c.Cache.Set(cacheKey, result, expiration)

	return result, nil
}

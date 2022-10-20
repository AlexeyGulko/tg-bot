package currency

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"golang.org/x/net/context"
	"golang.org/x/text/encoding/charmap"
)

type list struct {
	Currencies []Currency `xml:"Valute"`
}

type Currency struct {
	Code string `xml:"CharCode"`
	Rate string `xml:"Value"`
}

type Client struct{}

func New() *Client {
	return &Client{}
}

func (c *Client) GetExchangeRates(ctx context.Context, date time.Time) ([]dto.Currency, error) {
	ctx, cancelCtx := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancelCtx()

	url := fmt.Sprintf("https://www.cbr.ru/scripts/XML_daily.asp?date_req=%s", date.Format("02/01/2006"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		log.Print(err.Error())
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Print(err.Error())
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)

	m := list{Currencies: make([]Currency, 0, 34)}
	reader := bytes.NewReader(body)
	decoder := xml.NewDecoder(reader)

	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch charset {
		case "windows-1251":
			return charmap.Windows1251.NewDecoder().Reader(input), nil
		default:
			return nil, fmt.Errorf("unknown charset: %s", charset)
		}
	}
	err = decoder.Decode(&m)

	if err != nil {
		return nil, err
	}
	dtoMap := make([]dto.Currency, 0, 34)

	for _, v := range m.Currencies {
		rate, err := decimal.NewFromString(strings.Replace(v.Rate, ",", ".", 1))
		if err != nil {
			return nil, err
		}
		dtoMap = append(dtoMap, dto.Currency{Code: v.Code, Rate: rate})
	}

	return dtoMap, nil
}

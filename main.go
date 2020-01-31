package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// Articles all articles
type Articles struct {
	StatusCode int          `json:"httpStatus"`
	Articles   ArticleItems `json:"response"`
}

// ArticleItems root node for articles
type ArticleItems struct {
	Items []ArticleItem `json:"items"`
}

// ArticleItem single item for article
type ArticleItem struct {
	Articletype  string  `json:"type"`
	HarvesterID  string  `json:"harvesterId"`
	CerebroScore float32 `json:"cerebro-score"`
	URL          string  `json:"url"`
	Title        string  `json:"title"`
	CleanImage   string  `json:"cleanImage"`
}

// Advs all advertisements
type Advs struct {
	StatusCode int      `json:"httpStatus"`
	Advs       AdvItems `json:"response"`
}

// AdvItems root node for adv items
type AdvItems struct {
	Items []AdvItem `json:"items"`
}

// AdvItem struct for storing adv items
type AdvItem struct {
	AdvType           string  `json:"type"`
	HarvesterID       string  `json:"harvesterId"`
	CommercialPartner string  `json:"commercialPartner"`
	LogoURL           string  `json:"logoURL"`
	CerebroScore      float32 `json:"cerebro-score"`
	URL               string  `json:"url"`
	Title             string  `json:"title"`
	CleanImage        string  `json:"cleanImage"`
}

// DefaultAdv default value of advertising, when all advertisements were shown
type DefaultAdv struct {
	Type string `json:"type"`
}

// Collection of application errors
type appErrors struct {
	ErrorContent []appError `json:"error"`
}

// Single application error
type appError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Final structure for response
type responseItems struct {
	StatusCode int            `json:"httpStatus"`
	Content    []ResponseItem `json:"content"`
}

// JSONFillable interface of structure types
type JSONFillable interface {
	fillStructure(byteContents []byte) error
}

// Making Articles{} the type of JSONFillable
func (articles *Articles) fillStructure(byteContents []byte) error {
	return json.Unmarshal(byteContents, &articles)
}

// Making Advs{} the type of JSONFillable
func (advertizements *Advs) fillStructure(byteContents []byte) error {
	return json.Unmarshal(byteContents, &advertizements)
}

// Just adding error to appErrors{} collection to avoid of using global variable for error tracking
func (appErrors *appErrors) addError(code int, err error) {
	errorToAdd := appError{Code: code, Message: err.Error()}
	appErrors.ErrorContent = append(appErrors.ErrorContent, errorToAdd)
}

// ResponseItem interface for item of response
type ResponseItem interface {
	getAdvItemByKey()
}

// Type casting default advertising struct for sending response
func (defaultAd DefaultAdv) getAdvItemByKey() {
}

// Type casting advertising item struct for sending response
func (advItem AdvItem) getAdvItemByKey() {
}

// Type casting default article item struct for sending response
func (articleItem ArticleItem) getAdvItemByKey() {
}

// Initializing and checking required variables
func init() {
	if err := godotenv.Load(); err != nil {
		throwFatalError(fmt.Errorf("No .env file found"))
	}

	adsFrequency, err := strconv.Atoi(os.Getenv("ADS_FREQUENCY"))
	if (adsFrequency == 0) || (err != nil) {
		throwFatalError(fmt.Errorf("ADS_FREQUENCY can not be converted to integer more then 0"))
	}

}

func main() {
	http.HandleFunc(os.Getenv("REQUEST_URL"), getContent)
	log.Fatal(http.ListenAndServe(os.Getenv("SERVER_HOST"), nil))
}

// Handler for getting mixed content of articles and advertising items
// If writing data to structs was successefull (empty errors.ErrorContent) we are trying to combine content
// Otherwise we respond with marshalled errors content
func getContent(w http.ResponseWriter, req *http.Request) {
	articles, ads, errors := setStructs()
	var responseJSON = make([]byte, 0)
	var err error

	if len(errors.ErrorContent) == 0 {
		responseJSON, err = combineContentWithAds(articles, ads)
		if err != nil {
			errors.addError(502, fmt.Errorf("Error while combining content with ads: %s", err.Error()))
		}
	}

	if len(errors.ErrorContent) > 0 {
		responseJSON, _ = json.Marshal(appErrors{ErrorContent: errors.ErrorContent})
	}

	fmt.Fprintf(w, string(responseJSON))
}

// For every ADS_FREQUENCY articles we publish one advertising item.
// If no adv. item found we are publishing default advertising item (DefaultAdv{})
func combineContentWithAds(articles *Articles, ads *Advs) ([]byte, error) {
	responseItemsSlice := make([]ResponseItem, 0)
	adsFrequency, _ := strconv.Atoi(os.Getenv("ADS_FREQUENCY"))

	for i, article := range articles.Articles.Items {
		responseItemsSlice = append(responseItemsSlice, article)
		if (i+1)%adsFrequency == 0 {
			responseItemsSlice = append(responseItemsSlice, ads.Advs.getAdvertise(int((i+1)/adsFrequency)-1))
		}
	}

	return json.Marshal(responseItems{
		StatusCode: 200,
		Content:    responseItemsSlice,
	})
}

// getAdvertise getting advertise in AdvItems{} by key. If nothing found return default advertising
func (advItems *AdvItems) getAdvertise(key int) ResponseItem {
	if key < (len(advItems.Items)) {
		return advItems.Items[key]
	}

	return DefaultAdv{Type: os.Getenv("ADV_DEFAULT_TEXT")}
}

// Initializing Articles{} and Advs{} in goroutines with its content via http
func setStructs() (*Articles, *Advs, *appErrors) {
	articles := new(Articles)
	ads := new(Advs)
	appErrors := new(appErrors)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		writeDataToStruct(os.Getenv("ARTICLES_ARCHIVE"), articles, appErrors)
		wg.Done()
	}()

	go func() {
		writeDataToStruct(os.Getenv("ADS_ARCHIVE"), ads, appErrors)
		wg.Done()
	}()

	wg.Wait()

	return articles, ads, appErrors
}

// Writing data to structures with error handling
func writeDataToStruct(url string, structToWrite JSONFillable, appErrors *appErrors) {
	r, err := getHTTPBytes(url)
	if err != nil {
		appErrors.addError(503, fmt.Errorf("Error of getting content for %s structure: %s", reflect.TypeOf(structToWrite).String(), err.Error()))
	}

	err = prepareStruct(r, structToWrite)
	if err != nil {
		appErrors.addError(503, fmt.Errorf("Error of preparing structure %s: %s", reflect.TypeOf(structToWrite).String(), err.Error()))
	}
}

// Getting json via http
func getHTTPBytes(url string) ([]byte, error) {
	client := http.Client{Timeout: 3 * time.Second}
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return ioutil.ReadAll(r.Body)
}

// Trying to write data into structure of type JSONFillable
// with controlling of its result
func prepareStruct(bytesContent []byte, fillableStruct JSONFillable) error {
	if !json.Valid(bytesContent) {
		return fmt.Errorf("Json string is not a valid json")
	}

	err := fillableStruct.fillStructure(bytesContent)
	if err != nil {
		return err
	}

	return nil
}

// Throwing fatal error to avoid server starting
func throwFatalError(err error) {
	if err != nil {
		log.Fatal("Logging fatal ", err)
	}
}

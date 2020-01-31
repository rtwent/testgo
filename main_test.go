package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestMarshalArticlesValidJson(t *testing.T) {
	err := articlesJSONConditions(os.Getenv("TESTING_ARTICLES_ARCHIVE"), t)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestGetAdvertise(t *testing.T) {
	ads := new(Advs)
	jsonContent, _ := os.Open(os.Getenv("TESTING_ADS_ARCHIVE"))
	jsonBytes, _ := ioutil.ReadAll(jsonContent)
	prepareStruct(jsonBytes, ads)
	item := ads.Advs.getAdvertise(1)
	_, ok := item.(AdvItem)
	if !ok {
		t.Fatal("Instance of AdvItemExpected")
	}

	item = ads.Advs.getAdvertise(5)
	_, ok = item.(DefaultAdv)
	if !ok {
		t.Fatal("Instance of DefaultAdv")
	}
}

func TestMarshalArticlesAbsentJson(t *testing.T) {
	err := articlesJSONConditions("mocks/none.json", t)
	if err == nil {
		t.Fatal("Error expected, but got nothing")
	}
}

func TestMarshalArticlesBrokenJson(t *testing.T) {
	err := articlesJSONConditions(os.Getenv("TESTING_ARTICLES_BROKEN_ARCHIVE"), t)
	if err == nil {
		t.Fatal("Error expected, but got nothing")
	}
}

func TestAdvertiseGet(t *testing.T) {
	err := advJSONConditions(os.Getenv("TESTING_ADS_ARCHIVE"), t)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func advJSONConditions(mockPath string, t *testing.T) error {
	ads := new(Advs)
	jsonContent, err := os.Open(mockPath)
	if err != nil {
		return fmt.Errorf("Can not open file for testing")
	}

	jsonBytes, err := ioutil.ReadAll(jsonContent)
	if err != nil {
		return fmt.Errorf("Can not read Json content")
	}

	err = prepareStruct(jsonBytes, ads)
	if err != nil {
		return err
	}
	if ads.StatusCode != 200 {
		return fmt.Errorf("Response must have 200 code")
	}

	return nil
}

// articlesJSONConditions checking articles by mock file
func articlesJSONConditions(mockPath string, t *testing.T) error {
	jsonContent, err := os.Open(mockPath)
	if err != nil {
		return fmt.Errorf("Can not open file for testing")
	}
	jsonBytes, err := ioutil.ReadAll(jsonContent)
	if err != nil {
		return fmt.Errorf("Can not read Json content")
	}

	articles := new(Articles)
	err = prepareStruct(jsonBytes, articles)

	if err != nil {
		return err
	}
	if articles.StatusCode != 200 {
		return fmt.Errorf(fmt.Sprintf("Expecting code 200, but got %d", articles.StatusCode))
	}

	if len(articles.Articles.Items) != 136 {
		return fmt.Errorf(fmt.Sprintf("Expecting length of items  equal to 136, but got %d", len(articles.Articles.Items)))
	}

	for key, item := range articles.Articles.Items {
		if key == 0 {
			if item.Articletype != "Article" {
				return fmt.Errorf(fmt.Sprintf("Expecting value as Article, but got %s", item.Articletype))
			}
		}
	}

	return nil
}

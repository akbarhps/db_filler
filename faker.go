package main

import (
	"fmt"
	"github.com/jaswdr/faker"
	"go.bryk.io/pkg/ulid"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const (
	MaxUniqueIteration = 1000
	MaxLoremParagraph  = 100
	MaxLoremSentence   = 15
)

type Faker struct {
	faker.Faker
	sync.Mutex
	uniqueCacheMutex sync.Mutex
	uniqueCache      map[string]bool
}

func NewFaker() *Faker {
	return &Faker{
		Faker:       faker.New(),
		uniqueCache: make(map[string]bool),
	}
}

func (f *Faker) isCached(key string, value string) bool {
	isCached, _ := f.uniqueCache[GetCacheKey(key, value)]
	return isCached
}

func (f *Faker) setCache(key string, value string) {
	f.uniqueCache[GetCacheKey(key, value)] = true
}

func (f *Faker) ClearCache() {
	f.Lock()
	defer f.Unlock()

	f.uniqueCache = make(map[string]bool)
}

func (f *Faker) Get(key string) string {
	// this lock prevent faker random generator to be called concurrently
	// which can have unexpected behavior
	f.Lock()
	defer f.Unlock()

	switch key {
	case "address":
		return f.Address().Address()
	case "address_building_number":
		return f.Address().BuildingNumber()
	case "address_city":
		return f.Address().City()
	case "address_city_prefix":
		return f.Address().CityPrefix()
	case "address_city_suffix":
		return f.Address().CitySuffix()
	case "address_country":
		return f.Address().Country()
	case "address_country_abbr":
		return f.Address().CountryAbbr()
	case "address_country_code":
		return f.Address().CountryCode()
	case "address_latitude":
		return fmt.Sprintf("%f", f.Address().Latitude())
	case "address_longitude":
		return fmt.Sprintf("%f", f.Address().Longitude())
	case "address_post_code":
		return f.Address().PostCode()
	case "address_secondary_address":
		return f.Address().SecondaryAddress()
	case "address_state":
		return f.Address().State()
	case "address_state_abbr":
		return f.Address().StateAbbr()
	case "address_street_address":
		return f.Address().StreetAddress()
	case "address_street_name":
		return f.Address().StreetName()
	case "address_street_suffix":
		return f.Address().StreetSuffix()

	case "app_name":
		return f.App().Name()
	case "app_version":
		return f.App().Version()

	case "bool":
		return strconv.FormatBool(f.Bool())

	case "color_css":
		return f.Color().CSS()
	case "color":
		return f.Color().ColorName()
	case "color_hex":
		return f.Color().Hex()
	case "color_rgb":
		return f.Color().RGB()
	case "safe_color_name":
		return f.Color().SafeColorName()

	case "company_bs":
		return f.Company().BS()
	case "company_catch_phrase":
		return f.Company().CatchPhrase()
	case "company_ein":
		return fmt.Sprintf("%d", f.Company().EIN())
	case "company_job_title":
		return f.Company().JobTitle()
	case "company_name":
		return f.Company().Name()
	case "company_suffix":
		return f.Company().Suffix()

	case "currency_code":
		return f.Currency().Code()
	case "currency_country":
		return f.Currency().Country()
	case "currency":
		return f.Currency().Currency()

	case "net_company_email":
		return f.Internet().CompanyEmail()
	case "net_domain":
		return f.Internet().Domain()
	case "net_email":
		return f.Internet().Email()
	case "net_free_email":
		return f.Internet().FreeEmail()
	case "net_free_email_domain":
		return f.Internet().FreeEmailDomain()
	case "net_http_method":
		return f.Internet().HTTPMethod()
	case "net_ipv4":
		return f.Internet().Ipv4()
	case "net_ipv6":
		return f.Internet().Ipv6()
	case "net_mac_address":
		return f.Internet().MacAddress()
	case "net_password":
		return f.Internet().Password()
	case "net_query":
		return f.Internet().Query()
	case "net_safe_email":
		return f.Internet().SafeEmail()
	case "net_safe_email_domain":
		return f.Internet().SafeEmailDomain()
	case "net_slug":
		return f.Internet().Slug()
	case "net_status_code":
		return fmt.Sprintf("%d", f.Internet().StatusCode())
	case "net_status_code_message":
		return f.Internet().StatusCodeMessage()
	case "net_tld":
		return f.Internet().TLD()
	case "net_url":
		return f.Internet().URL()

	case "lang":
		return f.Language().Language()
	case "lang_abbr":
		return f.Language().LanguageAbbr()
	case "lang_programming":
		return f.Language().ProgrammingLanguage()

	case "lorem_paragraph":
		return f.Lorem().Paragraph(rand.Intn(MaxLoremParagraph))
	case "lorem_sentence":
		return f.Lorem().Sentence(rand.Intn(MaxLoremSentence))
	case "lorem_word":
		return f.Lorem().Word()

	case "person_name":
		return f.Person().Name()
	case "person_first_name":
		return f.Person().FirstName()
	case "person_first_name_male":
		return f.Person().FirstNameMale()
	case "person_first_name_female":
		return f.Person().FirstNameFemale()
	case "person_gender":
		return f.Person().Gender()
	case "person_last_name":
		return f.Person().LastName()
	case "person_name_male":
		return f.Person().NameMale()
	case "person_name_female":
		return f.Person().NameFemale()
	case "person_ssn":
		return f.Person().SSN()
	case "person_suffix":
		return f.Person().Suffix()
	case "person_title":
		return f.Person().Title()

	case "phone_area_code":
		return f.Phone().AreaCode()
	case "phone_exchange_code":
		return f.Phone().ExchangeCode()
	case "phone_number":
		return f.Phone().Number()

	case "time_ansic":
		return f.Time().ANSIC(time.Now())
	case "time_am_pm":
		return f.Time().AmPm()
	case "time_century":
		return f.Time().Century()
	case "time_day_of_month":
		return fmt.Sprintf("%d", f.Time().DayOfMonth())
	case "time_day_of_week":
		return fmt.Sprintf("%d", f.Time().DayOfWeek())
	case "time_iso8601":
		return f.Time().ISO8601(time.Now())
	case "time_kitchen":
		return f.Time().Kitchen(time.Now())
	case "time_month":
		return f.Time().Month().String()
	case "time_month_name":
		return f.Time().MonthName()
	case "time":
		return f.Time().Time(time.Now()).String()
	case "timezone":
		return f.Time().Timezone()
	case "time_unix":
		return fmt.Sprintf("%d", f.Time().Unix(time.Now()))
	case "time_unix_date":
		return f.Time().UnixDate(time.Now())
	case "time_year":
		return fmt.Sprintf("%d", f.Time().Year())
	case "time_date":
		return f.Time().Time(time.Now()).Format("2006-01-02")
	case "timestamp":
		return f.Time().Time(time.Now()).Format("2006-01-02 15:04:05")

	case "id":
	case "uuid":
		return f.UUID().V4()
	case "ulid":
		id, _ := ulid.New()
		return id.String()
	}

	panic(fmt.Sprintf("Unsupported faker key: %s", key))
}

func (f *Faker) GetUnique(key string) string {
	// have to make separate mutex for unique cache to avoid deadlock
	f.uniqueCacheMutex.Lock()
	defer f.uniqueCacheMutex.Unlock()

	for i := 0; i < MaxUniqueIteration; i++ {
		value := f.Get(key)
		if f.isCached(key, value) {
			continue
		}

		f.setCache(key, value)
		return value
	}

	panic(fmt.Sprintf("Unable to generate unique value after %d iteration for key: %s", MaxUniqueIteration, key))
}

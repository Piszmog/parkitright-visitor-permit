package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"log"
	"os"
	"path/filepath"
	"time"
)

type resident struct {
	PropertyName    string `json:"property_name"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	StreetAddress   string `json:"street_address"`
	ApartmentNumber string `json:"apartment_number"`
	City            string `json:"city"`
	State           string `json:"state"`
	ZipCode         string `json:"zipcode"`
}

type visitor struct {
	FirstName       string  `json:"first_name"`
	LastName        string  `json:"last_name"`
	EmailAddress    string  `json:"email_address"`
	PhoneNumber     string  `json:"phone_number"`
	StreetAddress   string  `json:"street_address"`
	ApartmentNumber string  `json:"apartment_number"`
	City            string  `json:"city"`
	ZipCode         string  `json:"zipcode"`
	Vehicle         Vehicle `json:"vehicle"`
}

type Vehicle struct {
	Year                    string `json:"year"`
	Make                    string `json:"make"`
	Model                   string `json:"model"`
	Color                   string `json:"color"`
	LicencePlateNumber      string `json:"licence_plate_number"`
	LicensePlateStateIssuer string `json:"license_plate_state_issuer"`
}

func main() {
	//
	// Input Data
	//
	resident, err := getResident("resident.json")
	if err != nil {
		log.Fatal(err)
	}
	visitor, err := getVisitor("visitor.json")
	if err != nil {
		log.Fatal(err)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false))
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	// create context
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// run task list
	err = chromedp.Run(ctx,
		chromedp.Navigate(`https://www.parkitrightpermit.com/park-it-right-contact-visitor-permit-request/`),
		chromedp.WaitVisible(createSelector("property-name")),
		// Resident
		chromedp.SendKeys(createSelector("property-name"), resident.PropertyName),
		chromedp.SendKeys(createSelector("first-name-of-resident"), resident.FirstName),
		chromedp.SendKeys(createSelector("last-name-of-resident"), resident.LastName),
		chromedp.SendKeys(createSelector("resident-address"), resident.StreetAddress),
		chromedp.SendKeys(createSelector("resident-apartment"), resident.ApartmentNumber),
		chromedp.SendKeys(createSelector("resident-city"), resident.City),
		chromedp.SendKeys(createSelector("resident-state"), resident.State),
		chromedp.SendKeys(createSelector("resident-zip"), resident.ZipCode),
		// Visitor
		chromedp.SendKeys(createSelector("visitor-first-name"), visitor.FirstName),
		chromedp.SendKeys(createSelector("visitor-last-name"), visitor.LastName),
		chromedp.SendKeys(createSelector("visitor-email"), visitor.EmailAddress),
		chromedp.SendKeys(createSelector("visitor-phone"), visitor.PhoneNumber),
		chromedp.SendKeys(createSelector("visitor-address"), visitor.StreetAddress),
		chromedp.SendKeys(createSelector("visitor-apt-number"), visitor.ApartmentNumber),
		chromedp.SendKeys(createSelector("visitor-city"), visitor.City),
		chromedp.SendKeys(createSelector("visitor-zip"), visitor.ZipCode),
		chromedp.SendKeys(createSelector("visitor-year"), visitor.Vehicle.Year),
		chromedp.SendKeys(createSelector("visitor-make"), visitor.Vehicle.Make),
		chromedp.SendKeys(createSelector("visitor-model"), visitor.Vehicle.Model),
		chromedp.SendKeys(createSelector("visitor-color"), visitor.Vehicle.Color),
		chromedp.SendKeys(createSelector("visitor-license-plate-number"), visitor.Vehicle.LicencePlateNumber),
		chromedp.SendKeys(createSelector("visitor-state-of-issuance"), visitor.Vehicle.LicensePlateStateIssuer),
		// Send
		chromedp.Submit("visitors"),
		// Sleep
		chromedp.Sleep(20*time.Second),
	)
	if err != nil {
		log.Println(err)
		return
	}
}

func getResident(filePath string) (resident, error) {
	file, err := openFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile(file)
	var resident resident
	err = json.NewDecoder(file).Decode(&resident)
	if err != nil {
		log.Fatal(err)
	}
	resident.ApartmentNumber = getApartmentNumber(resident.ApartmentNumber)
	return resident, err
}

func getVisitor(filePath string) (visitor, error) {
	file, err := openFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile(file)
	var visitor visitor
	err = json.NewDecoder(file).Decode(&visitor)
	if err != nil {
		log.Fatal(err)
	}
	visitor.ApartmentNumber = getApartmentNumber(visitor.ApartmentNumber)
	return visitor, err
}

func getApartmentNumber(aptNumber string) string {
	apartmentNumber := aptNumber
	if len(apartmentNumber) == 0 {
		apartmentNumber = "N/A"
	}
	return apartmentNumber
}

func openFile(filename string) (*os.File, error) {
	pathToFile, err := filepath.Abs(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get absolute path of %s", filename)
	}
	file, err := os.Open(pathToFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file %s", filename)
	}
	return file, nil
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Fatal(errors.Wrapf(err, "failed to close %s", file.Name()))
	}
}

func createSelector(propertyName string) string {
	return fmt.Sprintf(`//input[@name="%s"]`, propertyName)
}

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"math"
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
	residentJSONFile := flag.String("r", "", "Resident JSON file")
	visitorJSONFile := flag.String("v", "", "Visitor JSON file")
	flag.Parse()
	if len(*residentJSONFile) == 0 {
		log.Println("Resident file -r is required")
		flag.Usage()
		return
	} else if len(*visitorJSONFile) == 0 {
		log.Println("Visitor file -v is required")
		flag.Usage()
		return
	}
	//
	// Input Data
	//
	resident, err := getResident(*residentJSONFile)
	if err != nil {
		log.Fatal(err)
	}
	if err = validateResident(resident); err != nil {
		log.Fatal(err)
	}
	visitor, err := getVisitor(*visitorJSONFile)
	if err != nil {
		log.Fatal(err)
	}
	if err = validateVisitor(visitor); err != nil {
		log.Fatal(err)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false))
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	// create context
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// run task list
	var detailsBuf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate(`https://www.parkitrightpermit.com/park-it-right-contact-visitor-permit-request/`),
		chromedp.WaitVisible(createSelector("property-name")),
		chromedp.Sleep(2*time.Second),
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
		// Screenshot details
		chromedp.ActionFunc(func(ctx context.Context) error {
			return screenshot(ctx, 90, &detailsBuf)
		}),
		// Send
		chromedp.Submit("visitors"),
	)
	if err != nil {
		log.Println(err)
		return
	}
	detailsScreenshotFile := fmt.Sprintf("%s-%s-%s-registration.png", visitor.FirstName, visitor.LastName, time.Now().Format("20060102"))
	if err := ioutil.WriteFile(detailsScreenshotFile, detailsBuf, 0644); err != nil {
		log.Fatal(err)
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

func validateResident(resident resident) error {
	if len(resident.FirstName) == 0 {
		return errors.New("resident first name is required")
	} else if len(resident.LastName) == 0 {
		return errors.New("resident last name is required")
	} else if len(resident.PropertyName) == 0 {
		return errors.New("resident property name is required")
	} else if len(resident.StreetAddress) == 0 {
		return errors.New("resident street address is required")
	} else if len(resident.ApartmentNumber) == 0 {
		return errors.New("resident apartment number is required")
	} else if len(resident.City) == 0 {
		return errors.New("resident city is required")
	} else if len(resident.State) == 0 {
		return errors.New("resident state is required")
	} else if len(resident.ZipCode) == 0 {
		return errors.New("resident zipcode is required")
	}
	return nil
}

func validateVisitor(visitor visitor) error {
	if len(visitor.FirstName) == 0 {
		return errors.New("visitor first name is required")
	} else if len(visitor.LastName) == 0 {
		return errors.New("visitor last name is required")
	} else if len(visitor.StreetAddress) == 0 {
		return errors.New("visitor street address is required")
	} else if len(visitor.ApartmentNumber) == 0 {
		return errors.New("visitor apartment number is required")
	} else if len(visitor.City) == 0 {
		return errors.New("visitor city is required")
	} else if len(visitor.ZipCode) == 0 {
		return errors.New("visitor zipcode is required")
	} else if len(visitor.EmailAddress) == 0 {
		return errors.New("visitor email address is required")
	} else if len(visitor.PhoneNumber) == 0 {
		return errors.New("visitor phone number is required")
	} else if len(visitor.Vehicle.Year) == 0 {
		return errors.New("visitor vehicle year is required")
	} else if len(visitor.Vehicle.Make) == 0 {
		return errors.New("visitor vehicle make is required")
	} else if len(visitor.Vehicle.Model) == 0 {
		return errors.New("visitor vehicle model is required")
	} else if len(visitor.Vehicle.Color) == 0 {
		return errors.New("visitor vehicle color is required")
	} else if len(visitor.Vehicle.LicencePlateNumber) == 0 {
		return errors.New("visitor vehicle license plate number is required")
	} else if len(visitor.Vehicle.LicensePlateStateIssuer) == 0 {
		return errors.New("visitor vehicle license plate state issuer is required")
	}
	return nil
}

func createSelector(propertyName string) string {
	return fmt.Sprintf(`//input[@name="%s"]`, propertyName)
}

func screenshot(ctx context.Context, quality int64, res *[]byte) error {
	// get layout metrics
	_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
	if err != nil {
		return err
	}
	width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))
	// force viewport emulation
	err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
		WithScreenOrientation(&emulation.ScreenOrientation{
			Type:  emulation.OrientationTypePortraitPrimary,
			Angle: 0,
		}).Do(ctx)
	if err != nil {
		return err
	}

	// capture screenshot
	*res, err = page.CaptureScreenshot().
		WithQuality(quality).
		WithClip(&page.Viewport{
			X:      contentSize.X,
			Y:      contentSize.Y,
			Width:  contentSize.Width,
			Height: contentSize.Height,
			Scale:  1,
		}).Do(ctx)
	if err != nil {
		return err
	}
	return nil
}

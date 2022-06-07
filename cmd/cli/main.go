package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/darrenparkinson/ccw"
)

func main() {

	var username, password, clientID, clientSecret string
	mustMapEnv(&username, "CCW_USERNAME")
	mustMapEnv(&password, "CCW_PASSWORD")
	mustMapEnv(&clientID, "CCW_CLIENTID")
	mustMapEnv(&clientSecret, "CCW_CLIENTSECRET")
	c, err := ccw.NewClient(username, password, clientID, clientSecret, nil)
	if err != nil {
		log.Fatal(err)
	}

	qr, err := c.QuoteService.AcquireByDealID(context.Background(), "123456")
	if err != nil {
		log.Fatal(err)
	}
	data := [][]string{}
	data = append(data, []string{"Part Number", "List Price", "Discount", "Buy Price", "Import Currency", "Quantity", "Duration"})

	for _, item := range qr.LineItems {
		data = append(data, []string{item.PartNumber, fmt.Sprintf("%.2f", item.UnitPrice), fmt.Sprintf("%.2f", item.EffectiveDiscount), fmt.Sprintf("%.2f", item.UnitNetPrice), item.ImportCurrency, fmt.Sprintf("%d", item.Quantity), fmt.Sprintf("%.2f", item.ServiceDurationMonths)})
	}

	csvExport("export.csv", data)
	// err = c.EstimateService.List(context.Background())
	// if err != nil {
	// 	log.Fatal(err)
	// }

}

func mustMapEnv(target *string, envKey string) {
	v := os.Getenv(envKey)
	if v == "" {
		log.Fatalf("environment variable %q not set", envKey)
	}
	*target = v
}

func csvExport(filename string, data [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range data {
		if err := writer.Write(value); err != nil {
			return err
		}
	}
	return nil
}

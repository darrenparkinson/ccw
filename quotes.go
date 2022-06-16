package ccw

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

type ListQuoteRequest struct {
	DealID string
}

type AcquireQuoteRequest struct {
	DealID string
}

type AcquireQuoteResponse struct {
	QuoteName   string                     `json:"quoteName"`
	QuoteOwner  string                     `json:"quoteOwner"`
	QuoteStatus string                     `json:"quoteStatus"`
	PriceList   string                     `json:"priceList"`
	PriceListID string                     `json:"priceListId"`
	DealID      string                     `json:"dealId"`
	Customer    Company                    `json:"customer"`
	Partner     Company                    `json:"partner"`
	LineItems   []AcquireQuoteResponseItem `json:"items"`
}

type Company struct {
	Name     string  `json:"name"`
	Location Address `json:"location"`
	Contact  Contact `json:"contact"`
}

type Address struct {
	LineOne                string `json:"lineOne"`
	LineTwo                string `json:"lineTwo"`
	LineThree              string `json:"lineThree"`
	CityName               string `json:"cityName"`
	CountrySubDivisionCode string `json:"countrySubDivisionCode"`
	CountryCode            string `json:"countryCode"`
	PostalCode             string `json:"postalCode"`
}

type Contact struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	JobTitle  string `json:"jobTitle"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
	Website   string `json:"website"`
}

type AcquireQuoteResponseItem struct {
	LineNumber                string      `json:"lineNumber"`
	PartNumber                string      `json:"partNumber"`
	Description               string      `json:"description"`
	CCWLineNumber             string      `json:"ccwLineNumber"`
	UnitNetPriceBeforeCredits float64     `json:"unitNetPriceBeforeCredits"`
	UnitNetPrice              float64     `json:"unitNetPrice"`
	OriginalUnitListPrice     float64     `json:"originalUnitListPrice"`
	Quantity                  int64       `json:"quantity"`
	UnitPrice                 float64     `json:"unitPrice"`
	ExtendedAmount            float64     `json:"extendedAmount"`
	TotalAmount               float64     `json:"totalAmount"`
	TotalDiscount             float64     `json:"totalDiscount"`
	StandardDiscount          float64     `json:"standardDiscount"`
	PromotionalDiscount       float64     `json:"promotionalDiscount"`
	ContractualDiscount       float64     `json:"contractualDiscount"`
	NonStandardDiscount       float64     `json:"nonStandardDiscount"`
	PrePayDiscount            float64     `json:"prePayDiscount"`
	EffectiveDiscount         float64     `json:"effectiveDiscount"`
	ImportCurrency            string      `json:"importCurrency"`
	ISO8601ServiceDuration    string      `json:"iso8601ServiceDuration"`
	ServiceDurationMonths     floatOrNull `json:"serviceDurationMonths"`
	ISO8601LeadTime           string      `json:"iso8601LeadTime"`
	LeadTimeDays              floatOrNull `json:"leadTimeDays"`
	ProductTypeClassification string      `json:"productTypeClassification"`
	ServiceLevelName          string      `json:"serviceLevelName"`
	ServiceType               string      `json:"serviceType"`
	ParentLineNumber          *string     `json:"parentLineNumber,omitempty"`
	// UserArea fields
	MagicKey           *string `json:"magicKey,omitempty"`
	RequestedStartDate *string `json:"requestedStartDate,omitempty"`
	InitialTerm        *int64  `json:"initialTerm,omitempty"`
	AutoRenewalTerm    *int64  `json:"autoRenewalTerm,omitempty"`
	BillingModel       *string `json:"billingModel,omitempty"`
	ChargeType         *string `json:"chargeType,omitempty"`
	UnitOfMeasurement  *string `json:"unitOfMeasurement,omitempty"`
	AdditionalItemInfo *string `json:"additionalItemInfo,omitempty"`
	PricingTerm        *int64  `json:"pricingTerm,omitempty"`

	RemainingTerm           *float64 `json:"remainingTerm,omitempty"`
	SubscriptionReferenceID *string  `json:"subscriptionReferenceID,omitempty"`
}

type floatOrNull float64

func (f floatOrNull) MarshalJSON() ([]byte, error) {
	if math.IsInf(float64(f), 0) || math.IsNaN(float64(f)) {
		return nil, &json.UnsupportedValueError{
			Value: reflect.ValueOf(f),
			Str:   strconv.FormatFloat(float64(f), 'f', 2, 64),
		}
	}
	if float64(f) == 0 {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatFloat(float64(f), 'f', -1, 64)), nil
}

func (s *QuoteService) AcquireByDealID(ctx context.Context, dealID string) (*AcquireQuoteResponse, error) {
	// 1. Load the template
	template, err := template.ParseFS(templates, "templates/AcquireQuote_Request.xml")
	if err != nil {
		return nil, err
	}
	// 2. Create the data for the template
	data := AcquireQuoteRequest{DealID: dealID}

	// 3. Apply the data to the template
	var tpl bytes.Buffer
	if err := template.Execute(&tpl, data); err != nil {
		return nil, err
	}

	// 4. Make the request
	qurl := fmt.Sprintf("%s/AcquireQuoteService", s.BaseURL)
	req, err := http.NewRequest("POST", qurl, strings.NewReader(tpl.String()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/xml")
	req.Header.Add("Content-Type", "application/xml")

	var resp AcquireQuoteXMLResponse
	err = s.client.makeXMLRequest(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	quoteSuccess := resp.Body.ShowQuote.DataArea.Show.ResponseCriteria.ChangeStatus.Reason
	if quoteSuccess != "Success" {
		return nil, fmt.Errorf("%s: %s", resp.Body.ShowQuote.DataArea.Show.ResponseCriteria.ChangeStatus.Reason, resp.Body.ShowQuote.DataArea.Show.ResponseCriteria.ChangeStatus.Text)
	}

	// 5. Format the response
	var aqr AcquireQuoteResponse

	// get header details
	quoteHeader := resp.Body.ShowQuote.DataArea.Quote.QuoteHeader

	if quoteHeader.Extension.ValueText.TypeCode == "QuoteName" {
		aqr.QuoteName = quoteHeader.Extension.ValueText.Text
	}

	aqr.QuoteStatus = quoteHeader.Status.Code.Text
	for _, party := range quoteHeader.Party {
		switch party.Role {
		case "QuoteOwner":
			aqr.QuoteOwner = party.Contact.ID.Text
		case "End Customer":
			aqr.Customer = Company{
				Name: party.Name,
				Location: Address{
					LineOne:                party.Location.Address.LineOne,
					LineTwo:                party.Location.Address.LineTwo,
					LineThree:              party.Location.Address.LineThree,
					CityName:               party.Location.Address.CityName,
					CountrySubDivisionCode: party.Location.Address.CountrySubDivisionCode,
					CountryCode:            party.Location.Address.CountryCode,
					PostalCode:             party.Location.Address.PostalCode,
				},
				Contact: Contact{
					JobTitle:  party.Contact.JobTitle,
					Telephone: party.Contact.TelephoneCommunication.FormattedNumber,
					Email:     party.Contact.EMailAddressCommunication.EMailAddressID,
				},
			}
			for _, field := range party.Contact.Name {
				if field.SequenceName == "First Name" {
					aqr.Customer.Contact.FirstName = field.Text
				}
				if field.SequenceName == "Last Name" {
					aqr.Customer.Contact.LastName = field.Text
				}
			}
			if party.Contact.ID.SchemeName == "Website" {
				aqr.Customer.Contact.Website = party.Contact.ID.Text
			}
		case "Partner":
			aqr.Partner = Company{
				Name: party.PartyIDs.ID,
				Location: Address{
					LineOne:                party.Location.Address.LineOne,
					LineTwo:                party.Location.Address.LineTwo,
					LineThree:              party.Location.Address.LineThree,
					CityName:               party.Location.Address.CityName,
					CountrySubDivisionCode: party.Location.Address.CountrySubDivisionCode,
					CountryCode:            party.Location.Address.CountryCode,
					PostalCode:             party.Location.Address.PostalCode,
				},
				Contact: Contact{
					JobTitle:  party.Contact.JobTitle,
					Telephone: party.Contact.TelephoneCommunication.FormattedNumber,
					Email:     party.Contact.EMailAddressCommunication.EMailAddressID,
				},
			}
			for _, field := range party.Contact.Name {
				if field.SequenceName == "First Name" {
					aqr.Partner.Contact.FirstName = field.Text
				}
				if field.SequenceName == "Last Name" {
					aqr.Partner.Contact.LastName = field.Text
				}
			}
			if party.Contact.ID.SchemeName == "Website" {
				aqr.Partner.Contact.Website = party.Contact.ID.Text
			}
		}
	}

	if quoteHeader.QualificationTerm.TypeAttribute == "Deal" && quoteHeader.QualificationTerm.ID.SchemeAgencyName == "Cisco" {
		aqr.DealID = quoteHeader.QualificationTerm.ID.Text
	}

	for _, p := range quoteHeader.UserArea.CiscoExtensions.CiscoHeader.PriceList {
		if p.Description != "" && p.ID != "" {
			aqr.PriceList = p.Description
			aqr.PriceListID = p.ID
		}
	}

	// Now get the quote lines

	quoteLines := resp.Body.ShowQuote.DataArea.Quote.QuoteLine

	for _, line := range quoteLines {
		ql := AcquireQuoteResponseItem{}
		item := line.Item
		ql.LineNumber = line.LineNumber
		ql.PartNumber = line.Item.ItemID.ID.Text
		for _, d := range item.Description {
			if d.Text != "" && d.Type != "ServiceType" && d.Type != "ServiceLevelName" {
				ql.Description = d.Text
			}
			if d.Type == "ServiceType" {
				ql.ServiceType = d.Text
			}
			if d.Type == "ServiceLevelName" {
				ql.ServiceLevelName = d.Text
			}
		}
		if line.Item.Classification.Type.ListName == "ProductType" {
			ql.ProductTypeClassification = line.Item.Classification.Type.Text
		}
		for _, prop := range item.Specification.Property {

			if prop.ParentID != "" && prop.ParentID != "0" {
				ql.ParentLineNumber = String(prop.ParentID)
			}

			switch prop.NameValue.Name {
			case "CCWLineNumber":
				ql.CCWLineNumber = prop.NameValue.Text
			case "UnitNetPrice":
				v, _ := strconv.ParseFloat(prop.NameValue.Text, 64)
				ql.UnitNetPrice = v
			case "UnitNetPriceBeforeCredits":
				v, _ := strconv.ParseFloat(prop.NameValue.Text, 64)
				ql.UnitNetPriceBeforeCredits = v
			case "OriginalUnitListPrice":
				v, _ := strconv.ParseFloat(prop.NameValue.Text, 64)
				ql.OriginalUnitListPrice = v
			case "BundleIndicator":
				for _, eff := range prop.Effectivity {
					if eff.Type == "ServiceDuration" {
						ql.ISO8601ServiceDuration = eff.EffectiveTimePeriod.Duration
						duration, _ := isoDurationToMonthsFloat(eff.EffectiveTimePeriod.Duration)
						ql.ServiceDurationMonths = floatOrNull(duration)
					}
					if eff.Type == "LeadTime" {
						ql.ISO8601LeadTime = eff.EffectiveTimePeriod.Duration
						duration, _ := isoDurationToDaysFloat(eff.EffectiveTimePeriod.Duration)
						ql.LeadTimeDays = floatOrNull(duration)
					}
				}
			}
		}

		for _, discount := range line.PaymentTerm.Discount {
			switch discount.Type.Text {
			case "TotalDiscount":
				v, _ := strconv.ParseFloat(discount.DiscountPercent, 64)
				ql.TotalDiscount = v
			case "StandardDiscount":
				v, _ := strconv.ParseFloat(discount.DiscountPercent, 64)
				ql.StandardDiscount = v
			case "PromotionalDiscount":
				v, _ := strconv.ParseFloat(discount.DiscountPercent, 64)
				ql.PromotionalDiscount = v
			case "ContractualDiscount":
				v, _ := strconv.ParseFloat(discount.DiscountPercent, 64)
				ql.ContractualDiscount = v
			case "NonStandardDiscount":
				v, _ := strconv.ParseFloat(discount.DiscountPercent, 64)
				ql.NonStandardDiscount = v
			case "PrePay":
				v, _ := strconv.ParseFloat(discount.DiscountPercent, 64)
				ql.PrePayDiscount = v
			case "EffectiveDiscount":
				v, _ := strconv.ParseFloat(discount.DiscountPercent, 64)
				ql.EffectiveDiscount = v
			}
		}
		ql.ImportCurrency = line.UnitPrice.Amount.CurrencyID
		vUnitPrice, _ := strconv.ParseFloat(line.UnitPrice.Amount.Text, 64)
		ql.UnitPrice = vUnitPrice
		vQuantity, _ := strconv.ParseInt(line.Quantity, 10, 0)
		ql.Quantity = vQuantity

		// Additional UserArea fields
		ciscoLine := line.UserArea.CiscoExtensions.CiscoLine
		ql.SubscriptionReferenceID = String(ciscoLine.SubscriptionReferenceID)
		ql.MagicKey = String(ciscoLine.MagicKey)
		ql.RequestedStartDate = String(ciscoLine.RequestedStartDate)
		initialTerm, _ := strconv.ParseInt(ciscoLine.InitialTerm, 10, 0)
		ql.InitialTerm = IntOrNil(initialTerm)
		renewalTerm, _ := strconv.ParseInt(ciscoLine.AutoRenewalTerm, 10, 0)
		ql.AutoRenewalTerm = IntOrNil(renewalTerm)
		ql.BillingModel = String(ciscoLine.BillingModel)
		ql.ChargeType = String(ciscoLine.ChargeType)
		ql.UnitOfMeasurement = String(ciscoLine.UnitOfMeasurement)
		ql.AdditionalItemInfo = String(ciscoLine.AdditionalItemInfo)
		pricingTerm, _ := strconv.ParseInt(ciscoLine.PricingTerm, 10, 0)
		ql.PricingTerm = IntOrNil(pricingTerm)
		remainingTerm, _ := strconv.ParseFloat(ciscoLine.RemainingTerm, 64)
		ql.RemainingTerm = FloatOrNil(remainingTerm)

		aqr.LineItems = append(aqr.LineItems, ql)
	}

	return &aqr, nil
}

type AcquireQuoteXMLResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	Soapenv string   `xml:"soapenv,attr"`
	Header  struct {
		Text      string `xml:",chardata"`
		Messaging struct {
			Text           string `xml:",chardata"`
			Xsi            string `xml:"xsi,attr"`
			Eb             string `xml:"eb,attr"`
			MustUnderstand string `xml:"mustUnderstand,attr"`
			SchemaLocation string `xml:"schemaLocation,attr"`
			UserMessage    struct {
				Text        string `xml:",chardata"`
				MessageInfo struct {
					Text           string `xml:",chardata"`
					Timestamp      string `xml:"Timestamp"`
					MessageId      string `xml:"MessageId"`
					RefToMessageId string `xml:"RefToMessageId"`
				} `xml:"MessageInfo"`
				PartyInfo struct {
					Text string `xml:",chardata"`
					From struct {
						Text    string `xml:",chardata"`
						PartyId string `xml:"PartyId"`
						Role    string `xml:"Role"`
					} `xml:"From"`
					To struct {
						Text    string `xml:",chardata"`
						PartyId string `xml:"PartyId"`
						Role    string `xml:"Role"`
					} `xml:"To"`
				} `xml:"PartyInfo"`
				CollaborationInfo struct {
					Text    string `xml:",chardata"`
					Service struct {
						Text string `xml:",chardata"`
						Type string `xml:"type,attr"`
					} `xml:"Service"`
					Action         string `xml:"Action"`
					ConversationId string `xml:"ConversationId"`
				} `xml:"CollaborationInfo"`
				PayloadInfo struct {
					Text     string `xml:",chardata"`
					PartInfo struct {
						Text   string `xml:",chardata"`
						Schema struct {
							Text     string `xml:",chardata"`
							Location string `xml:"location,attr"`
						} `xml:"Schema"`
						PartProperties struct {
							Text     string `xml:",chardata"`
							Property []struct {
								Text string `xml:",chardata"`
								Name string `xml:"name,attr"`
							} `xml:"Property"`
						} `xml:"PartProperties"`
					} `xml:"PartInfo"`
				} `xml:"PayloadInfo"`
			} `xml:"UserMessage"`
		} `xml:"Messaging"`
		MeshaDoneBlock struct {
			Text      string `xml:",chardata"`
			Mesha     string `xml:"mesha,attr"`
			MeshaDone string `xml:"MeshaDone"`
		} `xml:"MeshaDoneBlock"`
	} `xml:"Header"`
	Body struct {
		Text      string `xml:",chardata"`
		ShowQuote struct {
			Text            string `xml:",chardata"`
			Ns1             string `xml:"ns1,attr"`
			Oa              string `xml:"oa,attr"`
			Xmlns           string `xml:"xmlns,attr"`
			Xs              string `xml:"xs,attr"`
			Cisco           string `xml:"cisco,attr"`
			ApplicationArea struct {
				Text   string `xml:",chardata"`
				Sender struct {
					Text      string `xml:",chardata"`
					LogicalID struct {
						Text             string `xml:",chardata"`
						SchemeAgencyName string `xml:"schemeAgencyName,attr"`
					} `xml:"LogicalID"`
					ReferenceID string `xml:"ReferenceID"`
				} `xml:"Sender"`
				Receiver struct {
					Text      string `xml:",chardata"`
					LogicalID struct {
						Text             string `xml:",chardata"`
						SchemeAgencyName string `xml:"schemeAgencyName,attr"`
					} `xml:"LogicalID"`
				} `xml:"Receiver"`
				CreationDateTime string `xml:"CreationDateTime"`
				BODID            string `xml:"BODID"`
			} `xml:"ApplicationArea"`
			DataArea struct {
				Text string `xml:",chardata"`
				Show struct {
					Text             string `xml:",chardata"`
					ResponseCriteria struct {
						Text         string `xml:",chardata"`
						ChangeStatus struct {
							Text   string `xml:",chardata"`
							Reason string `xml:"Reason"`
						} `xml:"ChangeStatus"`
					} `xml:"ResponseCriteria"`
				} `xml:"Show"`
				Quote struct {
					Text        string `xml:",chardata"`
					QuoteHeader struct {
						Text       string `xml:",chardata"`
						DocumentID struct {
							Text string `xml:",chardata"`
							ID   string `xml:"ID"`
						} `xml:"DocumentID"`
						LastModificationDateTime string `xml:"LastModificationDateTime"`
						DocumentDateTime         string `xml:"DocumentDateTime"`
						Description              []struct {
							Text string `xml:",chardata"`
							Type string `xml:"type,attr"`
						} `xml:"Description"`
						Status struct {
							Text string `xml:",chardata"`
							Code struct {
								Text           string `xml:",chardata"`
								ListName       string `xml:"listName,attr"`
								ListAgencyName string `xml:"listAgencyName,attr"`
							} `xml:"Code"`
						} `xml:"Status"`
						Party []struct {
							Text    string `xml:",chardata"`
							Role    string `xml:"role,attr"`
							Contact struct {
								Text string `xml:",chardata"`
								ID   struct {
									Text       string `xml:",chardata"`
									SchemeName string `xml:"schemeName,attr"`
								} `xml:"ID"`
								Name []struct {
									Text         string `xml:",chardata"`
									SequenceName string `xml:"sequenceName,attr"`
								} `xml:"Name"`
								JobTitle               string `xml:"JobTitle"`
								TelephoneCommunication struct {
									Text            string `xml:",chardata"`
									FormattedNumber string `xml:"FormattedNumber"`
								} `xml:"TelephoneCommunication"`
								EMailAddressCommunication struct {
									Text           string `xml:",chardata"`
									EMailAddressID string `xml:"EMailAddressID"`
								} `xml:"EMailAddressCommunication"`
							} `xml:"Contact"`
							Name     string `xml:"Name"`
							Location struct {
								Text    string `xml:",chardata"`
								Address struct {
									Text                   string `xml:",chardata"`
									LineOne                string `xml:"LineOne"`
									LineTwo                string `xml:"LineTwo"`
									LineThree              string `xml:"LineThree"`
									CityName               string `xml:"CityName"`
									CountrySubDivisionCode string `xml:"CountrySubDivisionCode"`
									CountryCode            string `xml:"CountryCode"`
									PostalCode             string `xml:"PostalCode"`
								} `xml:"Address"`
							} `xml:"Location"`
							PartyIDs struct {
								Text string `xml:",chardata"`
								ID   string `xml:"ID"`
							} `xml:"PartyIDs"`
						} `xml:"Party"`
						BillToParty struct {
							Text     string `xml:",chardata"`
							Location struct {
								Text    string `xml:",chardata"`
								Address struct {
									Text string `xml:",chardata"`
									ID   string `xml:"ID"`
								} `xml:"Address"`
								Description struct {
									Text string `xml:",chardata"`
									Type string `xml:"type,attr"`
								} `xml:"Description"`
							} `xml:"Location"`
						} `xml:"BillToParty"`
						EffectiveTimePeriod struct {
							Text        string `xml:",chardata"`
							EndDateTime string `xml:"EndDateTime"`
						} `xml:"EffectiveTimePeriod"`
						QualificationTerm struct {
							Text          string `xml:",chardata"`
							TypeAttribute string `xml:"typeAttribute,attr"`
							ID            struct {
								Text             string `xml:",chardata"`
								SchemeAgencyName string `xml:"schemeAgencyName,attr"`
							} `xml:"ID"`
						} `xml:"QualificationTerm"`
						UserArea struct {
							Text            string `xml:",chardata"`
							CiscoExtensions struct {
								Text        string `xml:",chardata"`
								CiscoHeader struct {
									Text            string `xml:",chardata"`
									IntendedUseCode string `xml:"IntendedUseCode"`
									PriceList       []struct {
										Text        string `xml:",chardata"`
										ID          string `xml:"ID"`
										Description string `xml:"Description"`
										ShortName   string `xml:"ShortName"`
									} `xml:"PriceList"`
									ConfigurationMessages struct {
										Text        string `xml:",chardata"`
										ID          string `xml:"ID"`
										Description string `xml:"Description"`
										Reason      string `xml:"Reason"`
									} `xml:"ConfigurationMessages"`
									NetPriceProtectionDate string `xml:"NetPriceProtectionDate"`
								} `xml:"CiscoHeader"`
							} `xml:"CiscoExtensions"`
						} `xml:"UserArea"`
						Extension struct {
							Text      string `xml:",chardata"`
							ValueText struct {
								Text     string `xml:",chardata"`
								TypeCode string `xml:"typeCode,attr"`
							} `xml:"ValueText"`
						} `xml:"Extension"`
					} `xml:"QuoteHeader"`
					QuoteLine []struct {
						Text       string `xml:",chardata"`
						LineNumber string `xml:"LineNumber"`
						Item       struct {
							Text   string `xml:",chardata"`
							ItemID struct {
								Text string `xml:",chardata"`
								ID   struct {
									Text             string `xml:",chardata"`
									SchemeName       string `xml:"schemeName,attr"`
									SchemeAgencyName string `xml:"schemeAgencyName,attr"`
								} `xml:"ID"`
							} `xml:"ItemID"`
							Description []struct {
								Text string `xml:",chardata"`
								Type string `xml:"type,attr"`
							} `xml:"Description"`
							Classification struct {
								Text  string `xml:",chardata"`
								Codes struct {
									Text string `xml:",chardata"`
									Code struct {
										Text           string `xml:",chardata"`
										ListAgencyName string `xml:"listAgencyName,attr"`
									} `xml:"Code"`
								} `xml:"Codes"`
								Type struct {
									Text           string `xml:",chardata"`
									ListName       string `xml:"listName,attr"`
									ListAgencyName string `xml:"listAgencyName,attr"`
								} `xml:"Type"`
							} `xml:"Classification"`
							Specification struct {
								Text     string `xml:",chardata"`
								Property []struct {
									Text      string `xml:",chardata"`
									ParentID  string `xml:"ParentID"`
									NameValue struct {
										Text string `xml:",chardata"`
										Name string `xml:"name,attr"`
									} `xml:"NameValue"`
									Description []struct {
										Text string `xml:",chardata"`
										Type string `xml:"type,attr"`
									} `xml:"Description"`
									Effectivity []struct {
										Text                string `xml:",chardata"`
										Type                string `xml:"Type"`
										EffectiveTimePeriod struct {
											Text     string `xml:",chardata"`
											Duration string `xml:"Duration"`
										} `xml:"EffectiveTimePeriod"`
									} `xml:"Effectivity"`
								} `xml:"Property"`
							} `xml:"Specification"`
						} `xml:"Item"`
						Quantity  string `xml:"Quantity"`
						UnitPrice struct {
							Text   string `xml:",chardata"`
							Amount struct {
								Text       string `xml:",chardata"`
								CurrencyID string `xml:"currencyID,attr"`
							} `xml:"Amount"`
						} `xml:"UnitPrice"`
						ExtendedAmount struct {
							Text       string `xml:",chardata"`
							CurrencyID string `xml:"currencyID,attr"`
						} `xml:"ExtendedAmount"`
						TotalAmount struct {
							Text       string `xml:",chardata"`
							CurrencyID string `xml:"currencyID,attr"`
						} `xml:"TotalAmount"`
						PaymentTerm struct {
							Text     string `xml:",chardata"`
							Discount []struct {
								Text string `xml:",chardata"`
								Type struct {
									Text     string `xml:",chardata"`
									ListName string `xml:"listName,attr"`
								} `xml:"Type"`
								DiscountPercent string `xml:"DiscountPercent"`
							} `xml:"Discount"`
						} `xml:"PaymentTerm"`
						Allowance []struct {
							Text string `xml:",chardata"`
							Type struct {
								Text     string `xml:",chardata"`
								ListName string `xml:"listName,attr"`
							} `xml:"Type"`
							Amount struct {
								Text       string `xml:",chardata"`
								CurrencyID string `xml:"currencyID,attr"`
							} `xml:"Amount"`
						} `xml:"Allowance"`
						UserArea struct {
							Text            string `xml:",chardata"`
							CiscoExtensions struct {
								Text      string `xml:",chardata"`
								CiscoLine struct {
									Text      string `xml:",chardata"`
									BuyMethod string `xml:"BuyMethod"`
									Party     []struct {
										Text     string `xml:",chardata"`
										Category string `xml:"category,attr"`
										Location struct {
											Text    string `xml:",chardata"`
											Address struct {
												Text                   string `xml:",chardata"`
												LineOne                string `xml:"LineOne"`
												LineTwo                string `xml:"LineTwo"`
												LineThree              string `xml:"LineThree"`
												CityName               string `xml:"CityName"`
												CountrySubDivisionCode string `xml:"CountrySubDivisionCode"`
												CountryCode            string `xml:"CountryCode"`
												PostalCode             string `xml:"PostalCode"`
											} `xml:"Address"`
										} `xml:"Location"`
									} `xml:"Party"`
									ConfigurationReference []struct {
										Text   string `xml:",chardata"`
										Status struct {
											Text   string `xml:",chardata"`
											Reason string `xml:"Reason"`
										} `xml:"Status"`
										VerifiedConfigurationIndicator string `xml:"VerifiedConfigurationIndicator"`
									} `xml:"ConfigurationReference"`
									ConfiguratorInformation struct {
										Text                          string `xml:",chardata"`
										ConfigurationPath             string `xml:"ConfigurationPath"`
										ProductConfigurationReference string `xml:"ProductConfigurationReference"`
										ConfigurationSelectCode       string `xml:"ConfigurationSelectCode"`
									} `xml:"ConfiguratorInformation"`
									MagicKey                string `xml:"MagicKey"`
									RequestedStartDate      string `xml:"RequestedStartDate"`
									InitialTerm             string `xml:"InitialTerm"`
									AutoRenewalTerm         string `xml:"AutoRenewalTerm"`
									BillingModel            string `xml:"BillingModel"`
									AdditionalItemInfo      string `xml:"AdditionalItemInfo"`
									ListPriceVersion        string `xml:"ListPriceVersion"`
									UtilityDrawdownAmount   string `xml:"UtilityDrawdownAmount"`
									TransactionInfoID       string `xml:"TransactionInfoID"`
									ChargeType              string `xml:"ChargeType"`
									UnitOfMeasurement       string `xml:"UnitOfMeasurement"`
									UsageQuantity           string `xml:"UsageQuantity"`
									PricingTerm             string `xml:"PricingTerm"`
									CIExtNetPrice           string `xml:"CIExtNetPrice"`
									SubscriptionReferenceID string `xml:"SubscriptionReferenceID"`
									CurrentBillingAmount    string `xml:"CurrentBillingAmount"`
									NewBillingAmount        string `xml:"NewBillingAmount"`
									BillingAmountNetChange  string `xml:"BillingAmountNetChange"`
									UnitNetPriceWithCredits string `xml:"UnitNetPriceWithCredits"`
									CurrentContractAmount   string `xml:"CurrentContractAmount"`
									NewContractAmount       string `xml:"NewContractAmount"`
									ContractAmountNetChange string `xml:"ContractAmountNetChange"`
									RemainingTerm           string `xml:"RemainingTerm"`
									OldQuantity             string `xml:"OldQuantity"`
									ShipToParty             struct {
										Text     string `xml:",chardata"`
										Location struct {
											Text    string `xml:",chardata"`
											Address struct {
												Text        string `xml:",chardata"`
												CountryCode string `xml:"CountryCode"`
											} `xml:"Address"`
										} `xml:"Location"`
									} `xml:"ShipToParty"`
									NetPriceProtectionDate string `xml:"NetPriceProtectionDate"`
								} `xml:"CiscoLine"`
							} `xml:"CiscoExtensions"`
						} `xml:"UserArea"`
					} `xml:"QuoteLine"`
				} `xml:"Quote"`
			} `xml:"DataArea"`
		} `xml:"ShowQuote"`
	} `xml:"Body"`
}

type ListQuoteXMLResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	Soapenv string   `xml:"soapenv,attr"`
	Header  struct {
		Text      string `xml:",chardata"`
		Messaging struct {
			Text           string `xml:",chardata"`
			Xsi            string `xml:"xsi,attr"`
			Eb             string `xml:"eb,attr"`
			MustUnderstand string `xml:"mustUnderstand,attr"`
			SchemaLocation string `xml:"schemaLocation,attr"`
			UserMessage    struct {
				Text        string `xml:",chardata"`
				MessageInfo struct {
					Text           string `xml:",chardata"`
					Timestamp      string `xml:"Timestamp"`
					MessageId      string `xml:"MessageId"`
					RefToMessageId string `xml:"RefToMessageId"`
				} `xml:"MessageInfo"`
				PartyInfo struct {
					Text string `xml:",chardata"`
					From struct {
						Text    string `xml:",chardata"`
						PartyId string `xml:"PartyId"`
						Role    string `xml:"Role"`
					} `xml:"From"`
					To struct {
						Text    string `xml:",chardata"`
						PartyId string `xml:"PartyId"`
						Role    string `xml:"Role"`
					} `xml:"To"`
				} `xml:"PartyInfo"`
				CollaborationInfo struct {
					Text    string `xml:",chardata"`
					Service struct {
						Text string `xml:",chardata"`
						Type string `xml:"type,attr"`
					} `xml:"Service"`
					Action         string `xml:"Action"`
					ConversationId string `xml:"ConversationId"`
				} `xml:"CollaborationInfo"`
				PayloadInfo struct {
					Text     string `xml:",chardata"`
					PartInfo struct {
						Text   string `xml:",chardata"`
						Schema struct {
							Text     string `xml:",chardata"`
							Location string `xml:"location,attr"`
						} `xml:"Schema"`
						PartProperties struct {
							Text     string `xml:",chardata"`
							Property []struct {
								Text string `xml:",chardata"`
								Name string `xml:"name,attr"`
							} `xml:"Property"`
						} `xml:"PartProperties"`
					} `xml:"PartInfo"`
				} `xml:"PayloadInfo"`
			} `xml:"UserMessage"`
		} `xml:"Messaging"`
		MeshaDoneBlock struct {
			Text      string `xml:",chardata"`
			Mesha     string `xml:"mesha,attr"`
			MeshaDone string `xml:"MeshaDone"`
		} `xml:"MeshaDoneBlock"`
	} `xml:"Header"`
	Body struct {
		Text      string `xml:",chardata"`
		ShowQuote struct {
			Text            string `xml:",chardata"`
			Ns1             string `xml:"ns1,attr"`
			Oa              string `xml:"oa,attr"`
			Xs              string `xml:"xs,attr"`
			Cisco           string `xml:"cisco,attr"`
			P               string `xml:"p,attr"`
			B2b             string `xml:"b2b,attr"`
			ReleaseID       string `xml:"releaseID,attr"`
			ApplicationArea struct {
				Text             string `xml:",chardata"`
				CreationDateTime string `xml:"CreationDateTime"`
			} `xml:"ApplicationArea"`
			DataArea struct {
				Text  string `xml:",chardata"`
				Show  string `xml:"Show"`
				Quote []struct {
					Text        string `xml:",chardata"`
					QuoteHeader struct {
						Text       string `xml:",chardata"`
						DocumentID struct {
							Text string `xml:",chardata"`
							ID   string `xml:"ID"`
						} `xml:"DocumentID"`
						LastModificationDateTime string `xml:"LastModificationDateTime"`
						DocumentDateTime         string `xml:"DocumentDateTime"`
						Description              struct {
							Text string `xml:",chardata"`
							Type string `xml:"type,attr"`
						} `xml:"Description"`
						Status struct {
							Text string `xml:",chardata"`
							Code struct {
								Text           string `xml:",chardata"`
								ListName       string `xml:"listName,attr"`
								ListAgencyName string `xml:"listAgencyName,attr"`
							} `xml:"Code"`
						} `xml:"Status"`
						Party struct {
							Text     string `xml:",chardata"`
							Role     string `xml:"role,attr"`
							Name     string `xml:"Name"`
							Location struct {
								Text    string `xml:",chardata"`
								Address struct {
									Text        string `xml:",chardata"`
									AddressLine struct {
										Text     string `xml:",chardata"`
										Sequence string `xml:"sequence,attr"`
									} `xml:"AddressLine"`
									CityName               string `xml:"CityName"`
									CountrySubDivisionCode string `xml:"CountrySubDivisionCode"`
									CountryCode            string `xml:"CountryCode"`
									PostalCode             string `xml:"PostalCode"`
								} `xml:"Address"`
							} `xml:"Location"`
						} `xml:"Party"`
						QualificationTerm struct {
							Text string `xml:",chardata"`
							ID   struct {
								Text             string `xml:",chardata"`
								SchemeAgencyName string `xml:"schemeAgencyName,attr"`
							} `xml:"ID"`
						} `xml:"QualificationTerm"`
						UserArea struct {
							Text            string `xml:",chardata"`
							CiscoExtensions struct {
								Text        string `xml:",chardata"`
								CiscoHeader struct {
									Text      string `xml:",chardata"`
									PriceList struct {
										Text        string `xml:",chardata"`
										Description string `xml:"Description"`
									} `xml:"PriceList"`
									ConfigurationMessages struct {
										Text        string `xml:",chardata"`
										ID          string `xml:"ID"`
										Description string `xml:"Description"`
										Reason      string `xml:"Reason"`
									} `xml:"ConfigurationMessages"`
								} `xml:"CiscoHeader"`
							} `xml:"CiscoExtensions"`
						} `xml:"UserArea"`
						Extension struct {
							Chardata string `xml:",chardata"`
							Amount   []struct {
								Text     string `xml:",chardata"`
								TypeCode string `xml:"typeCode,attr"`
							} `xml:"Amount"`
							Text []struct {
								Text     string `xml:",chardata"`
								TypeCode string `xml:"typeCode,attr"`
							} `xml:"Text"`
						} `xml:"Extension"`
						EffectiveTimePeriod struct {
							Text        string `xml:",chardata"`
							EndDateTime string `xml:"EndDateTime"`
						} `xml:"EffectiveTimePeriod"`
					} `xml:"QuoteHeader"`
				} `xml:"Quote"`
			} `xml:"DataArea"`
		} `xml:"ShowQuote"`
	} `xml:"Body"`
}

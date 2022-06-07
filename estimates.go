package ccw

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"text/template"
)

type ListEstimateRequest struct {
	Description string
}

func (s *EstimateService) List(ctx context.Context) error {

	// 1. Load the template
	template, err := template.ParseFS(templates, "templates/ListEstimate_Request.xml")
	if err != nil {
		return err
	}
	// 2. Create the data for the template
	data := ListEstimateRequest{
		Description: "Hello World",
	}
	// 3. Apply the data to the template
	var tpl bytes.Buffer
	if err := template.Execute(&tpl, data); err != nil {
		return err
	}

	// 4. Send the data

	qurl := fmt.Sprintf("%s/listEstimate", s.BaseURL)
	fmt.Println(qurl)
	req, err := http.NewRequest("POST", qurl, strings.NewReader(tpl.String()))
	if err != nil {
		return nil
	}
	req.Header.Add("Accept", "application/xml")
	req.Header.Add("Content-Type", "application/xml")
	var resp ListEstimateXMLResponse
	err = s.client.makeXMLRequest(ctx, req, &resp)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", resp)
	fmt.Println(resp.Body.ShowQuote.DataArea.Show.ResponseCriteria.ChangeStatus.Reason)
	fmt.Println(resp.Body.ShowQuote.DataArea.Quote.QuoteHeader.Message.Description)
	fmt.Println(resp.Body.ShowQuote.DataArea.Quote.QuoteHeader.Message.Extension.ValueText.TypeCode)
	return nil
}

type ListEstimateXMLResponse struct {
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
			Xmlns           string `xml:"xmlns,attr"`
			B2bapi          string `xml:"b2bapi,attr"`
			ApplicationArea struct {
				Text   string `xml:",chardata"`
				Sender struct {
					Text        string `xml:",chardata"`
					LogicalID   string `xml:"LogicalID"`
					ReferenceID string `xml:"ReferenceID"`
				} `xml:"Sender"`
				Receiver struct {
					Text      string `xml:",chardata"`
					LogicalID string `xml:"LogicalID"`
					ID        string `xml:"ID"`
				} `xml:"Receiver"`
				CreationDateTime string `xml:"CreationDateTime"`
				BODID            struct {
					Text            string `xml:",chardata"`
					SchemeAgencyID  string `xml:"schemeAgencyID,attr"`
					SchemeVersionID string `xml:"schemeVersionID,attr"`
				} `xml:"BODID"`
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
						Text    string `xml:",chardata"`
						Message struct {
							Text        string `xml:",chardata"`
							ID          string `xml:"ID"`
							Description string `xml:"Description"`
							Extension   struct {
								Text      string `xml:",chardata"`
								ValueText struct {
									Text     string `xml:",chardata"`
									TypeCode string `xml:"typeCode,attr"`
								} `xml:"ValueText"`
							} `xml:"Extension"`
						} `xml:"Message"`
					} `xml:"QuoteHeader"`
				} `xml:"Quote"`
			} `xml:"DataArea"`
		} `xml:"ShowQuote"`
	} `xml:"Body"`
}

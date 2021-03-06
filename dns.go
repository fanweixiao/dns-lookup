package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/miekg/dns"
	"strings"
)

type Request struct {
	DomainName string `json:"name"`
	Type       string `json:"type"`
}

type Response struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type"`
	TTL   uint32 `json:"ttl"`
}

func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod != "POST" {
		return &events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type":                 "application/json",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "*",
			},
		}, nil
	}
	var req Request
	if err := json.NewDecoder(strings.NewReader(request.Body)).Decode(&req); err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Failed to parse payload: %v", err),
		}, nil
	}

	var results *[]Response
	var err error
	if results, err = query(req.Type, req.DomainName); err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Failed to make DNS request: %v", err),
		}, nil
	}

	output, err := json.Marshal(results)

	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "*",
		},
		Body:            string(output),
		IsBase64Encoded: false,
	}, nil
}

func message(typ string, name string) *dns.Msg {
	recs := map[string]uint16{
		"a":       dns.TypeA,
		"aaaa":    dns.TypeAAAA,
		"cname":   dns.TypeCNAME,
		"mx":      dns.TypeMX,
		"ns":      dns.TypeNS,
		"ptr":     dns.TypePTR,
		"soa":     dns.TypeSOA,
		"srv":     dns.TypeSRV,
		"txt":     dns.TypeTXT,
		"dnskey":  dns.TypeDNSKEY,
		"ds":      dns.TypeDS,
		"nsec":    dns.TypeNSEC,
		"nsec3":   dns.TypeNSEC3,
		"rrsig":   dns.TypeRRSIG,
		"afsdb":   dns.TypeAFSDB,
		"atma":    dns.TypeATMA,
		"caa":     dns.TypeCAA,
		"cert":    dns.TypeCERT,
		"dhcid":   dns.TypeDHCID,
		"dname":   dns.TypeDNAME,
		"hinfo":   dns.TypeHINFO,
		"isdn":    dns.TypeISDN,
		"loc":     dns.TypeLOC,
		"mb":      dns.TypeMB,
		"mg":      dns.TypeMG,
		"minfo":   dns.TypeMINFO,
		"mr":      dns.TypeMR,
		"naptr":   dns.TypeNAPTR,
		"nsapptr": dns.TypeNSAPPTR,
		"rp":      dns.TypeRP,
		"rt":      dns.TypeRT,
		"tlsa":    dns.TypeTLSA,
		"x25":     dns.TypeX25,
	}
	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{name, recs[typ], dns.ClassINET}
	return m
}

func query(typ string, name string) (*[]Response, error) {
	message := message(typ, name)
	c := new(dns.Client)
	c.Net = "tcp"
	in, _, err := c.Exchange(message, "8.8.8.8:53")
	if err != nil {
		return nil, err
	}
	replies := make([]Response, 0)
	for _, a := range in.Answer {
		h := a.Header()
		parts := strings.Split(a.String(), "\t")
		replies = append(replies, Response{
			Value: parts[len(parts)-1],
			Name:  h.Name,
			TTL:   h.Ttl,
			Type:  dns.Type(h.Rrtype).String(),
		})
	}
	return &replies, nil
}

func main() {
	// Make the handler available for Remote Procedure Call by AWS Lambda
	replies, _ := query("a", "examplecat.com.")
	for _, a := range *replies {
		fmt.Println(a)
	}
	lambda.Start(handler)
}

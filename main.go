package main

import (
	"fmt"
	"log"
	"time"

	"github.com/miekg/dns"
)

func resolve(domain string, qtype uint16) (answers []dns.RR, err error) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)
	m.RecursionDesired = true

	c := dns.Client{
		Net:            "",
		UDPSize:        0,
		Timeout:        time.Second * 60,
		DialTimeout:    time.Second * 60,
		ReadTimeout:    time.Second * 2,
		WriteTimeout:   time.Second * 2,
		TsigSecret:     map[string]string{},
		TsigProvider:   nil,
		SingleInflight: true,
	}
	in, _, err := c.Exchange(m, "8.8.8.8:53")
	if err != nil {
		return nil, err
	} else {
		return in.Answer, err
	}
}

type dnsHandler struct{}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, question := range r.Question {
		fmt.Printf("Recieved query: %s\n", question.Name)
		answers, err := resolve(question.Name, question.Qtype)
		if err != nil {
			log.Println(err.Error())
			w.Close()
		}
		msg.Answer = append(msg.Answer, answers...)

	}

	w.WriteMsg(msg)
}

func main() {
	handler := new(dnsHandler)
	server := &dns.Server{
		Addr:          ":52",
		Net:           "udp",
		Handler:       handler,
		UDPSize:       65535,
		ReadTimeout:   time.Second * 2,
		WriteTimeout:  time.Second * 2,
		IdleTimeout:   func() time.Duration { return time.Second * 60 },
		TsigProvider:  nil,
		TsigSecret:    map[string]string{},
		MaxTCPQueries: -1,
		ReusePort:     true,
		ReuseAddr:     false,
	}


	log.Printf("Starting DNS server on %s", server.Addr)
	err := server.ListenAndServe(); if err != nil {
		log.Printf("The server crashed: %s\n", err.Error())
	}
}


// -*- Mode: Go; indent-tabs-mode: t -*-
// AMS - Anbox Management Service
// Copyright 2017 Canonical Ltd.  All rights reserved.

package examples

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/anbox-cloud/ams-sdk/client"
	"github.com/anbox-cloud/ams-sdk/shared"
)

// ConnectionCmd defines the options for an example to connect to the service
type ConnectionCmd struct {
	ClientCert string
	ClientKey  string
	ServiceURL string
}

// Parse parses command line arguments
func (c *ConnectionCmd) Parse() {
	flag.StringVar(&c.ClientCert, "cert", "", "Path to the file with the client certificate to use to connect to AMS")
	flag.StringVar(&c.ClientKey, "key", "", "Path to the file with the client key to use to connect to AMS")
	flag.StringVar(&c.ServiceURL, "url", "", "URL of the AMS server")

	flag.Parse()

	if err := c.Validate(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Validate returns error if provided two way SSL parameters are invalid
func (c *ConnectionCmd) Validate() error {
	if len(c.ServiceURL) == 0 {
		return fmt.Errorf("Please provide a service URL")
	}

	if len(c.ClientCert) == 0 {
		return fmt.Errorf("Please provide a certificate path")
	}

	if len(c.ClientKey) == 0 {
		return fmt.Errorf("Please provide a certificate key path")
	}

	if _, err := os.Stat(c.ClientCert); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Specified client cert does not exist")
		}
		return err
	}

	if _, err := os.Stat(c.ClientKey); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Specified client cert key does not exist")
		}
		return err
	}

	return nil
}

// NewClient returns a REST client to connect to AMS
func (c *ConnectionCmd) NewClient() client.Client {
	// Server URL is accessible from client
	u, err := url.Parse(c.ServiceURL)
	if err != nil {
		log.Fatal(err)
	}

	serverCert, err := shared.GetRemoteCertificate(c.ServiceURL)
	if err != nil {
		log.Fatal(err)
	}

	// Asuming that client cert, client key and server cert files exist and are valid
	// certificate files.
	// Server must have client cert amongst trusted client certificates before
	// connecting
	tlsConfig, err := shared.GetTLSConfig(c.ClientCert, c.ClientKey, "", serverCert)
	if err != nil {
		log.Fatal(err)
	}

	amsClient, err := client.New(u, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}

	return amsClient
}
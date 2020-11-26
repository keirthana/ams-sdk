// -*- Mode: Go; indent-tabs-mode: t -*-
// AMS - Anbox Management Service
// Copyright 2018 Canonical Ltd.  All rights reserved.

package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/anbox-cloud/ams-sdk/client"
	"github.com/anbox-cloud/ams-sdk/examples"
)

type appExportCmd struct {
	examples.ConnectionCmd
	id      string
	version string
	target  string
}

func (command *appExportCmd) Parse() {
	flag.StringVar(&command.id, "id", "", "Application id")
	flag.StringVar(&command.version, "version", "", "Application version to export")
	flag.StringVar(&command.target, "target", "", "Output name of exported package")

	command.ConnectionCmd.Parse()

	if len(command.id) == 0 || len(command.version) == 0 {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	cmd := &appExportCmd{}
	cmd.Parse()
	c := cmd.NewClient()

	err := exportApplicationVersion(c, cmd.id, cmd.version, cmd.target)
	if err != nil {
		log.Fatal(err)
	}
}

func exportApplicationVersion(c client.Client, id, version, target string) error {
	versionNum, err := strconv.Atoi(version)
	if err != nil {
		log.Fatal(err)
	}

	file, err := ioutil.TempFile("", "ams_application_export")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	if len(target) == 0 {
		const formatPattern = "2006-01-02-150405"
		target = fmt.Sprintf("%s_%s", id, time.Now().Format(formatPattern))
	}

	return c.ExportApplicationByVersion(id, versionNum, func(header *http.Header, body io.ReadCloser) error {
		if _, err = io.Copy(file, body); err != nil {
			return err
		}

		file.Seek(0, 0)
		hasher := sha256.New()
		_, err = io.Copy(hasher, file)
		if err != nil {
			return err
		}
		imageFingerprint := fmt.Sprintf("%x", hasher.Sum(nil))
		fingerprint := header.Get("X-AMS-Fingerprint")
		if imageFingerprint != fingerprint {
			return fmt.Errorf("Fingerprint doesn't match")
		}

		metaName := header.Get("Content-Disposition")
		extension := strings.SplitN(metaName, ".", 2)[1]
		if !strings.HasSuffix(target, extension) {
			target = fmt.Sprintf("%s.%s", target, extension)
		}
		if err := os.Rename(file.Name(), target); err != nil {
			return err
		}

		print("Application exported successfully!\n")
		return nil
	})
}
/*
 * Copyright 2018-2020, VMware, Inc. All Rights Reserved.
 * Proprietary and Confidential.
 * Unauthorized use, copying or distribution of this source code via any medium is
 * strictly prohibited without the express written consent of VMware, Inc.
 */

package libjvm

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/effect"
)

type CertificateLoader struct {
	KeyTool         string
	SourcePath      string
	DestinationPath string
	Executor        effect.Executor
	Logger          bard.Logger
}

func (c *CertificateLoader) Load() error {
	rest, err := ioutil.ReadFile(c.SourcePath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("unable to read %s\n%w", c.SourcePath, err)
	}

	var (
		block  *pem.Block
		blocks []*pem.Block
	)
	for len(rest) != 0 {
		block, rest = pem.Decode(rest)
		blocks = append(blocks, block)
	}

	c.Logger.Bodyf("Populating with %d container certificates", len(blocks))

	sIn := &bytes.Buffer{}
	sOut := &bytes.Buffer{}
	sErr := &bytes.Buffer{}
	for i, b := range blocks {
		sIn.Reset()
		sOut.Reset()
		sErr.Reset()

		if err := pem.Encode(sIn, b); err != nil {
			return fmt.Errorf("unable to encode certificate\n%w", err)
		}

		if err := c.Executor.Execute(effect.Execution{
			Command: c.KeyTool,
			Args: []string{
				"-importcert", "-trustcacerts", "-noprompt",
				"-alias", fmt.Sprintf("openssl-%03d", i),
				"-keystore", c.DestinationPath,
				"-storepass", "changeit",
			},
			Stdin:  sIn,
			Stdout: sOut,
			Stderr: sErr,
		}); err != nil {
			return fmt.Errorf("unable to invoke %s\n%s\n%s\n%w", c.KeyTool, sOut.String(), sErr.String(), err)
		}
	}

	return nil
}

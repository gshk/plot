// Copyright ©2016 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	baseUrl   = "https://releases.pagure.org/liberation-fonts/"
	fontsName = "liberation-fonts-ttf-2.00.1"
)

func main() {
	log.SetPrefix("mk-vg-fonts: ")
	log.SetFlags(0)

	tmpdir, err := ioutil.TempDir("", "gonum-mk-fonts-")
	if err != nil {
		log.Fatalf("error creating temporary directory: %v\n", err)
	}
	defer os.RemoveAll(tmpdir)

	tarf, err := os.Create(filepath.Join(tmpdir, fontsName+".tar.gz"))
	if err != nil {
		log.Fatalf("error creating local fonts tar file: %v\n", err)
	}
	defer tarf.Close()

	urlSrc := baseUrl + fontsName + ".tar.gz"
	log.Printf("downloading [%v]...\n", urlSrc)
	resp, err := http.DefaultClient.Get(urlSrc)
	if err != nil {
		log.Fatalf("error getting url %v: %v\n", urlSrc, err)
	}
	defer resp.Body.Close()

	err = untar(tmpdir, resp.Body)
	if err != nil {
		log.Fatalf("error untarring: %v\n", err)
	}

	err = exec.Command("go", "get", "github.com/jteeuwen/go-bindata/...").Run()
	if err != nil {
		log.Fatalf("error go-getting go-bindata: %v\n", err)
	}

	fontsDir := getFontsDir()
	enc, err := ioutil.ReadFile(filepath.Join(fontsDir, "cp1252.map"))
	if err != nil {
		log.Fatalf("could not read encoding map: %v", err)
	}
	err = ioutil.WriteFile(filepath.Join(tmpdir, fontsName, "cp1252.map"), enc, 0644)
	if err != nil {
		log.Fatalf("could not write encoding map: %v", err)
	}

	fname := filepath.Join(fontsDir, "liberation_fonts_generated.go")
	log.Printf("generating fonts: %v\n", fname)
	cmd := exec.Command("go-bindata", "-pkg=fonts", "-o", fname, ".")
	cmd.Dir = filepath.Join(tmpdir, fontsName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("error generating asset-data: %v\n", err)
	}

	err = prependHeaders(fname)
	if err != nil {
		log.Fatalf("error prepending headers to [%s]: %v\n", fname, err)
	}

	cmd = exec.Command("gofmt", "-w", fname)
	cmd.Dir = fontsDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = cmd.Stdout
	cmd.Stderr = cmd.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("error running gofmt on %v: %v\n", fname, err)
	}
}

func getFontsDir() string {
	dir := "github.com/gshk/plot/vg"
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Fatal("no GOPATH environment variable")
	}

	for _, p := range strings.Split(gopath, string(os.PathListSeparator)) {
		if p == "" {
			continue
		}
		n := filepath.Join(p, "src", dir, "fonts")
		_, err := os.Stat(n)
		if err != nil {
			continue
		}
		return n
	}
	log.Fatal("could not find %q anywhere under $GOPATH", dir)
	panic("unreachable")
}

func untar(odir string, r io.Reader) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	// Iterate through the files in the archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			dir := filepath.Join(odir, hdr.Name)
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
			continue

		case tar.TypeReg, tar.TypeRegA:
			// ok
		default:
			log.Printf("error: %v\n", hdr.Typeflag)
			return err
		}
		oname := filepath.Join(odir, hdr.Name)
		dir := filepath.Dir(oname)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		o, err := os.OpenFile(
			oname,
			os.O_WRONLY|os.O_CREATE,
			os.FileMode(hdr.Mode),
		)
		if err != nil {
			return err
		}
		defer o.Close()
		_, err = io.Copy(o, tr)
		if err != nil {
			return err
		}
		err = o.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func prependHeaders(name string) error {
	src, err := os.Open(name)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(name + ".tmp")
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = dst.WriteString(`// Automatically generated by vg/fonts/mk-fonts.go
// DO NOT EDIT.

// Copyright ©2016 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Digitized data copyright (c) 2010 Google Corporation
//         with Reserved Font Arimo, Tinos and Cousine.
// Copyright (c) 2012 Red Hat, Inc.
//         with Reserved Font Name Liberation.
//
// This Font Software is licensed under the SIL Open Font License,
// Version 1.1.

`)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	err = src.Close()
	if err != nil {
		return err
	}

	err = dst.Close()
	if err != nil {
		return err
	}

	return os.Rename(dst.Name(), src.Name())
}

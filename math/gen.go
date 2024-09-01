package main

import (
	"embed"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	//go:embed *
	srcmath        embed.FS
	replaceWithF64 = [][2]string{
		{"ms1", "md1"},
		{"ms2", "md2"},
		{"ms3", "md3"},
	}
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("generated files")
}

func run() error {
	for _, rep := range replaceWithF64 {
		files, err := srcmath.ReadDir(rep[0])
		if err != nil {
			return err
		}
		os.RemoveAll(rep[1])
		err = os.MkdirAll(rep[1], 0777)
		if err != nil {
			return err
		}
		for _, file := range files {

			src, err := srcmath.Open(filepath.Join(rep[0], file.Name()))
			if err != nil {
				return err
			}
			newName := strings.ReplaceAll(file.Name(), rep[0], rep[1])
			dst, err := os.Create(filepath.Join(rep[1], newName))
			if err != nil {
				return err
			}
			defer dst.Close()
			b, err := io.ReadAll(src)
			if err != nil {
				return err
			}

			repr := strings.NewReplacer(
				"float32", "float64",
				"package "+rep[0], "package "+rep[1],
				"\"github.com/chewxy/math32\"", "\"math\"",
				"\"github.com/soypat/glgl/math/ms1\"", "ms1 \"github.com/soypat/glgl/math/md1\"",
			)
			dst.WriteString(`// DO NOT EDIT.
// This file was generated automatically
// from gen.go. Please do not edit this file.
`)
			_, err = repr.WriteString(dst, string(b))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

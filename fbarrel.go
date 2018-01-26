package main

import (
	"fmt"
	"os"
	"github.com/jessevdk/go-flags"
	"io/ioutil"
	"bufio"
	"path"
	"strings"
)

type opts struct {
	Path string `short:"p" long:"path" description:"Path to typescript folder where barrel should be created" required:"true"`
	Name string `short:"n" long:"name" description:"Name of barrel (omit .ts) - will be uppercased for namespace name" required:"true"`
}

func fatal(err error){
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	var opts = opts{}

	_, err := flags.Parse(&opts)
	fatal(err)

	files,err := listFiles(opts.Path)
	fatal(err)
	err = writeBarrel(opts.Path, files)
	fatal(err)

	err = writeNamespace(opts.Name, opts.Path)

	os.Exit(0)
}

func writeNamespace(name string, ts_path string) error {
	fd, err := os.Create(path.Join(ts_path, fmt.Sprintf("%s.ts", name))); if err != nil {
		return err
	}
	namespace := strings.Title(name)
	fmt.Printf("Writing namespace { %s }\n", namespace)
	defer fd.Close()
	w := bufio.NewWriter(fd)

		_, err = w.WriteString(fmt.Sprintf("import * as %s from './barrel';\nexport { %s };\n", namespace, namespace)); if err != nil {
			return err
		}
	w.Flush()
	return nil
}

func writeBarrel(ts_path string, files []os.FileInfo) error {
	fd, err := os.Create(path.Join(ts_path, "barrel.ts")); if err != nil {
		return err
	}
	defer fd.Close()
	w := bufio.NewWriter(fd)

	for _, f := range files {
		var name = f.Name()
		if(strings.HasPrefix(name, ".") || ! strings.HasSuffix(name, ".ts") || name == "barrel.ts"){ continue }
		name_without_ext := name[0:strings.LastIndex(name, ".ts")]
		fmt.Printf("Writing to barrel for %s (%s)\n", name_without_ext, name)
		_, err = w.WriteString(fmt.Sprintf("export * from './%s';\n", name_without_ext)); if err != nil {
			return err
		}
	}
	w.Flush()

	return nil
}

func listFiles(ts_path string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(ts_path)
	if err != nil {
		return nil,err
	}
	return files,nil
}

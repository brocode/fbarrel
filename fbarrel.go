package main

import (
	"fmt"
	"os"
	"github.com/jessevdk/go-flags"
	"io/ioutil"
	"bufio"
	"path"
	"strings"
	"regexp"
	"Set"
	"github.com/deckarep/golang-set"
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
  dir, err := os.Getwd()
	fatal(err)
	err = writeBarrel(dir, opts.Path, files)
	fatal(err)

	err = writeNamespace(opts.Name, dir)

	os.Exit(0)
}

func writeNamespace(name string, out_path string) error {
	fd, err := os.Create(path.Join(out_path, fmt.Sprintf("%s.ts", name))); if err != nil {
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

func writeBarrel(out_path string, ts_path string, files []os.FileInfo) error {
	fd, err := os.Create(path.Join(out_path, "barrel.ts")); if err != nil {
		return err
	}
	defer fd.Close()
	w := bufio.NewWriter(fd)

	for _, f := range files {
		var name = f.Name()

		if(strings.HasPrefix(name, ".") || ! strings.HasSuffix(name, ".tsx") || name == "barrel.ts"){ continue }

		content, err := ioutil.ReadFile(ts_path + "/" + f.Name())
		if err != nil {
			return err;
		}

		extractExports(string(content[:]));

		name_without_ext := name[0:strings.LastIndex(name, ".tsx")]
		default_name := strings.Title(strings.Replace(strings.Replace(name_without_ext, "_", "", -1), "-", "", -1))
		fmt.Printf("Writing to barrel for %s (%s)\n", name_without_ext, name)
		_, err = w.WriteString(fmt.Sprintf("import %s from './%s';\n", default_name, ts_path + "/" + name_without_ext)); if err != nil {
			return err
		}
		_, err = w.WriteString(fmt.Sprintf("export { %s };\n", default_name)); if err != nil {
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

func extractExports(content string) Set {
	exports := mapset.NewSet();

	defaultResult := regexp.MustCompile(`export default (class|interface|type )?(\w+)`).FindAllStringSubmatch(content, -1)
	for _, value := range defaultResult {
		exports.Add(value[2])
	}

	regularResult := regexp.MustCompile(`export (class|interface|type) (\w+)`).FindAllStringSubmatch(content, -1)
	for _, value := range regularResult {
		if !exports.Contains(value[2]) && value[2] != "Props" && value[2] != "State" { exports.Add(value[2]) }
	}

	fmt.Println(exports);

	return exports
}

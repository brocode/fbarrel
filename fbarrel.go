package main

import (
	"fmt"
	"os"
	"github.com/jessevdk/go-flags"
	"io/ioutil"
	"bufio"
	"path"
	"sort"
	"strings"
	"regexp"
	"github.com/deckarep/golang-set"
)

type opts struct {
	Path string `short:"p" long:"path" description:"Path to typescript folder where barrel should be created" required:"true"`
	Recursive bool `short:"r" long:"recursive" description:"Scans given directory recursively" required:"false"`
}

func toFullPath(path string, files []os.FileInfo) []string {
	var result []string;

	for _,file := range files {
		result = append(result, path + file.Name());
	}

	return result;
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

	files,err := listFiles(opts.Path, opts.Recursive)
	fatal(err)
  dir, err := os.Getwd()
	fatal(err)
	err = writeBarrel(dir, opts.Path, files)
	fatal(err)

	os.Exit(0)
}

func writeBarrel(out_path string, ts_path string, files []string) error {
	fd, err := os.Create(path.Join(out_path, "barrel.ts")); if err != nil {
		return err
	}
	defer fd.Close()
	w := bufio.NewWriter(fd)

	for _, f := range files {
		var name = f

		if(! strings.HasSuffix(name, ".tsx") || strings.Contains(name, "barrel.ts") ){ continue }

		content, err := ioutil.ReadFile(f)
		if err != nil {
			return err;
		}

		exports := extractExports(string(content[:]));
    sortedExports := setToSortedArray(exports)

		exportList := ""
		for i,ex := range sortedExports {
			if i==0 {
				exportList = ex
			} else {
				exportList = exportList + ", " + ex;
			}
		}

		if exportList != "" {
			name_without_ext := name[0:strings.LastIndex(name, ".tsx")]
			fmt.Printf("Writing to barrel for %s (%s)\n", name_without_ext, name)
			_, err = w.WriteString(fmt.Sprintf("export { %s }  from '%s';\n", exportList, name_without_ext)); if err != nil {
				return err
			}
		}
	}
	w.Flush()

	return nil
}

func setToSortedArray(set mapset.Set) []string {
	var setArray []string
	for item := range set.Iterator().C {
    setArray = append(setArray, item.(string));
	}
	sort.Strings(setArray)
	return setArray;
}

func listFiles(ts_path string, recursive bool) ([]string, error) {
	var result []string;
	files, err := ioutil.ReadDir(ts_path)

	if err != nil {
		return nil,err
	}

	if recursive {
		for _, file := range files {
			if(file.IsDir()) {
				fmt.Printf("Scanning folder %s\n", file.Name())
				recResult, recError := listFiles(ts_path + "/" + file.Name(), true);
				if recError != nil {
					fatal(recError);
				}
				result = append(result, recResult...)
			} else {
				result = append(result, ts_path +"/"+ file.Name())
			}
		}
	} else {
		result = toFullPath(ts_path, files);
	}

	return result,nil
}

func extractExports(content string) mapset.Set {
	exports := mapset.NewSet();
  defaultExportName := ""

	defaultResult := regexp.MustCompile(`export default (class|interface|type )?(\w+)`).FindAllStringSubmatch(content, -1)
	for _, value := range defaultResult {
		defaultExportName = value[2]
		exports.Add("default as " + value[2])
	}

	regularResult := regexp.MustCompile(`export (class|interface|type) (\w+)`).FindAllStringSubmatch(content, -1)
	for _, value := range regularResult {
		if !exports.Contains(value[2]) && value[2] != defaultExportName {
			if value[2] != "Props" && value[2] != "State" {
				exports.Add(value[2])
			} else if (value[2] == "Props" || value[2] == "State") && defaultExportName != "" {
				exports.Add(value[2] + " as " + defaultExportName+value[2])
			}
	  }
	}

	return exports
}

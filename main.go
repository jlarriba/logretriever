package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	exportFolder string
	logsFolder   string
)

func main() {
	fmt.Printf("Logging Main process \n")
	//logsFolder = os.Getenv("LOGS_FOLDER")
	//exportFolder = os.Getenv("EXPORT_FOLDER")
	interval, err := strconv.Atoi(os.Getenv("EXEC_INTERVAL"))
	if err != nil {
		interval = 7
	}
	if logsFolder == "" {
		logsFolder = "/var/log/containers"
	}
	if exportFolder == "" {
		exportFolder = "/opt/logs"
	}

	go gather(interval)
	runWebServer()
}

func runWebServer() {
	http.Handle("/", http.FileServer(http.Dir(exportFolder)))
	http.HandleFunc("/healthz", healthz)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func gather(interval int) {
	for {
		fmt.Println("Gathering logs...")
		err2 := filepath.Walk(logsFolder, visit)
		if err2 != nil {
			fmt.Println("Error visiting file", err2)
		}
		fmt.Printf("Pausing the process during %d minutes \n", interval)
		time.Sleep(time.Minute * time.Duration(interval))
	}
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func visit(path string, f os.FileInfo, err error) error {
        if err != nil {
		fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
		return err
	}
	if f != nil && !strings.Contains(f.Name(), "POD") {
		if f.IsDir() {
			fmt.Println(path, "is a directory, ignore")
		}
		realPath, errSymLink := filepath.EvalSymlinks(path)
		if errSymLink != nil {
			fmt.Println(path, " is not a SymLink")
			realPath = path
		}

		sName := strings.Split(f.Name(), "_")
		if sName[1] != "kube-system" {
			sNamespace := strings.Split(sName[1], "-")
			if len(sNamespace) != 2 {
				err = copy(realPath, exportFolder+"/"+sNamespace[0]+"/pro/"+sName[2])
			} else {
				err = copy(realPath, exportFolder+"/"+sNamespace[0]+"/"+sNamespace[1]+"/"+sName[2])
			}

			if err != nil {
				fmt.Println("Error copying file", realPath, err)
			}
		} else {
			fmt.Println(sName[1], "Ignoring namespace")
			return nil
		}
	}
	return err
}

func copy(src, dst string) error {
	fmt.Printf("Opening file to be copied %s \n", src)
	in, err := os.Open(src)
	if err != nil {
		in.Close()
		return err
	}
	//defer in.Close() ==>Avoiding to use defer in order to close files
	fmt.Printf("Creating destination file %s \n", dst)
	os.MkdirAll(filepath.Dir(dst), os.ModePerm)
	out, err := os.Create(dst)
	if err != nil {
		in.Close()
		out.Close()
		return err
	}
	//defer out.Close() ==>Avoiding to use defer in order to close files
	fmt.Printf("Copying from %s to %s\n", src, dst)
	bytesWritten, err := io.Copy(out, in)
	if err != nil {
		in.Close()
		out.Close()
		return err
	}
	fmt.Printf("Copied %d bytes. \n", bytesWritten)
	fmt.Printf("Copied from %s to %s\n", src, dst)
	//cerr := out.Close()
	//return cerr
	in.Close()
	out.Close()
	return err
}

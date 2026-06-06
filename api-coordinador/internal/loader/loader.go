package loader

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
)

func LeerCSVMasivo(filePath string, jobs chan<- []string) {
	file, _ := os.Open(filePath)
	defer file.Close()

	bufferedReader := bufio.NewReaderSize(file, 64*1024*1024)
	reader := csv.NewReader(bufferedReader)
	_, _ = reader.Read()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		jobs <- record
	}

	close(jobs)
}

package loader

import (
	"encoding/csv"
	"io"
	"os"
)

func LeerCSVMasivo(filePath string, jobs chan<- []string) {
	file, _ := os.Open(filePath)
	defer file.Close()

	reader := csv.NewReader(file)
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

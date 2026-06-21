package loader

import (
	"bufio"
	"encoding/csv"
	"io"
)

func LeerCSVMasivo(r io.Reader, jobs chan<- []string) {
	bufferedReader := bufio.NewReaderSize(r, 64*1024*1024)
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

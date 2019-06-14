package utils

import "os"

type VolatileTempFile struct {
	file *os.File
}

func (f VolatileTempFile) Read(p []byte) (n int, err error) {
	return f.file.Read(p)
}

func (f VolatileTempFile) Close() error {
	defer func() { _ = os.Remove(f.file.Name()) }()
	return f.file.Close()
}

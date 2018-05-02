package unite

import (
	"fmt"
	"os"
)

func WritePid(path string) error {
	pid := os.Getpid()
	f, err := os.Create(path)
	if err != nil {
		fmt.Println("writepid err :", err)
		return err
	}
	fmt.Println("pid=", pid)
	_, err = f.WriteString(fmt.Sprintf("%d", pid))

	if err != nil {
		fmt.Println("writepid.go err:", err)
		return err
	}

	f.Sync()
	return nil
}

package db

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	convertAndWriteToFile(File1+"\n"+File4, "./A.yinao.txt")
	convertAndWriteToFile(File2+"\n"+File3, "./B.yinao.txt")
	db, err := NewDBFromFiles([]string{"./A.yinao.txt", "./B.yinao.txt"})
	assert.Equal(t, nil, err)

	infoList := []string{"张若虚，男，2019",
		"张若美，女，2018",
		"张若虚，男，1999",
		"张若，男，2010",
		"李若美，女，1988",
		"张若，男，2010",
		"张若美，女，2018",
	}
	for _, info := range infoList {
		h := sha256.Sum256([]byte(info))
		recList, err := db.QueryBaseInfo(h)
		assert.Equal(t, nil, err)
		for _, rec := range recList {
			fmt.Printf("%s\n", strings.Join(rec.ToLines(), "\n"))
		}
	}

	idList := []string{"11010920190401911X",
		"11010920190401811X",
		"11010920180401911X",
	}
	for _, id := range idList {
		h := sha256.Sum256([]byte(id))
		recList, err := db.QueryID(h)
		assert.Equal(t, nil, err)
		for _, rec := range recList {
			fmt.Printf("%s\n", strings.Join(rec.ToLines(), "\n"))
		}
	}
	os.RemoveAll("./A.yinao.txt")
	os.RemoveAll("./B.yinao.txt")
}

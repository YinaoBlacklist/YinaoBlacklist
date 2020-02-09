package db

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var File1 = `
张若虚，男，2019
11010920190401911X
19.0
春江潮水连海平，海上明月共潮生。
滟滟随波千万里，何处春江无月明？

张若美，女，2018
11010920180401911X
98.0
江流宛转绕芳甸，月照花林皆似霰。
空里流霜不觉飞，汀上白沙看不见。`

var File2 = `
张若虚，男，1999
11010920190401811X
97.0
春江潮水连海平，海上明月共潮生。
滟滟随波千万里，何处春江无月明？

张若美，女，2018
11010920180401911X
94.0
江流宛转绕芳甸，月照花林皆似霰。
空里流霜不觉飞，汀上白沙看不见。`

var File3 = `
张若虚，男，2019
11010920190401911X
99.0
春江潮水连海平，海上明月共潮生。
滟滟随波千万里，何处春江无月明？

张若，男，2010
11010920190401911X
97.0
举头望明月，低头思故乡。
露从今夜白，月是故乡明。

李若美，女，1988
NA
8.0
江流宛转绕芳甸，月照花林皆似霰。
空里流霜不觉飞，汀上白沙看不见。`

var File4 = `
张若，男，2010
NA
97.0
春江潮水连海平，海上明月共潮生。
滟滟随波千万里，何处春江无月明？

张若美，女，2018
11010920180401911X
98.0
空里流霜不觉飞，汀上白沙看不见。`

func TestConvert(t *testing.T) {
	txt := `
张若虚，男，2019
11010920190401911X
99.0

张若美，女，2018
11010920180401911X
98.0
江流宛转绕芳甸，月照花林皆似霰。
空里流霜不觉飞，汀上白沙看不见。`
	err := ioutil.WriteFile("./dat.txt", []byte(txt), 0644)
	assert.Equal(t, nil, err)

	_, err = ExtractRecordsFromRawFile("./dat.txt")
	assert.NotEqual(t, nil, err)

	err = ioutil.WriteFile("./dat.txt", []byte(File1), 0644)
	assert.Equal(t, nil, err)

	recList, err := ExtractRecordsFromRawFile("./dat.txt")
	out, err := os.OpenFile("./dat.byn.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)

	for _, rec := range recList {
		assert.Equal(t, true, rec.VerifyChecksum())
		recB := *rec
		recB.Confidence = 1.0
		assert.Equal(t, true, rec.IsSame(recB))
	}

	err = WriteRecordsToFile(recList, out)
	assert.Equal(t, nil, err)
	out.Close()

	dat, err := ioutil.ReadFile("./dat.byn.txt")
	assert.Equal(t, nil, err)
	fmt.Print(string(dat))

	err = os.Remove("./dat.txt")
	assert.Equal(t, nil, err)
	err = os.Remove("./dat.byn.txt")
	assert.Equal(t, nil, err)

	res := `77tvf7/2R95SVd7b0p6V05SVgvPZc+aspC7oj5Q5WOY=
qRrk3oLyVGxF1zpO6bW4n0E4INNTxTy+8p+FRj0H2jE=
19.000000
春江潮水连海平，海上明月共潮生。\n滟滟随波千万里，何处春江无月明？
770a6955

NodIhJ4+FFFrYJmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
uKZuKBb825p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
98.000000
江流宛转绕芳甸，月照花林皆似霰。\n空里流霜不觉飞，汀上白沙看不见。
ea839b7a

`
	assert.Equal(t, res, string(dat))
}

func TestCheckBaseInfo(t *testing.T) {
	err := CheckBaseInfo("李四，女，1990")
	assert.Equal(t, nil, err)
	err = CheckBaseInfo("张三,男，1943")
	assert.NotEqual(t, nil, err)
	err = CheckBaseInfo("张三，nan，1943")
	assert.NotEqual(t, nil, err)
	err = CheckBaseInfo("张三，男，一九四二")
	assert.NotEqual(t, nil, err)
	err = CheckBaseInfo("张三，男，1911")
	assert.NotEqual(t, nil, err)
	err = CheckBaseInfo("张三，男，2051")
	assert.NotEqual(t, nil, err)
}

func TestCheckID(t *testing.T) {
	err := CheckID("110109201904019110X")
	assert.NotEqual(t, nil, err)
	err = CheckID("A1010920190401911X")
	assert.NotEqual(t, nil, err)
	err = CheckID("11010920190401911A")
	assert.NotEqual(t, nil, err)
	err = CheckID("NA")
	assert.Equal(t, nil, err)
}

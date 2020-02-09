package db

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func convertAndWriteToFile(inTxt string, outFile string) {
	err := ioutil.WriteFile("./in.txt", []byte(inTxt), 0644)
	if err != nil {
		panic(err)
	}
	recList, err := ExtractRecordsFromRawFile("./in.txt")
	if err != nil {
		panic(err)
	}
	out, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	err = WriteRecordsToFile(recList, out)
	if err != nil {
		panic(err)
	}
	os.RemoveAll("./in.txt")
}

func prepareInput() {
	os.MkdirAll("./a", os.ModePerm)
	os.MkdirAll("./b/c", os.ModePerm)
	convertAndWriteToFile(File1, "./a/1.yinao.txt")
	convertAndWriteToFile(File2, "./a/2.yinao.txt")
	convertAndWriteToFile(File3, "./b/3.yinao.txt")
	convertAndWriteToFile(File4, "./b/c/4.yinao.txt")
}

var result = `NodIhJ4+FFFrYJmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
uKZuKBb825p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
88.000000
空里流霜不觉飞，汀上白沙看不见。
3812db55

pE/xp/CKOkj6X6sqPIM2iBqO6vC2YUAn2EYrNurHejE=
IO8PDI0O6ph3JBLOqbO5JhLj5Ty15ZFStXAxZfVuilM=
0.000000
江流宛转绕芳甸，月照花林皆似霰。\n空里流霜不觉飞，汀上白沙看不见。
4eeed005

QkedkZm8Eq2KyQg/DeQWN+b92Z2M651NfpxF/9UCmVU=
IO8PDI0O6ph3JBLOqbO5JhLj5Ty15ZFStXAxZfVuilM=
87.000000
春江潮水连海平，海上明月共潮生。\n滟滟随波千万里，何处春江无月明？
5e437216

77tvf7/2R95SVd7b0p6V05SVgvPZc+aspC7oj5Q5WOY=
qRrk3oLyVGxF1zpO6bW4n0E4INNTxTy+8p+FRj0H2jE=
89.000000
春江潮水连海平，海上明月共潮生。\n滟滟随波千万里，何处春江无月明？
770a6955

mZyjgFob2OSxIqbUXIK9vHqdmttJJSsJ/xbm++tAfZs=
fOf06IsxgCcL4uarx8i4D4r6hyuNPRZll116H+v7pl4=
100.000000
春江潮水连海平，海上明月共潮生。\n滟滟随波千万里，何处春江无月明？
be038bab

QkedkZm8Eq2KyQg/DeQWN+b92Z2M651NfpxF/9UCmVU=
qRrk3oLyVGxF1zpO6bW4n0E4INNTxTy+8p+FRj0H2jE=
87.000000
举头望明月，低头思故乡。\n露从今夜白，月是故乡明。
c2af2fa4

NodIhJ4+FFFrYJmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
uKZuKBb825p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
100.000000
江流宛转绕芳甸，月照花林皆似霰。\n空里流霜不觉飞，汀上白沙看不见。
ea839b7a

`

func TestMerge(t *testing.T) {
	prepareInput()
	records := NewRecords()
	logfile, _ := os.OpenFile("./log.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	records.AddEncRecordsInDir("./a", 10, logfile)
	records.AddEncRecordsInDir("./b", -10, logfile)

	fmt.Printf("%v\n", records.m)
	outfile, _ := os.OpenFile("./out.yinao.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	err := records.WriteToFile(outfile)
	assert.Equal(t, nil, err)
	outfile.Close()
	logfile.Close()

	dat, err := ioutil.ReadFile("./out.yinao.txt")
	assert.Equal(t, nil, err)
	assert.Equal(t, result, string(dat))

	os.RemoveAll("./a")
	os.RemoveAll("./b")
	os.RemoveAll("./out.yinao.txt")
	os.RemoveAll("./log.txt")
}

func TestParseLines(t *testing.T) {
	txt := `NodIhJ4+FFFrYJmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
uKZuKBb825p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
江流宛转绕芳甸，月照花林皆似霰。\n空里流霜不觉飞，汀上白沙看不见。
ea839b7a`
	logfile, _ := os.OpenFile("./log.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	parseLines(strings.Split(txt, "\n"), logfile)

	txt = `NodIhJ4+FFFrY@*JmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
uKZuKBb825p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
100.000000
江流宛转绕芳甸，月照花林皆似霰。\n空里流霜不觉飞，汀上白沙看不见。
ea839b7a`
	parseLines(strings.Split(txt, "\n"), logfile)

	txt = `NodIhJ4IhJ4+FFFrYJmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
uKZuKBb825p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
100.000000
江流宛转绕芳甸，月照花林皆似霰。\n空里流霜不觉飞，汀上白沙看不见。
ea839b7a`
	parseLines(strings.Split(txt, "\n"), logfile)

	txt = `NodIhJ4+FFFrYJmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
uKZuKBb82*5p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
100.000000
江流宛转绕芳甸，月照花林皆似霰。\n空里流霜不觉飞，汀上白沙看不见。
ea839b7a`
	parseLines(strings.Split(txt, "\n"), logfile)

	txt = `NodIhJ4+FFFrYJmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
uKZuKBb825p2v5p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
100.000000
江流宛转绕芳甸，月照花林皆似霰。\n空里流霜不觉飞，汀上白沙看不见。
ea839b7a`
	parseLines(strings.Split(txt, "\n"), logfile)

	txt = `NodIhJ4+FFFrYJmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
uKZuKBb825p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
100.100000
江流宛转绕芳甸，月照花林皆似霰。\n空里流霜不觉飞，汀上白沙看不见。
ea839b7a`
	parseLines(strings.Split(txt, "\n"), logfile)

	txt = `NodIhJ4+FFFrYJmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
uKZuKBb825p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
100.00000
江转绕芳甸，月照花林皆似霰。\n空里流霜不觉飞，汀上白沙看不见。
ea839b7a`
	parseLines(strings.Split(txt, "\n"), logfile)

	txt = `NodIhJ4+FFFrYJmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
uKZuKBb825p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
100.00000
江流宛转绕芳甸，月照花林皆似霰。\n空里流霜不觉飞，汀上白沙看不见。
ea83-9b7a`
	parseLines(strings.Split(txt, "\n"), logfile)

	logfile.Close()

	result := `记录的长度错误，必须正好有5行：NodIhJ4+FFFrYJmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
uKZuKBb825p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
江流宛转绕芳甸，月照花林皆似霰。\n空里流霜不觉飞，汀上白沙看不见。
ea839b7a
base64编码错误：NodIhJ4+FFFrY@*JmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
哈希值长度错误：NodIhJ4IhJ4+FFFrYJmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
base64编码错误：uKZuKBb82*5p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
哈希值长度错误：uKZuKBb825p2v5p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
100.100000 太大了
校验码错误：NodIhJ4+FFFrYJmBrfMwB/VaxZwfsGcCaeHzyL/cZjk=
uKZuKBb825p2vFxrb4iapZj5v8K4GCVmd6VWY8y5bw4=
100.00000
江转绕芳甸，月照花林皆似霰。\n空里流霜不觉飞，汀上白沙看不见。
ea839b7a
校验码编码格式错误：ea83-9b7a
`
	dat, err := ioutil.ReadFile("./log.txt")
	assert.Equal(t, nil, err)
	assert.Equal(t, result, string(dat))

	os.RemoveAll("./log.txt")
}

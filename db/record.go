package db

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"strconv"
	"strings"
)

// 一条医闹记录
type Record struct {
	BaseInfoHash [sha256.Size]byte // 基本信息（姓名，性别，出生年份）的哈希值
	IDHash       [sha256.Size]byte // 身份证号码的哈希值
	Confidence   float32           // 置信参数
	Description  string            // 对医闹行为的描述
	Crc32        uint32            // 用BaseInfoHash, IDHash, Description生成的校验码
}

func NewRecord(baseInfo string, id string, confidence float32, description string) *Record {
	rec := &Record{
		BaseInfoHash: sha256.Sum256([]byte(baseInfo)),
		IDHash:       sha256.Sum256([]byte(id)),
		Confidence:   confidence,
		Description:  description,
	}
	h := crc32.NewIEEE()
	h.Write(rec.BaseInfoHash[:])
	h.Write(rec.IDHash[:])
	h.Write([]byte(rec.Description))
	rec.Crc32 = h.Sum32()
	return rec
}

// 得到此条记录的校验码
func (rec *Record) VerifyChecksum() bool {
	h := crc32.NewIEEE()
	h.Write(rec.BaseInfoHash[:])
	h.Write(rec.IDHash[:])
	h.Write([]byte(rec.Description))
	return h.Sum32() == rec.Crc32
}

// 将记录转为5行纯文本
func (rec *Record) ToLines() []string {
	var lines [5]string
	lines[0] = base64.StdEncoding.EncodeToString(rec.BaseInfoHash[:])
	lines[1] = base64.StdEncoding.EncodeToString(rec.IDHash[:])
	lines[2] = fmt.Sprintf("%f", rec.Confidence)
	lines[3] = rec.Description
	lines[4] = fmt.Sprintf("%08x", rec.Crc32)
	return lines[:]
}

func WriteRecordsToFile(recList []*Record, file io.Writer) (err error) {
	for _, rec := range recList {
		for _, line := range rec.ToLines() {
			_, err = file.Write([]byte(line))
			if err != nil {
				return
			}
			_, err = file.Write([]byte("\n"))
			if err != nil {
				return
			}
		}
		_, err = file.Write([]byte("\n"))
		if err != nil {
			return
		}
	}
	return
}

// rec和other是同一条医闹记录（它们只有置信参数不同）
func (rec *Record) IsSame(other Record) bool {
	return bytes.Equal(rec.BaseInfoHash[:], other.BaseInfoHash[:]) &&
		bytes.Equal(rec.IDHash[:], other.IDHash[:]) &&
		rec.Description == other.Description
}

// 将基本信息中的年份减一和加一，形成两个新的基本信息
func BaseInfoToAdjacentYears(info string) [2]string {
	parts := strings.Split(info, "，")
	year, _ := strconv.Atoi(parts[2])
	return [2]string{
		fmt.Sprintf("%s，%s，%d", parts[0], parts[1], year-1),
		fmt.Sprintf("%s，%s，%d", parts[0], parts[1], year+1),
	}
}

// 检查基本信息的输入是否有误
func CheckBaseInfo(info string) error {
	parts := strings.Split(info, "，")
	if len(parts) != 3 {
		return fmt.Errorf("基本信息的格式是：姓名，性别，出生年份。注意中间要用两个中文逗号隔开。错误输入%s", info)
	}
	if parts[1] != "男" && parts[1] != "女" {
		return fmt.Errorf("性别必须是 男 或者 女。错误输入：%s", info)
	}
	year, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("%s 无法被转为一个年份。", parts[2])
	}
	if year <= 1911 {
		return fmt.Errorf("出生年份 %s 太小了。", info)
	}
	if year >= 2051 {
		return fmt.Errorf("出生年份 %s 太大了。", info)
	}
	return nil
}

// 检查身份证号的输入是否有误
func CheckID(id string) error {
	if id == "NA" {
		return nil
	}
	runes := []rune(id)
	if len(runes) != 18 {
		return fmt.Errorf("身份证号 %s 不是18位。", id)
	}
	last := len(runes) - 1
	for _, r := range runes[:last] {
		if '0' <= r && r <= '9' {
			continue
		}
		return fmt.Errorf("身份证号 %s 格式错误，非法字符：%s", id, string(r))
	}
	r := runes[last]
	if '0' <= r && r <= '9' || 'X' == r {
		return nil
	} else {
		return fmt.Errorf("身份证号 %s 格式错误，最后一个字符必须是数字0～9或者字母X", id)
	}
}

// 解析置信参数
func parseConfidence(conf string) (float32, error) {
	s, err := strconv.ParseFloat(conf, 32)
	if err != nil {
		return 0.0, fmt.Errorf("%s 不是一个合法的数字", conf)
	}
	if s > 100.0 {
		return 0.0, fmt.Errorf("%s 太大了", conf)
	}
	if s < 0.0 {
		return 0.0, fmt.Errorf("%s 太小了", conf)
	}
	return float32(s), nil
}

// 将若干行的原始医闹记录，转换为一个Record
func parseRawLines(recLines []string) (*Record, error) {
	if len(recLines) < 4 {
		return nil, fmt.Errorf("记录太短了，必须至少有四行：%s", strings.Join(recLines, "\n"))
	}
	//第一行是患者基本信息（姓名，性别，出生年份），用中文逗号隔开
	err := CheckBaseInfo(recLines[0])
	if err != nil {
		return nil, err
	}
	//第二行是患者的身份证号，如果无法提供则以NA代替
	err = CheckID(recLines[1])
	if err != nil {
		return nil, err
	}
	//第三行是置信指数，它是一个百分数，最大为100，最小为0，
	conf, err := parseConfidence(recLines[2])
	if err != nil {
		return nil, err
	}
	//其他行是对于患者医闹记录的文本描述
	description := strings.Join(recLines[3:], "\\n")
	return NewRecord(recLines[0], recLines[1], conf, description), nil
}

// 从文本文件中读取医闹记录，并且将它们转换为Record列表
// fn用来解析记录的文本，根据读取的记录的类型（原始的还是加密的），可采用不同的fn
func extractRecordsFromFile(fname string, fn func(recLines []string, off int64) error) error {
	file, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer file.Close()

	recLines := make([]string, 0, 20)
	offset := int64(0)
	start := int64(0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(recLines) == 0 { //一条记录的开始的位置记录在start中
			start = offset
		}
		offset += int64(len(line) + 1)
		line = strings.TrimSpace(line)
		if len(line) == 0 { //空行标志着一条记录的结束
			if len(recLines) == 0 {
				continue //遇到连续的空行
			}
			err := fn(recLines, start)
			if err != nil {
				return err
			}
			recLines = recLines[:0] // 清空其内容
		} else { //没有遇到空行，则一条记录继续新增行
			recLines = append(recLines, line)
		}
	}
	if len(recLines) != 0 {
		err := fn(recLines, start)
		if err != nil {
			return err
		}
	}
	return nil
}

// 从文本文件中读取原始医闹记录，并且将它们转换为Record列表
func ExtractRecordsFromRawFile(fname string) ([]*Record, error) {
	res := make([]*Record, 0, 100)
	err := extractRecordsFromFile(fname, func(recLines []string, off int64) error {
		rec, err := parseRawLines(recLines)
		if err != nil {
			return err
		}
		res = append(res, rec)
		return nil
	})
	return res, err
}

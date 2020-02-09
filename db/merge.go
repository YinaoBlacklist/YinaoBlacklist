package db

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"sort"
	"strings"
)

const (
	EncFileSuffix = ".yinao.txt"
)

type Records struct {
	m map[uint32][]*Record
}

func NewRecords() *Records {
	return &Records{m: make(map[uint32][]*Record)}
}

// 增加一条新的记录
func (recs *Records) Add(rec Record, confDelta float32) {
	rec.Confidence = rec.Confidence + confDelta
	if rec.Confidence > 100.0 {
		rec.Confidence = 100.0
	}
	if rec.Confidence < 0.0 {
		rec.Confidence = 0.0
	}
	recList, ok := recs.m[rec.Crc32]
	if !ok { //如果尚没有校验码的记录，则直接插入
		recs.m[rec.Crc32] = []*Record{&rec}
		return
	}
	foundSame := false
	//扫描已有的、校验码相同的记录
	for i, oldRec := range recList {
		if oldRec.IsSame(rec) { //如果存在相同内容的记录
			if oldRec.Confidence < rec.Confidence {
				//留下置信度更高的那条记录
				recList[i] = &rec
			}
			foundSame = true
			break
		}
	}
	if !foundSame {
		//如果不存在相同内容的记录（即校验码只是碰巧相同），则将记录追加在末尾
		recs.m[rec.Crc32] = append(recList, &rec)
	}
}

// 扫描一个目录及其子目录下的.yinao.txt文件，读取其中的加密医闹记录，并且调整这些记录的置信参数
func (recs *Records) AddEncRecordsInDir(dir string, confDelta float32, errLog io.Writer) {
	files, subdirs := getFilesAndSubDirs(dir, errLog)
	for _, f := range files {
		err := extractRecordsFromFile(f, func(recLines []string, off int64) error {
			rec := parseLines(recLines, errLog)
			if rec != nil {
				recs.Add(*rec, confDelta)
			}
			return nil
		})
		if err != nil {
			errLog.Write([]byte(err.Error()))
			errLog.Write([]byte("\n"))
			return
		}
	}
	for _, subdir := range subdirs { // 递归地扫描子目录
		recs.AddEncRecordsInDir(subdir, confDelta, errLog)
	}
}

// 将各条记录写入文件
func (recs *Records) WriteToFile(file io.Writer) (err error) {
	keys := make([]uint32, 0, len(recs.m))
	for k := range recs.m {
		keys = append(keys, k)
	}
	// 对键（crc32的值）进行排序
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, key := range keys {
		recList := recs.m[key]
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
	}
	return
}

// 将若干行的加密医闹记录，转换为一个Record
func parseLines(recLines []string, errLog io.Writer) *Record {
	if len(recLines) != 5 {
		s := fmt.Sprintf("记录的长度错误，必须正好有5行：%s\n", strings.Join(recLines, "\n"))
		errLog.Write([]byte(s))
		return nil
	}
	rec := &Record{}
	//第一行是患者基本信息的哈希
	bz, err := base64.StdEncoding.DecodeString(recLines[0])
	copy(rec.BaseInfoHash[:], bz)
	if err != nil {
		errLog.Write([]byte(fmt.Sprintf("base64编码错误：%s\n", recLines[0])))
		return nil
	}
	if len(bz) != sha256.Size {
		s := fmt.Sprintf("哈希值长度错误：%s\n", recLines[0])
		errLog.Write([]byte(s))
		return nil
	}
	//第二行是患者的身份证号的哈希
	bz, err = base64.StdEncoding.DecodeString(recLines[1])
	copy(rec.IDHash[:], bz)
	if err != nil {
		errLog.Write([]byte(fmt.Sprintf("base64编码错误：%s\n", recLines[1])))
		return nil
	}
	if len(bz) != sha256.Size {
		errLog.Write([]byte(fmt.Sprintf("哈希值长度错误：%s\n", recLines[1])))
		return nil
	}
	//第三行是置信指数，它是一个百分数，最大为100，最小为0，
	rec.Confidence, err = parseConfidence(recLines[2])
	if err != nil {
		errLog.Write([]byte(err.Error()))
		errLog.Write([]byte("\n"))
		return nil
	}
	//第四行是对于患者医闹记录的文本描述
	rec.Description = recLines[3]
	//第五行是前面第一、二、四行的CRC32校验码（Hex编码）
	crcBz, err := hex.DecodeString(recLines[4])
	if err != nil {
		errLog.Write([]byte(fmt.Sprintf("校验码编码格式错误：%s\n", recLines[4])))
		return nil
	}
	rec.Crc32 = binary.BigEndian.Uint32(crcBz)
	if !rec.VerifyChecksum() {
		s := fmt.Sprintf("校验码错误：%s\n", strings.Join(recLines, "\n"))
		errLog.Write([]byte(s))
		return nil
	}
	return rec
}

// 获得目录下面的文件列表和子目录列表
func getFilesAndSubDirs(dir string, errLog io.Writer) (files []string, subdirs []string) {
	items, err := ioutil.ReadDir(dir)
	if err != nil {
		errLog.Write([]byte(err.Error()))
		return
	}
	for _, item := range items {
		fullName := path.Join(dir, item.Name())
		if item.IsDir() {
			subdirs = append(subdirs, fullName)
		} else if strings.HasSuffix(fullName, EncFileSuffix) {
			files = append(files, fullName)
		}
	}
	return
}

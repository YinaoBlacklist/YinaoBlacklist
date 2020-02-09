package db

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Position struct {
	FileName string //此记录定义在哪个文件中
	Offset   int64  //此记录开始于文件的哪个位置
}

type PositionMap = map[[8]byte][]Position // 为了节省内存，索引只使用了最低的8个字节作为键

// 用于查询医闹记录的数据库
type DB struct {
	FileMap     map[string]*os.File // 保存若干已打开的文件
	BaseInfoMap PositionMap         // 从BaseInfoHash的低8个字节定位到记录在文件中的位置
	IDMap       PositionMap         // 从IDHash的低8个字节定位到记录在文件中的位置
}

func (db *DB) Close() {
	for _, file := range db.FileMap {
		file.Close()
	}
}

func appendPostion(m PositionMap, buf [8]byte, pos Position) {
	posList, ok := m[buf]
	if !ok {
		posList = []Position{}
	}
	m[buf] = append(posList, pos)
}

// 利用若干文件初始化数据库
func NewDBFromFiles(fnameList []string) (*DB, error) {
	var buf [8]byte
	shaNA := sha256.Sum256([]byte("NA"))
	db := &DB{
		FileMap:     make(map[string]*os.File),
		BaseInfoMap: make(PositionMap),
		IDMap:       make(PositionMap),
	}
	for _, fname := range fnameList {
		err := extractRecordsFromFile(fname, func(recLines []string, off int64) error {
			pos := Position{FileName: fname, Offset: off}
			var b bytes.Buffer
			rec := parseLines(recLines, &b)
			if b.Len() != 0 {
				return fmt.Errorf("读取文件%s时，遇到错误：%s\n", pos.FileName, b.String())
			}
			if rec != nil {
				copy(buf[:], rec.BaseInfoHash[:8])
				appendPostion(db.BaseInfoMap, buf, pos)
			}
			if rec != nil && !bytes.Equal(rec.IDHash[:], shaNA[:]) {
				copy(buf[:], rec.IDHash[:8])
				appendPostion(db.IDMap, buf, pos)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		db.FileMap[fname], _ = os.Open(fname)
	}
	return db, nil
}

type RecordInFile struct {
	Record
	FileName string
}

// 将记录转为6行纯文本
func (rec *RecordInFile) ToLines() []string {
	var lines [6]string
	lines[0] = "======= 来自文件：" + rec.FileName
	copy(lines[1:], rec.Record.ToLines())
	return lines[:]
}

// 给定文件中的一个位置，利用已经打开的文件，从这个位置读取一个医闹记录出来
func readRecord(fm map[string]*os.File, pos Position) (*RecordInFile, error) {
	file := fm[pos.FileName]
	_, err := file.Seek(pos.Offset, os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	recLines := make([]string, 0, 5)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		recLines = append(recLines, line)
		if len(recLines) == 5 {
			break
		}
	}
	var b bytes.Buffer
	rec := parseLines(recLines, &b)
	if b.Len() != 0 {
		return nil, fmt.Errorf("读取文件%s时，遇到错误：%s", pos.FileName, b.String())
	}
	return &RecordInFile{Record: *rec, FileName: pos.FileName}, nil
}

// 用于查询医闹记录的函数，m保存从哈希值低8位到位置的索引，fm保存若干打开的文件，hash为哈希值，
// filter是对获得的记录进行过滤的函数（返回true时才保留记录）
func query(m PositionMap, fm map[string]*os.File, hash [sha256.Size]byte, filter func(*RecordInFile) bool) ([]*RecordInFile, error) {
	res := make([]*RecordInFile, 0, 10)
	var buf [8]byte
	copy(buf[:], hash[:8])
	posList, ok := m[buf]
	if !ok {
		return nil, nil
	}
	for _, pos := range posList {
		rec, err := readRecord(fm, pos)
		if err != nil {
			return nil, err
		}
		rec.FileName = pos.FileName
		if filter(rec) {
			res = append(res, rec)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Confidence < res[j].Confidence
	})
	return res, nil
}

// 给定基本信息的哈希，查询医闹记录
func (db *DB) QueryBaseInfo(hash [sha256.Size]byte) ([]*RecordInFile, error) {
	return query(db.BaseInfoMap, db.FileMap, hash, func(rec *RecordInFile) bool {
		return bytes.Equal(rec.BaseInfoHash[:], hash[:]) //哈希的所有32个字节都必须相等
	})
}

// 给定身份证的哈希，查询医闹记录
func (db *DB) QueryID(hash [sha256.Size]byte) ([]*RecordInFile, error) {
	return query(db.IDMap, db.FileMap, hash, func(rec *RecordInFile) bool {
		return bytes.Equal(rec.IDHash[:], hash[:]) //哈希的所有32个字节都必须相等
	})
}

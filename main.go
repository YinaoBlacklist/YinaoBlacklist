package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"

	"github.com/YinaoBlacklist/YinaoBlacklist/db"
)

const (
	Count = 10
)

var mainwin *ui.Window

func setupUI() {
	mainwin = ui.NewWindow("医闹黑名单", 1124, 568, true)
	mainwin.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		mainwin.Destroy()
		return true
	})

	tab := ui.NewTab()
	mainwin.SetChild(tab)
	mainwin.SetMargined(true)

	tab.Append("将原始记录文件转为加密记录文件", makeConvertPage())
	tab.SetMargined(0, true)

	tab.Append("扫描并且合并加密记录文件", makeMergePage())
	tab.SetMargined(1, true)

	tab.Append("将加密记录载入内存以供查询", makeLoadPage())
	tab.SetMargined(2, true)

	tab.Append("使用内存中的加密记录进行查询", makeQueryPage())
	tab.SetMargined(3, true)

	mainwin.Show()
}

func makeConvertPage() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	vbox.Append(hbox, false)

	selBtn := ui.NewButton("选择原始记录文件")
	entry := ui.NewEntry()
	entry.SetReadOnly(true)
	selBtn.OnClicked(func(*ui.Button) {
		filename := ui.OpenFile(mainwin)
		entry.SetText(filename)
	})
	hbox.Append(selBtn, false)
	hbox.Append(entry, true)

	runBtn := ui.NewButton("转换为加密记录文件")
	runBtn.OnClicked(func(*ui.Button) {
		runConvert(entry.Text())
	})
	vbox.Append(runBtn, false)
	return vbox
}

func makeMergePage() ui.Control {
	vbox := ui.NewVerticalBox()
	dirEntryList := make([]*ui.Entry, Count)
	deltaEntryList := make([]*ui.Entry, Count)
	grid := ui.NewGrid()
	for i := 0; i < Count; i++ {
		s := fmt.Sprintf("目录%02d：", i+1)
		grid.Append(ui.NewLabel(s), 0, i, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
		dirEntryList[i] = ui.NewEntry()
		dirEntryList[i].SetReadOnly(false)
		grid.Append(dirEntryList[i], 1, i, 2, 1, true, ui.AlignFill, false, ui.AlignFill)
		grid.Append(ui.NewLabel("置信参数调整："), 3, i, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
		deltaEntryList[i] = ui.NewEntry()
		deltaEntryList[i].SetText("0")
		grid.Append(deltaEntryList[i], 4, i, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	}
	vbox.Append(grid, true)
	runBtn := ui.NewButton("合并为单一记录文件")
	runBtn.OnClicked(func(*ui.Button) {
		dirList := make([]string, Count)
		deltaList := make([]string, Count)
		for i, e := range dirEntryList {
			dirList[i] = e.Text()
		}
		for i, e := range deltaEntryList {
			deltaList[i] = e.Text()
		}
		runMerge(dirList, deltaList)
	})
	vbox.Append(runBtn, false)
	return vbox
}

func makeLoadPage() ui.Control {
	vbox := ui.NewVerticalBox()
	fileEntryList := make([]*ui.Entry, Count)
	grid := ui.NewGrid()
	for i := 0; i < Count; i++ {
		s := fmt.Sprintf("选择记录文件%02d：", i+1)
		selBtn := ui.NewButton(s)
		entry := ui.NewEntry()
		entry.SetReadOnly(true)
		selBtn.OnClicked(func(*ui.Button) {
			filename := ui.OpenFile(mainwin)
			entry.SetText(filename)
		})
		fileEntryList[i] = entry
		grid.Append(selBtn, 0, i, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
		grid.Append(entry, 1, i, 1, 1, true, ui.AlignFill, false, ui.AlignFill)
	}
	vbox.Append(grid, true)
	runBtn := ui.NewButton("载入上述所有记录文件到内存中")
	runBtn.OnClicked(func(*ui.Button) {
		fileList := make([]string, 0, Count)
		for _, e := range fileEntryList {
			if len(e.Text()) != 0 {
				fileList = append(fileList, e.Text())
			}
		}
		fmt.Printf("iiiiii %v\n", fileList)
		runLoad(fileList)
	})
	vbox.Append(runBtn, false)
	return vbox
}

func makeQueryPage() ui.Control {
	vbox := ui.NewVerticalBox()
	resultEntry := ui.NewMultilineEntry()
	resultEntry.SetReadOnly(true)

	vbox.Append(ui.NewLabel("输入基本信息（格式为“姓名，性别，出生年份”，注意中间要用中文逗号隔开）："), false)
	hbox := ui.NewHorizontalBox()
	baseInfoEntry := ui.NewEntry()
	hbox.Append(baseInfoEntry, true)
	baseInfoBtn := ui.NewButton("按基本信息进行查询")
	var idEntry *ui.Entry
	baseInfoBtn.OnClicked(func(*ui.Button) {
		idEntry.SetText("")
		runQueryWithBaseInfo(resultEntry, baseInfoEntry.Text())
	})
	hbox.Append(baseInfoBtn, false)
	hbox.SetPadded(true)
	vbox.Append(hbox, false)

	vbox.Append(ui.NewLabel("输入身份证号："), false)
	hbox = ui.NewHorizontalBox()
	idEntry = ui.NewEntry()
	hbox.Append(idEntry, true)
	idBtn := ui.NewButton("按身份证号进行查询")
	idBtn.OnClicked(func(*ui.Button) {
		baseInfoEntry.SetText("")
		runQueryWithID(resultEntry, idEntry.Text())
	})
	hbox.Append(idBtn, false)
	hbox.SetPadded(true)
	vbox.Append(hbox, false)

	vbox.Append(resultEntry, true)

	return vbox
}

func main() {
	ui.Main(setupUI)
}

// ==================================

// 检查某文件或者目录是否存在
func checkExist(path string, isDir bool) bool {
	t := "文件"
	if isDir {
		t = "目录"
	}
	if fileInfo, err := os.Stat(path); os.IsNotExist(err) || fileInfo.IsDir() != isDir {
		ui.MsgBoxError(mainwin, "不存在的"+t, "不存在的"+t+"：'"+path+"'")
		return false
	}
	return true
}

// 将原始记录文件转为加密记录文件
func runConvert(fname string) {
	if !checkExist(fname, false) {
		ui.MsgBoxError(mainwin, "错误！", "文件 "+fname+" 不存在！")
		return
	}
	if !strings.HasSuffix(fname, ".txt") {
		ui.MsgBoxError(mainwin, "非文本文件", "您选择的文件不是文本文件，无法进行处理。")
		return
	}
	recList, err := db.ExtractRecordsFromRawFile(fname)
	if err != nil {
		ui.MsgBoxError(mainwin, "错误！", err.Error())
		return
	}
	outFile := fname[:len(fname)-4] + db.EncFileSuffix
	out, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		ui.MsgBoxError(mainwin, "错误！", err.Error())
		return
	}
	defer out.Close()
	err = db.WriteRecordsToFile(recList, out)
	if err != nil {
		ui.MsgBoxError(mainwin, "错误！", err.Error())
		return
	}
	ui.MsgBox(mainwin, "转换成功", "转换成功，输出文件位于："+outFile)
}

// 扫描并且合并加密记录文件
func runMerge(dirList, deltaStrList []string) {
	for _, dir := range dirList {
		if len(dir) != 0 && !checkExist(dir, true) {
			ui.MsgBoxError(mainwin, "错误！", "目录 "+dir+" 不存在！")
			return
		}
	}
	deltaList := make([]float32, len(deltaStrList))
	for i, deltaStr := range deltaStrList {
		delta, err := strconv.ParseFloat(deltaStr, 32)
		if err != nil {
			ui.MsgBoxError(mainwin, "错误！", deltaStr+" 不是合法的数字！")
			return
		}
		deltaList[i] = float32(delta / 100.0)
	}

	ex, _ := os.Executable()
	logName := path.Join(filepath.Dir(ex), "log.txt")
	logfile, err := os.OpenFile(logName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		ui.MsgBoxError(mainwin, "错误！", err.Error())
		return
	}

	records := db.NewRecords()
	for i, dir := range dirList {
		if len(dir) == 0 {
			continue
		}
		records.AddEncRecordsInDir(dir, deltaList[i], logfile)
	}
	logfile.Close()

	fileInfo, err := os.Stat(logName)
	if err != nil {
		ui.MsgBoxError(mainwin, "错误！", err.Error())
		return
	}
	if fileInfo.Size() != 0 {
		ui.MsgBoxError(mainwin, "发现错误！", "转换时发现错误，请打开 "+logName+" 文件查看详情。")
		return
	}

	ui.MsgBox(mainwin, "请选择输出文件", "合并完毕！接下来请您选择一个输出文件用于保存合并后的记录。")
	outFile := ui.SaveFile(mainwin)
	if len(outFile) == 0 {
		ui.MsgBoxError(mainwin, "未选择文件", "您并未选择一个输出文件，转换后的结果不会被保存。")
		return
	}
	if !strings.HasSuffix(outFile, db.EncFileSuffix) {
		outFile = outFile + db.EncFileSuffix
	}
	out, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		ui.MsgBoxError(mainwin, "错误！", err.Error())
		return
	}
	defer out.Close()
	err = records.WriteToFile(out)
	if err != nil {
		ui.MsgBoxError(mainwin, "错误！", err.Error())
		return
	}
}

var YiNaoDB *db.DB //内存中保存的医闹记录

// 将记录载入内存(供查询用)
func runLoad(fileList []string) {
	for _, file := range fileList {
		if !checkExist(file, false) {
			return
		}
	}
	if YiNaoDB != nil {
		YiNaoDB.Close()
	}
	var err error
	YiNaoDB, err = db.NewDBFromFiles(fileList)
	if err != nil {
		ui.MsgBoxError(mainwin, "错误！", err.Error())
		return
	}
	ui.MsgBox(mainwin, "成功", "记录已成功载入内存")
}

// 按基本信息进行查询(使用内存中载入的记录)
func runQueryWithBaseInfo(resultEntry *ui.MultilineEntry, baseInfo string) {
	if YiNaoDB == nil {
		ui.MsgBoxError(mainwin, "错误！", "尚未载入任何数据")
		return
	}
	if err := db.CheckBaseInfo(baseInfo); err != nil {
		ui.MsgBoxError(mainwin, "错误！", err.Error())
		return
	}
	h := sha256.Sum256([]byte(baseInfo))
	recList, err := YiNaoDB.QueryBaseInfo(h)
	if err != nil {
		ui.MsgBoxError(mainwin, "错误！", err.Error())
		return
	}
	info2 := db.BaseInfoToAdjacentYears(baseInfo)
	var hash2 [2][sha256.Size]byte
	for i, info := range info2 {
		hash2[i] = sha256.Sum256([]byte(info))
		extraList, _ := YiNaoDB.QueryBaseInfo(hash2[i])
		recList = append(recList, extraList...)
	}
	writeResult(resultEntry, recList, func(in string) string {
		out := strings.ReplaceAll(in, base64.StdEncoding.EncodeToString(h[:]), baseInfo)
		out = strings.ReplaceAll(out, base64.StdEncoding.EncodeToString(hash2[0][:]), info2[0])
		return strings.ReplaceAll(out, base64.StdEncoding.EncodeToString(hash2[1][:]), info2[1])
	})
}

// 按身份证信息进行查询(使用内存中载入的记录)
func runQueryWithID(resultEntry *ui.MultilineEntry, id string) {
	if YiNaoDB == nil {
		ui.MsgBoxError(mainwin, "错误！", "尚未载入任何数据")
		return
	}
	if err := db.CheckID(id); err != nil {
		ui.MsgBoxError(mainwin, "错误！", err.Error())
		return
	}
	h := sha256.Sum256([]byte(id))
	recList, err := YiNaoDB.QueryID(h)
	if err != nil {
		ui.MsgBoxError(mainwin, "错误！", err.Error())
		return
	}
	writeResult(resultEntry, recList, func(in string) string {
		return strings.ReplaceAll(in, base64.StdEncoding.EncodeToString(h[:]), id)
	})
}

func writeResult(resultEntry *ui.MultilineEntry, recList []*db.RecordInFile, fn func(string) string) {
	if len(recList) == 0 {
		resultEntry.SetText("没有查询到记录")
	} else {
		resultEntry.SetText("")
	}
	for _, rec := range recList {
		for _, line := range rec.ToLines() {
			line = strings.ReplaceAll(line, "\\n", "\n")
			line = fn(line)
			resultEntry.Append(line + "\n")
		}
		resultEntry.Append("\n")
	}
}

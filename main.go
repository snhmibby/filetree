package main

//little program to display a file-system tree and basic info
//once a file-path is selected, print it on stdout
//doubles as a stress test for the filesystem readdir & stat functions :)

import (
	"fmt"
	"io/fs"

	//"log"
	"os"

	G "github.com/AllenDang/giu"
	I "github.com/AllenDang/imgui-go"
)

var (
	statCache       = make(map[string]fs.FileInfo)
	dirCache        = make(map[string][]fs.FileInfo)
	showHiddenFiles = false
	selectedFile    string
	currentDir      string
)

const (
	timeFmt    = "02 Jan 06 15:04"
	nodeFlags  = I.TreeNodeFlagsOpenOnArrow | I.TreeNodeFlagsOpenOnDoubleClick
	leafFlags  = I.TreeNodeFlagsLeaf
	tableFlags = I.TableFlags_ScrollX | I.TableFlags_ScrollY | I.TableFlags_Resizable | I.TableFlags_SizingStretchProp
)

func mkSize(sz_ int64) string {
	sizes := []string{"KB", "MB", "GB", "TB"}
	sz := float64(sz_)
	add := ""
	for _, n := range sizes {
		if sz < 1024 {
			break
		}
		sz = sz / 1024
		add = n
	}
	if add == "" {
		return fmt.Sprint(sz_)
	} else {
		return fmt.Sprintf("%.2f %s", sz, add)
	}
}

//statFile follows symbolic links
func statFile(path string) (fs.FileInfo, error) {
	st, ok := statCache[path]
	if !ok {
		var err error
		st, err = os.Stat(path)
		if err != nil {
			return nil, err
		}
		statCache[path] = st
	}
	return st, nil
}

//return a list of directory entries
func readDir(path string) ([]fs.FileInfo, error) {
	entry, ok := dirCache[path]
	if !ok {
		direntry, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		entry := make([]fs.FileInfo, len(direntry))
		for i, f := range direntry {
			childPath := path + "/" + f.Name()
			if path == "/" {
				childPath = "/" + f.Name()
			}

			if f.Type()&fs.ModeSymlink == 0 {
				entry[i], err = f.Info()
			} else {
				entry[i], err = statFile(childPath)
			}
			if err != nil {
				return nil, err
			}
		}
		dirCache[path] = entry
	}
	return entry, nil
}

func getDirInfo(path string) (int, fs.FileInfo, []fs.FileInfo, bool) {
	info, err := statFile(path)
	if err != nil {
		return 0, nil, nil, false
	}

	entries, err := readDir(path)
	if err != nil {
		return 0, nil, nil, false
	}

	if info.Name()[0] == '.' && !showHiddenFiles {
		return 0, nil, nil, false
	}

	flags := leafFlags
	for _, e := range entries {
		if e.IsDir() {
			flags = nodeFlags
			break
		}
	}

	if path == currentDir {
		flags |= I.TreeNodeFlagsSelected
	}

	return flags, info, entries, true
}

func dirTree(path string) {
	flags, info, entries, ok := getDirInfo(path)
	if !ok {
		return
	}

	I.PushStyleVarFloat(I.StyleVarIndentSpacing, 7)
	open := I.TreeNodeV(info.Name(), flags)
	if I.IsItemClicked(int(G.MouseButtonLeft)) {
		currentDir = path
	}
	if open {
		defer I.TreePop()
		for _, e := range entries {
			if e.IsDir() {
				name := path + "/" + e.Name()
				if path == "/" {
					name = name[1:]
				}
				dirTree(name)
			}
		}
	}
	I.PopStyleVar()
}

func isHidden(entry fs.FileInfo) bool {
	return entry.Name()[0] == '.'
}

func fileTable() {
	if I.BeginTable("FSTable", 3, tableFlags, I.ContentRegionAvail(), 0) {
		defer I.EndTable()
		I.TableSetupColumn("Path", 0, 10, 0)
		I.TableSetupColumn("Size", 0, 2, 0)
		I.TableSetupColumn("Time", 0, 4, 0)
		I.TableSetupScrollFreeze(1, 1)
		I.TableHeadersRow()

		entries, err := readDir(currentDir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() || isHidden(e) {
				continue
			}
			path := currentDir + "/" + e.Name()
			if currentDir == "/" {
				path = path[1:]
			}

			I.TableNextRow(0, 0)
			if path == selectedFile {
				color := I.GetColorU32(I.CurrentStyle().GetColor(I.StyleColorTextSelectedBg))
				I.TableSetBgColor(I.TableBgTarget_RowBg0, uint32(color), -1)
			}

			I.TableNextColumn()
			I.Text(e.Name())
			if I.IsItemClicked(int(G.MouseButtonLeft)) {
				selectedFile = path
			}
			I.TableNextColumn()
			I.Text(mkSize(e.Size()))
			I.TableNextColumn()
			I.Text(e.ModTime().Format(timeFmt))
		}
	}
}

func selectFile() {
	fmt.Println(selectedFile)
	os.Exit(0)
}

func cancel() {
	os.Exit(1)
}

func loop() {
	G.SingleWindow().Layout(
		G.Row(
			G.Button("Select").OnClick(selectFile),
			G.Button("Cancel").OnClick(cancel),
			G.InputText(&selectedFile),
			G.Checkbox("Show Hidden", &showHiddenFiles),
		),
		G.SplitLayout(G.DirectionHorizontal, true, 200,
			G.Custom(func() { dirTree("/") }),
			G.Custom(fileTable),
		),
	)
}

func main() {
	G.SetDefaultFont("DejavuSansMono.ttf", 12)
	w := G.NewMasterWindow("FileTree", 800, 600, 0)
	w.Run(loop)
}

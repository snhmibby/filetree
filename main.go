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
	showHiddenFiles = false
	selectedFile    string
	currentDir      string
)

const (
	timeFmt    = "02 Jan 06 15:04"
	leafFlags  = I.TreeNodeFlagsLeaf | I.TreeNodeFlagsNoTreePushOnOpen
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
	var err error
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return st, nil
}

//return a list of directory entries
func readDir(path string) ([]string, error) {
	entry, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	name := make([]string, len(entry))
	for i, f := range entry {
		var childPath string
		if path == "/" {
			childPath = "/" + f.Name()
		} else {
			childPath = path + "/" + f.Name()
		}
		name[i] = childPath
	}
	return name, nil
}

func fsWalk(path string) error {
	f, err := statFile(path)
	if err != nil {
		return err
	}
	if f.Name()[0] == '.' && !showHiddenFiles {
		return nil
	}

	if !f.IsDir() {
		flags := leafFlags
		if path == selectedFile {
			flags |= I.TreeNodeFlagsSelected
		}
		I.TableNextRow(0, 0)
		I.TableNextColumn()
		I.TreeNodeV(f.Name(), flags)
		if I.IsItemClicked(int(G.MouseButtonLeft)) {
			selectedFile = path
		}
		I.TableNextColumn()
		I.Text(mkSize(f.Size()))
		I.TableNextColumn()
		I.Text(f.ModTime().Format(timeFmt))
	} else {
		flags := 0
		if path == selectedFile {
			flags |= I.TreeNodeFlagsSelected
		}
		I.TableNextRow(0, 0)
		I.TableNextRow(0, 0)
		I.TableNextColumn()

		open := I.TreeNodeV(f.Name(), flags)
		if I.IsItemClicked(int(G.MouseButtonLeft)) {
			selectedFile = path
		}
		I.TableNextColumn()
		I.Text("--")
		I.TableNextColumn()
		I.Text(f.ModTime().Format(timeFmt))

		if open {
			defer I.TreePop()
			entries, err := readDir(path)
			if err != nil {
				//log.Println(err)
			}
			var regular []string
			for _, name := range entries {
				st, err := statFile(name)
				if err != nil {
					//log.Println(err)
					continue
				}
				if st.IsDir() {
					err = fsWalk(name)
					if err != nil {
						//log.Println(err)
						continue
					}
				} else {
					regular = append(regular, name)
				}
			}
			for _, c := range regular {
				err := fsWalk(c)
				if err != nil {
					//log.Println(err)
					continue
				}
			}

		}
	}
	return nil
}

func mkFsTree() {
	if I.BeginTable("FSTable", 3, tableFlags, I.ContentRegionAvail(), 0) {
		defer I.EndTable()
		I.TableSetupColumn("Path", 0, 70, 0)
		I.TableSetupColumn("Size", 0, 15, 0)
		I.TableSetupColumn("Time", 0, 15, 0)
		I.TableSetupScrollFreeze(1, 1)
		I.TableHeadersRow()
		fsWalk("/")
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
		G.Child().Layout(G.Custom(mkFsTree)),
	)
}

func main() {
	w := G.NewMasterWindow("FileTree", 800, 600, 0)
	w.Run(loop)
}

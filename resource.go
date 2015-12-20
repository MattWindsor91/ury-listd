package main

type File struct {
	Name string
	Data interface{}
}

func NewFile(name string, data interface{}) *File {
	return &File{
		Name: name,
		Data: data,
	}
}

type Folder struct {
	Name    string
	Files   map[string]*File
	Folders map[string]*Folder
}

func NewFolder(name string) *Folder {
	return &Folder{
		Name:    name,
		Files:   make(map[string]*File),
		Folders: make(map[string]*Folder),
	}
}

// TODO: something with foo/bar paths?
func (f *Folder) Mkdir(name string) *Folder {
	f.Folders[name] = NewFolder(name)
	return f
}

func (f *Folder) Echo(name string, data interface{}) {
	f.Files[name] = NewFile(name, data)
}

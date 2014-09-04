package main

type File struct {
	Path string `json:"path"`
	Size string `json:"size"`
	Dir  bool   `json:"dir"`
}

type ByPath []File

func (b ByPath) Len() int           { return len(b) }
func (b ByPath) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByPath) Less(i, j int) bool { return b[i].Path < b[j].Path }

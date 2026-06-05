// Package model
package lib

type FrontmatterMeta struct {
	Format string   `yaml:"format"`
	Date   string   `yaml:"date"`
	Title  string   `yaml:"title"`
	Tags   []string `yaml:"tags"`
}

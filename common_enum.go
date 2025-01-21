package qs

//go:generate go run github.com/dmji/go-stringer@latest -type=OptionSliceSeparator --trimprefix=@me -output common_enum_string.go -nametransform=lower -fromstringgenfn

type OptionSliceSeparator int8

const (
	OptionSliceSeparatorUnspecified OptionSliceSeparator = iota
	OptionSliceSeparatorNone
	OptionSliceSeparatorComma
	OptionSliceSeparatorSemicolon
	OptionSliceSeparatorSpace
)

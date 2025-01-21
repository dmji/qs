package qs

import "fmt"

type CommonTagOptions struct {
	SliceSeparator OptionSliceSeparator
}

func (o *CommonTagOptions) InitDefaults() {
	if o.SliceSeparator == OptionSliceSeparatorUnspecified {
		o.SliceSeparator = OptionSliceSeparatorNone
	}
}

func (o *CommonTagOptions) ApplyDefaults(d *CommonTagOptions) {
	if o.SliceSeparator == OptionSliceSeparatorUnspecified {
		o.SliceSeparator = d.SliceSeparator
	}
}

func (o *CommonTagOptions) ParseOption(option string) (bool, error) {
	bOk := false

	// OptionSliceSeparator
	if value, err := OptionSliceSeparatorFromString(option); err == nil {
		if o.SliceSeparator != OptionSliceSeparatorUnspecified {
			return false, fmt.Errorf(fmtOptionNotUniqueError, "OptionSliceSeparator", o.SliceSeparator, value)
		}
		o.SliceSeparator = value
		bOk = true
	}

	return bOk, nil
}

func NewUndefinedCommonTagOptions() *CommonTagOptions {
	return &CommonTagOptions{
		SliceSeparator: OptionSliceSeparatorUnspecified,
	}
}

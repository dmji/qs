package qs

import "fmt"

type UnmarshalTagOptions struct {
	// DefaultUnmarshalPresence is used for the unmarshaling of struct fields
	// that don't have an explicit UnmarshalPresence option set in their tags.
	Presence UnmarshalPresence

	SliceValues UnmarshalSliceValues

	SliceUnexpectedValue UnmarshalSliceUnexpectedValue
}

func (o *UnmarshalTagOptions) InitDefaults() {
	if o.Presence == UnmarshalPresenceUPUnspecified {
		o.Presence = UnmarshalPresenceOpt
	}
	if o.SliceValues == UnmarshalSliceValuesUPUnspecified {
		o.SliceValues = UnmarshalSliceValuesOverrideOld
	}
	if o.SliceUnexpectedValue == UnmarshalSliceUnexpectedValueUPUnspecified {
		o.SliceUnexpectedValue = UnmarshalSliceUnexpectedValueBreakWithError
	}
}

func (o *UnmarshalTagOptions) ApplyDefaults(d *UnmarshalTagOptions) {
	if o.Presence == UnmarshalPresenceUPUnspecified {
		o.Presence = d.Presence
	}
	if o.SliceValues == UnmarshalSliceValuesUPUnspecified {
		o.SliceValues = d.SliceValues
	}
	if o.SliceUnexpectedValue == UnmarshalSliceUnexpectedValueUPUnspecified {
		o.SliceUnexpectedValue = d.SliceUnexpectedValue
	}
}

func (o *UnmarshalTagOptions) ParseOption(option string) (bool, error) {
	bOk := false

	// UnmarshalPresence
	if value, err := UnmarshalPresenceFromString(option); err == nil {
		if o.Presence != UnmarshalPresenceUPUnspecified {
			return false, fmt.Errorf(fmtOptionNotUniqueError, "UnmarshalPresence", o.Presence, value)
		}
		o.Presence = value
		bOk = true
	}

	// UnmarshalSliceValues
	if value, err := UnmarshalSliceValuesFromString(option); err == nil {
		if o.SliceValues != UnmarshalSliceValuesUPUnspecified {
			return false, fmt.Errorf(fmtOptionNotUniqueError, "UnmarshalSliceValues", o.SliceValues, value)
		}
		o.SliceValues = value
		bOk = true
	}

	// UnmarshalSliceUnexpectedValue
	if value, err := UnmarshalSliceUnexpectedValueFromString(option); err == nil {
		if o.SliceUnexpectedValue != UnmarshalSliceUnexpectedValueUPUnspecified {
			return false, fmt.Errorf(fmtOptionNotUniqueError, "UnmarshalSliceUnexpectedValue", o.SliceUnexpectedValue, value)
		}
		o.SliceUnexpectedValue = value
		bOk = true
	}

	return bOk, nil
}

func NewUndefinedUnmarshalTagOptions() *UnmarshalTagOptions {
	return &UnmarshalTagOptions{
		Presence:             UnmarshalPresenceUPUnspecified,
		SliceValues:          UnmarshalSliceValuesUPUnspecified,
		SliceUnexpectedValue: UnmarshalSliceUnexpectedValueUPUnspecified,
	}
}

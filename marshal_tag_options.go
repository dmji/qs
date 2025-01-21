package qs

import "fmt"

type MarshalTagOptions struct {
	// DefaultMarshalPresence is used for the marshaling of struct fields that
	// don't have an explicit MarshalPresence option set in their tags.
	// This option is used for every item when you marshal a map[string]WhateverType
	// instead of a struct because map items can't have a tag to override this.
	Presence MarshalPresence
}

func (o *MarshalTagOptions) InitDefaults() {
	if o.Presence == MarshalPresenceMPUnspecified {
		o.Presence = MarshalPresenceKeepEmpty
	}
}

func (o *MarshalTagOptions) ApplyDefaults(d *MarshalTagOptions) {
	if o.Presence == MarshalPresenceMPUnspecified {
		o.Presence = d.Presence
	}
}

func (o *MarshalTagOptions) ParseOption(option string) (bool, error) {
	bOk := false

	// OptionPresence
	if value, err := MarshalPresenceFromString(option); err == nil {
		if o.Presence != MarshalPresenceMPUnspecified {
			return false, fmt.Errorf(fmtOptionNotUniqueError, "OptionPresence", o.Presence, value)
		}
		o.Presence = value
		bOk = true
	}

	return bOk, nil
}

func NewUndefinedMarshalTagOptions() *MarshalTagOptions {
	return &MarshalTagOptions{
		Presence: MarshalPresenceMPUnspecified,
	}
}

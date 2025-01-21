package qs

import (
	"reflect"
	"strings"
	"testing"
)

type defaultPresenceTestCase struct {
	tagStr    reflect.StructTag
	defaultMO MarshalTagOptions
	mo        MarshalTagOptions
	defaultUO UnmarshalTagOptions
	uo        UnmarshalTagOptions
}

func TestParseTag_DefaultPresence(t *testing.T) {
	testCases := []defaultPresenceTestCase{
		{
			`qs:"name"`,
			MarshalTagOptions{MarshalPresenceKeepEmpty},
			MarshalTagOptions{MarshalPresenceKeepEmpty},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
		},
		{
			`qs:"name"`,
			MarshalTagOptions{MarshalPresenceOmitEmpty},
			MarshalTagOptions{MarshalPresenceOmitEmpty},
			UnmarshalTagOptions{Presence: UnmarshalPresenceOpt},
			UnmarshalTagOptions{Presence: UnmarshalPresenceOpt},
		},
		{
			`qs:"name"`,
			MarshalTagOptions{MarshalPresenceOmitEmpty},
			MarshalTagOptions{MarshalPresenceOmitEmpty},
			UnmarshalTagOptions{Presence: UnmarshalPresenceReq},
			UnmarshalTagOptions{Presence: UnmarshalPresenceReq},
		},
		{
			`qs:"name,omitempty"`,
			MarshalTagOptions{MarshalPresenceKeepEmpty},
			MarshalTagOptions{MarshalPresenceOmitEmpty},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
		},
		{
			`qs:"name,keepempty"`,
			MarshalTagOptions{MarshalPresenceKeepEmpty},
			MarshalTagOptions{MarshalPresenceKeepEmpty},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
		},
		{
			`qs:"name,omitempty"`,
			MarshalTagOptions{MarshalPresenceOmitEmpty},
			MarshalTagOptions{MarshalPresenceOmitEmpty},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
		},
		{
			`qs:"name,keepempty"`,
			MarshalTagOptions{MarshalPresenceOmitEmpty},
			MarshalTagOptions{MarshalPresenceKeepEmpty},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
		},
		{
			`qs:"name,nil"`,
			MarshalTagOptions{MarshalPresenceKeepEmpty},
			MarshalTagOptions{MarshalPresenceKeepEmpty},
			UnmarshalTagOptions{Presence: UnmarshalPresenceOpt},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
		},
		{
			`qs:"name,opt"`,
			MarshalTagOptions{MarshalPresenceKeepEmpty},
			MarshalTagOptions{MarshalPresenceKeepEmpty},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
			UnmarshalTagOptions{Presence: UnmarshalPresenceOpt},
		},
		{
			`qs:"name,req"`,
			MarshalTagOptions{MarshalPresenceKeepEmpty},
			MarshalTagOptions{MarshalPresenceKeepEmpty},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
			UnmarshalTagOptions{Presence: UnmarshalPresenceReq},
		},
		{
			`qs:"name,keepempty,opt"`,
			MarshalTagOptions{MarshalPresenceOmitEmpty},
			MarshalTagOptions{MarshalPresenceKeepEmpty},
			UnmarshalTagOptions{Presence: UnmarshalPresenceNil},
			UnmarshalTagOptions{Presence: UnmarshalPresenceOpt},
		},
	}

	defaultCommon := NewUndefinedCommonTagOptions()
	defaultCommon.InitDefaults()

	for _, tc := range testCases {
		t.Run("",
			func(t *testing.T) {
				tc.defaultUO.InitDefaults()
				tc.uo.ApplyDefaults(&tc.defaultUO)

				tc.defaultMO.InitDefaults()
				tc.mo.ApplyDefaults(&tc.defaultMO)

				tag, err := parseFieldTag(tc.tagStr, &tc.defaultMO, &tc.defaultUO, defaultCommon)
				if err != nil {
					t.Errorf("unexpected error - tag: %q :: %v", tc.tagStr, err)
					return
				}
				if tag.Name != "name" {
					t.Errorf("tag.Name == %q, want %q", tag.Name, "name")
				}
				if tag.MarshalPresence != tc.mo.Presence {
					t.Errorf("tag=%q, DefaultMarshalPresence=%v, MarshalPresence=%v, want %v",
						tc.tagStr, tc.defaultMO.Presence, tag.MarshalPresence, tc.mo.Presence)
				}
				if tag.UnmarshalOpts.Presence != tc.uo.Presence {
					t.Errorf("tag=%q, DefaultUnmarshalPresence=%v, UnmarshalPresence=%v, want %v",
						tc.tagStr, tc.defaultUO, tag.UnmarshalOpts.Presence, tc.uo.Presence)
				}
			},
		)
	}
}

func TestParseTag_SurplusComma(t *testing.T) {
	tagStrList := []reflect.StructTag{
		`qs:","`,
		`qs:"-,"`,
		`qs:"name,"`,
		`qs:",opt,"`,
		`qs:"-,opt,"`,
		`qs:"name,opt,"`,
		`qs:",,opt"`,
		`qs:"-,,opt"`,
		`qs:"name,,opt"`,
	}

	defaultCommon := NewUndefinedCommonTagOptions()
	defaultCommon.InitDefaults()

	defaultUO := &UnmarshalTagOptions{Presence: UnmarshalPresenceOpt}
	defaultUO.InitDefaults()

	defaultMO := &MarshalTagOptions{Presence: MarshalPresenceKeepEmpty}
	defaultMO.InitDefaults()

	for _, tagStr := range tagStrList {
		_, err := parseFieldTag(tagStr, defaultMO, defaultUO, defaultCommon)
		if err == nil {
			t.Errorf("unexpected success - tag: %q", tagStr)
			continue
		}
		if !strings.Contains(err.Error(), "tag string contains a surplus comma") {
			t.Errorf("expected a different error :: %v", err)
		}
	}
}

func TestParseTag_IncompatibleOptions(t *testing.T) {
	tagStrList := []reflect.StructTag{
		`qs:",opt,req"`,
		`qs:",nil,opt,req"`,
		`qs:",nil,req,opt"`,
		`qs:",opt,req,nil"`,
		`qs:",opt,nil,req"`,
		`qs:",req,nil,opt"`,
		`qs:",req,opt,nil"`,
		`qs:",req,opt"`,
		`qs:",req,nil"`,
		`qs:",nil,req"`,
		`qs:",nil,opt"`,
		`qs:",opt,nil"`,
		`qs:",keepempty,omitempty"`,
		`qs:",omitempty,keepempty"`,
	}

	defaultCommon := NewUndefinedCommonTagOptions()
	defaultCommon.InitDefaults()

	defaultUO := &UnmarshalTagOptions{Presence: UnmarshalPresenceOpt}
	defaultUO.InitDefaults()

	defaultMO := &MarshalTagOptions{Presence: MarshalPresenceKeepEmpty}
	defaultMO.InitDefaults()

	for _, tagStr := range tagStrList {
		_, err := parseFieldTag(tagStr, defaultMO, defaultUO, defaultCommon)
		if err == nil {
			t.Errorf("unexpected success - tag: %q", tagStr)
			continue
		}
		if !strings.Contains(err.Error(), "option is allwed - you've specified at least two") {
			t.Errorf("expected a different error :: %v", err)
		}
	}
}

var snakeTestCases = map[string]string{
	"woof_woof":                     "woof_woof",
	"_woof_woof":                    "_woof_woof",
	"woof_woof_":                    "woof_woof_",
	"WOOF":                          "woof",
	"Woof":                          "woof",
	"woof":                          "woof",
	"woof0_woof1":                   "woof0_woof1",
	"_woof0_woof1_2":                "_woof0_woof1_2",
	"woof0_WOOF1_2":                 "woof0_woof1_2",
	"WOOF0":                         "woof0",
	"Woof1":                         "woof1",
	"woof2":                         "woof2",
	"woofWoof":                      "woof_woof",
	"woofWOOF":                      "woof_woof",
	"woof_WOOF":                     "woof_woof",
	"Woof_WOOF":                     "woof_woof",
	"WOOFWoofWoofWOOFWoofWoof":      "woof_woof_woof_woof_woof_woof",
	"WOOF_Woof_woof_WOOF_Woof_woof": "woof_woof_woof_woof_woof_woof",
	"Woof_W":                        "woof_w",
	"Woof_w":                        "woof_w",
	"WoofW":                         "woof_w",
	"Woof_W_":                       "woof_w_",
	"Woof_w_":                       "woof_w_",
	"WoofW_":                        "woof_w_",
	"WOOF_":                         "woof_",
	"W_Woof":                        "w_woof",
	"w_Woof":                        "w_woof",
	"WWoof":                         "w_woof",
	"_W_Woof":                       "_w_woof",
	"_w_Woof":                       "_w_woof",
	"_WWoof":                        "_w_woof",
	"_WOOF":                         "_woof",
	"_woof":                         "_woof",
}

func TestSnakeCase(t *testing.T) {
	for input, output := range snakeTestCases {
		if snakeCase(input) != output {
			t.Errorf("snakeCase(%q) -> %q, want %q", input, snakeCase(input), output)
		}
	}
}

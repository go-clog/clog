// Copyright 2018 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package clog

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_discord_Init(t *testing.T) {
	Convey("Init Discord logger", t, func() {
		Convey("Mismatched config object", func() {
			err := New(DISCORD, struct{}{})
			So(err, ShouldNotBeNil)
			_, ok := err.(ErrConfigObject)
			So(ok, ShouldBeTrue)
		})

		Convey("Valid config object", func() {
			So(New(DISCORD, DiscordConfig{
				URL: "https://discordapp.com",
			}), ShouldBeNil)

			Convey("Incorrect level", func() {
				err := New(SLACK, SlackConfig{
					Level: LEVEL(-1),
				})
				So(err, ShouldNotBeNil)
				_, ok := err.(ErrInvalidLevel)
				So(ok, ShouldBeTrue)
			})
		})
	})
}

func Test_buildDiscordPayload(t *testing.T) {
	Convey("Build Discord payload", t, func() {
		payload, err := buildDiscordPayload("clog", &Message{
			Level: INFO,
			Body:  "[ INFO] test message",
		})
		So(err, ShouldBeNil)

		obj := &discordPayload{}
		So(json.Unmarshal([]byte(payload), obj), ShouldBeNil)
		So(obj.Username, ShouldEqual, "clog")
		So(len(obj.Embeds), ShouldEqual, 1)
		So(obj.Embeds[0].Title, ShouldEqual, "Information")
		So(obj.Embeds[0].Description, ShouldEqual, "test message")
		So(obj.Embeds[0].Timestamp, ShouldNotBeEmpty)
		So(obj.Embeds[0].Color, ShouldEqual, 3843043)
	})
}

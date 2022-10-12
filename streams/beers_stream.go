package streams

import "github.com/theobitoproject/kankuro/pkg/protocol"

// BeersName is the unique name and path for beers API
const BeersName = "beers"

// Beer defines all attributes that represents it
type Beer struct {
	Id      int    `json:"id"`
	Uid     string `json:"uid"`
	Name    string `json:"name"`
	Brand   string `json:"brand"`
	Style   string `json:"style"`
	Hop     string `json:"hop"`
	Yeast   string `json:"yeast"`
	Malts   string `json:"malts"`
	Ibu     string `json:"ibu"`
	Alcohol string `json:"alcohol"`
	Blg     string `json:"blg"`
}

// GetBeersStream returns the stream for the beers API
func GetBeersStream() protocol.Stream {
	return protocol.Stream{
		Name: "beers",
		JSONSchema: protocol.Properties{
			Properties: map[protocol.PropertyName]protocol.PropertySpec{
				"id": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.Integer, protocol.Null},
					},
					Description: "id of the beer record",
				},
				"uid": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.String, protocol.Null},
					},
					Description: "unique id of the beer record",
				},
				"name": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.String, protocol.Null},
					},
					Description: "name of the beer",
				},
				"brand": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.String, protocol.Null},
					},
					Description: "brand company name that produces the beer",
				},
				"style": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.String, protocol.Null},
					},
					Description: "style of the beer",
					Examples:    []string{"Scottish And Irish Ale", "European Amber Lager"},
				},
				"hop": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.String, protocol.Null},
					},
					Description: "type of hop used for the beer",
					Examples:    []string{"Crystal", "Hallertau"},
				},
				"yeast": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.String, protocol.Null},
					},
					Description: "type of yeast used for the beer",
					Examples:    []string{"1469 - West Yorkshire Ale", "2042 - Danish Lager"},
				},
				"malts": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.String, protocol.Null},
					},
					Description: "type of malts used for the beer",
					Examples:    []string{"Rye malt", "Victory"},
				},
				"ibu": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.String, protocol.Null},
					},
					Description: "level of ibu of the beer (international bitterness units)",
					Examples:    []string{"22 IBU"},
				},
				"alcohol": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.String, protocol.Null},
					},
					Description: "level of alcohol of the beer",
					Examples:    []string{"8.5%"},
				},
				"blg": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.String, protocol.Null},
					},
					Description: "level of bailing scale (sugar) of the beer",
					Examples:    []string{"10.5°Blg"},
				},
			},
		},
		SupportedSyncModes: []protocol.SyncMode{
			protocol.SyncModeFullRefresh,
		},
		SourceDefinedCursor: false,
		Namespace:           "random_api",
	}
}

package streams

import "github.com/theobitoproject/kankuro/pkg/protocol"

// AppliancesName is the unique name and path for appliances API
const AppliancesName = "appliances"

// Appliance defines all attributes that represents it
type Appliance struct {
	Id        int    `json:"id"`
	Uid       string `json:"uid"`
	Brand     string `json:"brand"`
	Equipment string `json:"equipment"`
}

// GetAppliancesStream returns the stream for the appliances API
func GetAppliancesStream() protocol.Stream {
	return protocol.Stream{
		Name: AppliancesName,
		JSONSchema: protocol.Properties{
			Properties: map[protocol.PropertyName]protocol.PropertySpec{
				"id": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.Integer, protocol.Null},
					},
					Description: "id of the appliance record",
				},
				"uid": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.String, protocol.Null},
					},
					Description: "unique id of the appliance record",
				},
				"brand": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.String, protocol.Null},
					},
					Description: "brand company name that manufactures the appliance",
				},
				"equipment": {
					PropertyType: protocol.PropertyType{
						Type: []protocol.PropType{protocol.String, protocol.Null},
					},
					Description: "name of the appliance",
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

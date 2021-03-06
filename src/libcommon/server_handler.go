package libcommon

import (
	"app"
	"errors"
	json "github.com/json-iterator/go"
	"libcommon/bridgev2"
	"libservicev2"
	"util/logger"
)

var ErrNilFrame = errors.New("frame is null")

// validate connection
func ValidateConnectionHandler(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame) error {
	if frame == nil {
		return ErrNilFrame
	}

	var meta = &bridgev2.ConnectMeta{}
	e1 := json.Unmarshal(frame.FrameMeta, meta)
	if e1 != nil {
		return e1
	}
	response := &bridgev2.ConnectResponseMeta{
		UUID:        app.UUID,
		New4Tracker: false,
	}
	responseFrame := &bridgev2.Frame{}

	if !IsInstanceIdUnique(meta.UUID) {
		logger.Error("register failed: instance_id is not unique")
		responseFrame.SetStatus(bridgev2.StatusInstanceIdExist)
		return manager.Send(responseFrame)
	} else {
		HoldUUID(meta.UUID)
	}

	if meta.Secret == app.Secret {
		responseFrame.SetStatus(bridgev2.StatusSuccess)
		manager.UUID = meta.UUID
		manager.State = bridgev2.StateValidated
		exist, e2 := libservicev2.ExistsStorage(meta.UUID)
		if e2 != nil {
			responseFrame.SetStatus(bridgev2.StatusInternalErr)
		} else {
			if exist {
				response.New4Tracker = false
			} else {
				response.New4Tracker = true
			}
		}
		// only valid client uuid (means storage client) will log into db.
		if meta.UUID != "" && len(meta.UUID) == 30 {
			storage := &app.StorageDO{
				Uuid:       meta.UUID,
				Host:       "",
				Port:       0,
				Status:     app.StatusEnabled,
				TotalFiles: 0,
				Group:      "",
				InstanceId: "",
				HttpPort:   0,
				HttpEnable: false,
				StartTime:  0,
				Download:   0,
				Upload:     0,
				Disk:       0,
				ReadOnly:   false,
				Finish:     0,
				IOin:       0,
				IOout:      0,
			}
			e3 := libservicev2.SaveStorage("", *storage)
			if e3 != nil {
				responseFrame.SetStatus(bridgev2.StatusInternalErr)
			}
		}
		responseFrame.SetMeta(response)
	} else {
		responseFrame.SetStatus(bridgev2.StatusBadSecret)
	}
	return manager.Send(responseFrame)
}

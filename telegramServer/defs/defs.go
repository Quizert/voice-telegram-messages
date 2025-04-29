package defs

const (
	WaitingModelName       = "waiting_model_name"
	WaitingDeleteModelName = "waiting_delete_model_name"
	WaitingVoice           = "waiting_voice"
	FreeState              = "free_state"
	MaxModels              = 5
)

type ErrNoModel struct {
	error string
}

func (e ErrNoModel) Error() string {
	return e.error
}

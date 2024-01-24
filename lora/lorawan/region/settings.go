package region

type Settings interface {
	JoinRequestChannel() Channel
	JoinAcceptChannel() Channel
	UplinkChannel() Channel
}

type settings struct {
	joinRequestChannel Channel
	joinAcceptChannel  Channel
	uplinkChannel      Channel
}

func (r *settings) JoinRequestChannel() Channel {
	return r.joinRequestChannel
}

func (r *settings) JoinAcceptChannel() Channel {
	return r.joinAcceptChannel
}

func (r *settings) UplinkChannel() Channel {
	return r.uplinkChannel
}

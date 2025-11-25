package model

type Model struct {
	Meta      `json:",inline"`
	KrakenRef string `json:"kraken_ref"`
}

func (m *Model) DeepCopy() *Model {
	if m == nil {
		return nil
	}
	return &Model{
		Meta:      m.Meta.DeepCopy(),
		KrakenRef: m.KrakenRef,
	}
}

package models

type PerfilPaciente struct {
	Diabetes012          uint8   `json:"diabetes_012"`
	HighBP               uint8   `json:"high_bp"`
	HighChol             uint8   `json:"high_chol"`
	HexCholCheck         uint8   `json:"chol_check"`
	BMI                  uint8   `json:"bmi"`
	Smoker               uint8   `json:"smoker"`
	Stroke               uint8   `json:"stroke"`
	HeartDiseaseorAttack uint8   `json:"heart_disease_or_attack"`
	PhysActivity         uint8   `json:"phys_activity"`
	Fruits               uint8   `json:"fruits"`
	Veggies              uint8   `json:"veggies"`
	HvyAlcoholConsump    uint8   `json:"hvy_alcohol_consump"`
	AnyHealthcare        uint8   `json:"any_healthcare"`
	NoDocbcCost          uint8   `json:"no_doc_bc_cost"`
	GenHlth              uint8   `json:"gen_hlth"`
	MentHlth             float64 `json:"ment_hlth"`
	PhysHlth             float64 `json:"phys_hlth"`
	DiffWalk             uint8   `json:"diff_walk"`
	Sex                  uint8   `json:"sex"`
	Age                  uint8   `json:"age"`
	Education            uint8   `json:"education"`
	Income               uint8   `json:"income"`
}

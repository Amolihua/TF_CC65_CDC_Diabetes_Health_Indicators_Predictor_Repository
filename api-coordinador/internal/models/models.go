package models

type PerfilPaciente struct {
	Diabetes012          uint8   `json:"diabetes_012"`
	HighBP               uint8   `json:"high_bp"`
	HighChol             uint8   `json:"high_chol"`
	CholCheck            uint8   `json:"chol_check"`
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

// Numerico -> conteo, media, varianza, desviación estándar y mediana
type MetricasNumericas struct {
	Conteo   int     `json:"conteo"`
	Media    float64 `json:"media"`
	Varianza float64 `json:"varianza"`
	DesvStd  float64 `json:"desv_std"`
	Mediana  float64 `json:"mediana"`
}

// Reporte estadistico de las métricas descriptivas
type ReporteEstadistico struct {
	FreqDiabetes012 map[uint8]int     `json:"freq_diabetes_012"`
	FreqHighBP      map[uint8]int     `json:"freq_high_bp"`
	FreqHighChol    map[uint8]int     `json:"freq_high_chol"`
	FreqSex         map[uint8]int     `json:"freq_sex"`
	MentHlth        MetricasNumericas `json:"ment_hlth"`
	PhysHlth        MetricasNumericas `json:"phys_hlth"`
}

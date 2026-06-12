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

// MetricasNumericas -> conteo, media, varianza, desviación estándar y mediana
type MetricasNumericas struct {
	Conteo   int     `json:"conteo"`
	Media    float64 `json:"media"`
	Varianza float64 `json:"varianza"`
	DesvStd  float64 `json:"desv_std"`
	Mediana  float64 `json:"mediana"`
}

// ReporteEstadistico contiene los mapas unificados de las 22 variables descriptivas
type ReporteEstadistico struct {
	FreqDiabetes012   map[uint8]int     `json:"freq_diabetes_012"`
	FreqHighBP        map[uint8]int     `json:"freq_high_bp"`
	FreqHighChol      map[uint8]int     `json:"freq_high_chol"`
	FreqCholCheck     map[uint8]int     `json:"freq_chol_check"`
	FreqSmoker        map[uint8]int     `json:"freq_smoker"`
	FreqStroke        map[uint8]int     `json:"freq_stroke"`
	FreqHeartDisease  map[uint8]int     `json:"freq_heart_disease"`
	FreqPhysActivity  map[uint8]int     `json:"freq_phys_activity"`
	FreqFruits        map[uint8]int     `json:"freq_fruits"`
	FreqVeggies       map[uint8]int     `json:"freq_veggies"`
	FreqHvyAlcohol    map[uint8]int     `json:"freq_hvy_alcohol"`
	FreqAnyHealthcare map[uint8]int     `json:"freq_any_healthcare"`
	FreqNoDocbcCost   map[uint8]int     `json:"freq_no_doc_bc_cost"`
	FreqGenHlth       map[uint8]int     `json:"freq_gen_hlth"`
	FreqDiffWalk      map[uint8]int     `json:"freq_diff_walk"`
	FreqSex           map[uint8]int     `json:"freq_sex"`
	FreqAge           map[uint8]int     `json:"freq_age"`
	FreqEducation     map[uint8]int     `json:"freq_education"`
	FreqIncome        map[uint8]int     `json:"freq_income"`
	BMIDistribuido    MetricasNumericas `json:"bmi_distribuido"`
	MentHlth          MetricasNumericas `json:"ment_hlth"`
	PhysHlth          MetricasNumericas `json:"phys_hlth"`
}

//go:build ignore

package main

import (
	"fmt"
	"os"
	"time"

	"api-coordinador/internal/analisis"
	"api-coordinador/internal/limpieza"
	"api-coordinador/internal/loader"
)

func main() {
	inicio := time.Now()

	numWorkers := 12
	datasetPath := os.Getenv("DATASET_PATH")
	if datasetPath == "" {
		datasetPath = "../datos_raw/diabetes_1M_extended.csv"
	}

	jobs := make(chan []string, 10000)

	go loader.LeerCSVMasivo(datasetPath, jobs)

	canalLimpio := limpieza.IniciarWorkerPool(numWorkers, jobs)

	reporte := analisis.EjecutarAnalisisConcurrente(numWorkers, canalLimpio)

	fmt.Printf("Pipeline de Análisis Concurrente completado en: %s\n\n", time.Since(inicio))

	fmt.Println("==========================================================================")
	fmt.Println("===             REPORTE EPIDEMIOLÓGICO DESCRIPTIVO GLOBAL              ===")
	fmt.Println("==========================================================================")

	fmt.Printf(" [Variables Principales]\n")
	fmt.Printf("  • Freq Diabetes (0:Sano, 1:Pre, 2:Diab): %v\n", reporte.FreqDiabetes012)
	fmt.Printf("  • Freq Presión Alta (0:No, 1:Sí):        %v\n", reporte.FreqHighBP)
	fmt.Printf("  • Freq Colesterol Alto (0:No, 1:Sí):     %v\n", reporte.FreqHighChol)
	fmt.Printf("  • Freq Evaluación Colesterol (0:No, 1:Sí):%v\n", reporte.FreqCholCheck)
	fmt.Printf("  • Freq Fumador (0:No, 1:Sí):             %v\n", reporte.FreqSmoker)
	fmt.Printf("  • Freq Ataque Cerebral (0:No, 1:Sí):     %v\n", reporte.FreqStroke)
	fmt.Printf("  • Freq Enfermedad Cardiaca (0:No, 1:Sí): %v\n\n", reporte.FreqHeartDisease)

	fmt.Printf(" [Variables de Estilo de Vida y Socioeconómicas]\n")
	fmt.Printf("  • Freq Actividad Física (0:No, 1:Sí):    %v\n", reporte.FreqPhysActivity)
	fmt.Printf("  • Freq Consume Frutas (0:No, 1:Sí):      %v\n", reporte.FreqFruits)
	fmt.Printf("  • Freq Consume Verduras (0:No, 1:Sí):    %v\n", reporte.FreqVeggies)
	fmt.Printf("  • Freq Consumo Alcohol Pesado (0:No,1:Sí):%v\n", reporte.FreqHvyAlcohol)
	fmt.Printf("  • Freq Seguro Médico (0:No, 1:Sí):       %v\n", reporte.FreqAnyHealthcare)
	fmt.Printf("  • Freq No va al Doc por Costo (0:No,1:Sí):%v\n", reporte.FreqNoDocbcCost)
	fmt.Printf("  • Freq Dificultad para Caminar (0:No,1:Sí):%v\n", reporte.FreqDiffWalk)
	fmt.Printf("  • Freq Sexo (0:Mujer, 1:Varón):          %v\n", reporte.FreqSex)
	fmt.Printf("  • Distribución Escala Salud General (1-5):%v\n", reporte.FreqGenHlth)
	fmt.Printf("  • Distribución Grupos de Edad (1-13):    %v\n", reporte.FreqAge)
	fmt.Printf("  • Distribución Nivel Educativo (1-6):    %v\n", reporte.FreqEducation)
	fmt.Printf("  • Distribución Rangos de Ingresos (1-8): %v\n\n", reporte.FreqIncome)

	fmt.Printf(" [Métricas Continuas Completa (Análisis Estadístico Local)]\n")
	fmt.Printf("  • Índice Masa Corporal (BMI) -> Media: %.2f | Varianza: %.2f | Mediana: %.2f\n",
		reporte.BMIDistribuido.Media, reporte.BMIDistribuido.Varianza, reporte.BMIDistribuido.Mediana)
	fmt.Printf("  • Días Mala Salud Mental     -> Media: %.2f | Varianza: %.2f | Mediana: %.2f\n",
		reporte.MentHlth.Media, reporte.MentHlth.Varianza, reporte.MentHlth.Mediana)
	fmt.Printf("  • Días Mala Salud Física     -> Media: %.2f | Varianza: %.2f | Mediana: %.2f\n",
		reporte.PhysHlth.Media, reporte.PhysHlth.Varianza, reporte.PhysHlth.Mediana)
	fmt.Println("==========================================================================")
}
